package workers

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/NaySoftware/go-fcm"
	nsq "github.com/bitly/go-nsq"
	"github.com/boltdb/bolt"
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

func SetupNSQ(host, port string) {
	nsqHost = host
	nsqPort = port
	config := nsq.NewConfig()
	pub, _ = nsq.NewProducer(host+":"+port, config)
}

type Topic struct {
	TopicString string `json:"topic,omitempty"`
}

type WorkRequest struct {
	//Sensor `json:"sensor"`
	Token string          `json:"token"`
	Data  json.RawMessage `json:"data"`
}

type Payload struct {
	Device *string          `json:"device"`
	Name   *string          `json:"name"`
	Data   *json.RawMessage `json:"payload"`
}

func (work *WorkRequest) PublishToFirebase() error {
	if useFirebase && work.verifyAPIKey() {
		fcmClient := fcm.NewFcmClient(key)
		//log.Println("Data: ", string(work.Data))
		//Use a buffer to concat strings, it's much faster
		buffer := bytes.NewBuffer(make([]byte, 0, 32))
		buffer.WriteString("/topics/")
		buffer.WriteString(work.Token)
		topic := buffer.String()

		//payload := work.transformToPayload()
		//data, err := json.Marshal(payload)
		//log.Println("Payload: ", string(data))
		fcmClient.NewFcmMsgTo(topic, work.Data)
		fcmClient.SetTimeToLive(0)
		status, err := fcmClient.Send()
		if err != nil {
			status.PrintResults()
			log.Println(err)
			//status.PrintResults()
		}
		return err
	}
	return nil
}

func (work *WorkRequest) PublishToNSQ() error {
	if useNSQ {
		//err := pub.Publish(work.Topic.TopicString, []byte(work.Data))
		//return err
	}
	return nil
}

func (work *WorkRequest) transformToPayload() *Payload {
	//payload := Payload{Device: &work.Sensor.Device, Name: &work.Sensor.Name, Data: &work.Data}
	//return &payload
	return nil
}

func (work *WorkRequest) verifyAPIKey() bool {
	valid := false
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(apiBucket))
		data := b.Get([]byte(work.Token))
		if string(data) == "valid" {
			valid = true
		}
		return nil
	})
	return valid
}
