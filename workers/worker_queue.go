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
	Topic
	Token       string `json:"token"`
	MessageType string `json:"messagetype"`
	Data        json.RawMessage
}

func (work WorkRequest) PublishToFirebase() error {
	fcmClient := fcm.NewFcmClient(key)

	//Use a buffer to concat strings, it's much faster
	buffer := bytes.NewBuffer(make([]byte, 0, 32))
	buffer.WriteString("/topics/")
	buffer.WriteString(work.Topic.TopicString)
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
