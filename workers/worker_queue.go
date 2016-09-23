package workers

import (
	"bytes"
	"encoding/json"
	"log"

	fcm "github.com/NaySoftware/go-fcm"
	"github.com/boltdb/bolt"
)

var (
	boltDB  *bolt.DB
	boltDir = "./senso.db"
)

func init() {
	db, err := bolt.Open(boltDir, 0600, nil)
	if err != nil {
		panic(err)
	}
	boltDB = db
	boltDB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("topics"))
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

func (work WorkRequest) RegisterClient() error {
	boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("topics"))
		return nil
	})
	return nil
}

func (work WorkRequest) RefreshToken() error {
	return nil
}

func (work WorkRequest) JoinTopic() error {
	return nil
}

func (work WorkRequest) LeaveTopic() error {
	return nil
}

type Register struct {
	Token string `json:"token"`
	Topic string `json:"topic"`
}

type RefreshToken struct {
	NewToken string `json:"newtoken"`
}

func findKey(token string) (bool, error) {
	return false, nil
}
