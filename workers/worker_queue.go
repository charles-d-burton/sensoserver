package workers

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/NaySoftware/go-fcm"
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

func AddJob(message WorkRequest) {
	//Add message to the work queue
	WorkQueue <- message
}

type Topic struct {
	TopicString string `json:"topic"`
}

type WorkRequest struct {
	Sensor `json:"sensor"`
	Data   json.RawMessage `json:"data"`
}

type Payload struct {
	Device *string          `json:"device"`
	Name   *string          `json:"name"`
	Data   *json.RawMessage `json:"payload"`
}

func (work *WorkRequest) PublishToFirebase() error {
	fcmClient := fcm.NewFcmClient(key)
	//log.Println("Data: ", string(work.Data))
	//Use a buffer to concat strings, it's much faster
	buffer := bytes.NewBuffer(make([]byte, 0, 32))
	buffer.WriteString("/topics/")
	buffer.WriteString(work.Sensor.Topic.TopicString)
	topic := buffer.String()

	payload := work.transformToPayload()
	//data, err := json.Marshal(payload)
	//log.Println("Payload: ", string(data))
	fcmClient.NewFcmMsgTo(topic, payload)
	fcmClient.SetTimeToLive(0)
	status, err := fcmClient.Send()
	if err != nil {
		status.PrintResults()
		log.Println(err)
		//status.PrintResults()
	}
	return err
}

func (work *WorkRequest) transformToPayload() *Payload {
	payload := Payload{Device: &work.Sensor.Device, Name: &work.Sensor.Name, Data: &work.Data}
	return &payload
}
