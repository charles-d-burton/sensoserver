package workers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"tempserver/helpers"

	"github.com/NaySoftware/go-fcm"
	"github.com/boltdb/bolt"
)

var (
	boltDB  *bolt.DB
	boltDir = "./senso.db"
)

const bucketTopics = "topics"

func init() {
	db, err := bolt.Open(boltDir, 0600, nil)
	if err != nil {
		panic(err)
	}
	boltDB = db
	boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketTopics))
		if err != nil {
			log.Println(err)
		}
		return nil
	})
}

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

func AddJob(message WorkRequest) {
	//Add message to the work queue
	WorkQueue <- message
}

type WorkRequest struct {
	Token       string `json:"token"`
	MessageType string `json:"messagetype"`
	Topic       string `json:"topic"`
	Data        json.RawMessage
}

func (work WorkRequest) PublishToFirebase() error {
	fcmClient := fcm.NewFcmClient(key)

	//Use a buffer to concat strings, it's much faster
	buffer := bytes.NewBuffer(make([]byte, 0, 32))
	buffer.WriteString("/topics/")
	buffer.WriteString(work.Topic)
	topic := buffer.String()

	log.Println("Topic:", topic)
	fcmClient.NewFcmMsgTo(topic, work.Data)
	fcmClient.SetTimeToLive(0)
	status, err := fcmClient.Send()
	if err == nil {
		status.PrintResults()
	} else {
		status.PrintResults()
		log.Println(err)
	}
	return err
}

type Registration struct {
	Token string `json:"token"`
	Topic string `json:"topic"`
}

func (register Registration) Register() error {
	token := make([]string, 0)
	token = append(token, register.Token)
	tokens := Tokens{
		Tokens: token,
	}
	data, err := json.Marshal(tokens)
	err = boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		err := b.Put([]byte(register.Topic), data)
		return err
	})
	return err
}

func (register Registration) JoinTopic() error {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		data := b.Get([]byte(register.Topic))
		if data != nil {
			var tokens Tokens
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.AppendIfMissing(tokens.Tokens, register.Token)
				data, err := json.Marshal(tokens)
				log.Println("Tokens joined!", string(data))
				err = b.Put([]byte(register.Topic), data)
				return err
			}
		} else {
			return errors.New("Topic Not Found")
		}
		return nil
	})
	return err
}

func (register Registration) LeaveTopic() error {
	return nil
}

type RefreshToken struct {
	NewToken string `json:"newtoken"`
}

type Tokens struct {
	Tokens []string `json:"tokens"`
}

//Discover if the token is already in a topic
func findToken(token string) (bool, string) {
	found := false
	topic := ""
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		c := b.Cursor()
		for key, value := c.First(); key != nil; key, value = c.Next() {
			var tokenStruct Tokens
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
