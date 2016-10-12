package workers

import (
	"encoding/json"
	"errors"
	"log"
	"tempserver/helpers"

	"github.com/boltdb/bolt"
)

var (
	boltDB    *bolt.DB
	boltDir   = "./senso.db"
	words     = 2
	separator = "-"
)

const (
	topicsBucket = "topics"
)

func init() {
	db, err := bolt.Open(boltDir, 0600, nil)
	if err != nil {
		panic(err)
	}
	boltDB = db
	boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(topicsBucket))
		if err != nil {
			log.Println(err)
		}
		return nil
	})

}

type Registration struct {
	Topic
	Token string `json:"token"`
}

type Devices struct {
	Tokens  []string `json:"tokens"`
	Sensors []string `json:"sensors"`
}

//Register ... Register a new endpoint device
func (register *Registration) Register() error {
	token := make([]string, 0)
	token = append(token, register.Token)
	tokens := Devices{
		Tokens: token,
	}

	data, err := json.Marshal(tokens)
	err = boltDB.Update(func(tx *bolt.Tx) error {
		topicBucket := tx.Bucket([]byte(topicsBucket))
		err := topicBucket.Put([]byte(register.Topic.TopicString), data)
		return err
	})
	return err
}

//JoinTopic ... Join a new endpoint device to an already created topic
func (register *Registration) JoinTopic() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topicsBucket))
		data := b.Get([]byte(register.Topic.TopicString))
		if data != nil {
			var tokens Devices
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.AppendStringIfMissing(tokens.Tokens, register.Token)
				data, err := json.Marshal(tokens)
				log.Println("Tokens joined!", string(data))
				//Update in the background
				go updateBolt(register.Topic.TopicString, data, topicsBucket)
				return err
			}
		} else {
			return errors.New("Topic Not Found")
		}
		return nil
	})
	return err
}

//LeaveTopic ... Remove a device token from a topic
func (register *Registration) LeaveTopic() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topicsBucket))
		data := b.Get([]byte(register.Topic.TopicString))
		if data != nil {
			var tokens Devices
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.RemoveStringByValue(tokens.Tokens, register.Token)
				log.Println("Token removed: ", register.Token, tokens.Tokens)
				go updateBolt(register.Topic.TopicString, data, topicsBucket)
				return err
			}

		}
		return nil
	})
	return err
}

//RefreshToken ... Describes a struct to update a device token with a refreshed token
type RefreshToken struct {
	NewToken string `json:"newtoken"`
	OldToken string `json:"oldtoken"`
	Topic
}

//Refresh ... Logic to replace one token with another
func (refreshToken *RefreshToken) Refresh() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topicsBucket))
		data := b.Get([]byte(refreshToken.Topic.TopicString))
		if data != nil {
			var tokens Devices
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.ReplaceStringByValue(tokens.Tokens, refreshToken.NewToken, refreshToken.OldToken)
				data, err := json.Marshal(tokens)
				go updateBolt(refreshToken.Topic.TopicString, data, topicsBucket)
				return err
			}
		}
		return nil
	})
	return err
}

type Sensor struct {
	Topic
	Device string `json:"device"`
	Name   string `json:"name,omitempty"`
}

/*type Sensors struct {
	Topic
	Sensors map[string]string `json:"sensors"`
}*/

func (sensor *Sensor) Register() error {

	log.Println("Topic: ", sensor.Topic.TopicString)
	log.Println("Device: ", sensor.Device)
	log.Println("Name: ", sensor.Name)
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topicsBucket))
		data := b.Get([]byte(sensor.Topic.TopicString))
		if data != nil && len(data) > 0 {
			var devices Devices
			err := json.Unmarshal(data, &devices)
			if err == nil && (len(devices.Sensors) == 0 || devices.Sensors == nil) {
				devices.Sensors = make([]string, 0)
			}
			devices.Sensors = append(devices.Sensors, sensor.Device)
			data, err = json.Marshal(devices)
			if err == nil {
				log.Println(string(data))
				go updateBolt(sensor.Topic.TopicString, data, topicsBucket)
			}
		}
		return nil
	})
	return err
}

//Change the the friendly name of a sensor
/*func (sensor *Sensor) Update() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		var sensors Sensors
		b := tx.Bucket([]byte(deviceBucket))
		data := b.Get([]byte(sensor.Topic.TopicString))
		err := json.Unmarshal(data, &sensors)
		if err == nil {
			sensors.Sensors[sensor.Device] = sensor.Name
			data, err = json.Marshal(sensors.Sensors)
			if err == nil {
				log.Println(string(data))
				go updateBolt(sensor.Topic.TopicString, data, deviceBucket)
			}
		}

		return err
	})
	return err
}*/

func (sensor *Sensor) Exists() bool {
	exists := false
	boltDB.View(func(tx *bolt.Tx) error {
		var devices Devices
		b := tx.Bucket([]byte(topicsBucket))
		data := b.Get([]byte(sensor.Topic.TopicString))
		err := json.Unmarshal(data, &devices)
		if err == nil && (devices.Sensors != nil || len(devices.Sensors) > 0) {
			for _, device := range devices.Sensors {
				if sensor.Device == device {
					exists = true
					break
				}
			}
		}
		return err
	})
	return exists
}

//Short and sweet function to update bolt
func updateBolt(topic string, message []byte, bucket string) {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(topic), message)
		return err
	})
	log.Println(err)
}

//Discover if the token is already in a topic
func findToken(token string) (bool, string) {
	found := false
	topic := ""
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topicsBucket))
		c := b.Cursor()
		for key, value := c.First(); key != nil; key, value = c.Next() {
			var tokenStruct Devices
			err := json.Unmarshal(value, &tokenStruct)
			if err != nil {
				for _, tVal := range tokenStruct.Tokens {
					if tVal == token {
						found = true
						topic = string(key)
					}
				}
			}
		}
		return nil
	})
	return found, topic
}
