package workers

import (
	"encoding/json"
	"log"

	"github.com/NaySoftware/go-fcm"
	nsq "github.com/bitly/go-nsq"
	"github.com/boltdb/bolt"
	"github.com/tidwall/gjson"
)

var (
	pub         *nsq.Producer
	useNSQ      = false
	useFirebase = false
	nsqPort     string
	nsqHost     string
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

func AddJob(message WorkRequest) {
	//Add message to the work queue
	WorkQueue <- message
}

func SetQueueType(queue string) {
	if queue == "nsq" {
		useNSQ = true
	} else if queue == "firebase" {
		useFirebase = true
	} else if queue == "both" {
		useNSQ = true
		useFirebase = true
	}
}

type Topic struct {
	TopicString string `json:"topic,omitempty"`
}

type WorkRequest struct {
	//Sensor `json:"sensor"`
	Token string           `json:"token"`
	Data  *json.RawMessage `json:"data"`
}

func (work *WorkRequest) PublishToFirebase() error {
	log.Println("Received firebase publish request")
	fireBaseKeys := work.getKeys()
	if fireBaseKeys != nil && len(fireBaseKeys) > 0 {
		log.Println("Publishing to Firebase")
		fcmClient := fcm.NewFcmClient(key)

		//log.Println("API_KEY: ", key)
		//fmt.Printf("%v", fireBaseKeys)
		//log.Println("Topic: ", topic)
		//log.Println("\nPayload: ", data)

		//fcmClient.NewFcmMsgTo(topic, string(data))
		fcmClient.NewFcmRegIdsMsg(fireBaseKeys, work.Data)
		fcmClient.SetTimeToLive(0)
		status, err := fcmClient.Send()
		if err != nil {
			status.PrintResults()
			log.Println(err)
			//status.PrintResults()
		}
		status.PrintResults()
		work.recordLastEvent()
		return err
	}
	return nil
}

func (work *WorkRequest) getKeys() []string {
	var fireBaseKeys FirebaseKeys
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(work.Token))

		data := b.Get([]byte("firebase-keys"))
		err := json.Unmarshal(data, &fireBaseKeys)
		if err == nil && len(fireBaseKeys.Keys) > 0 {
			log.Println("Valid API Key Accepted")
			return nil
		} else {
			log.Println("No valid keys")
			return nil
		}
	})
	return fireBaseKeys.Keys
}

func (work *WorkRequest) recordLastEvent() error {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		apiBucket := tx.Bucket([]byte(work.Token))
		userData, err := apiBucket.CreateBucketIfNotExists([]byte("lastReadings"))
		j, err := json.Marshal(&work.Data)
		device := gjson.GetBytes(j, "sensor.device")
		log.Println("Device: ", device.String())
		log.Println("Data: ", string(j))
		err = userData.Put([]byte(device.String()), j)
		return err
	})
	return err
}
