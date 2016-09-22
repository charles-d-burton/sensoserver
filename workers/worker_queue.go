package workers

import "encoding/json"

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

type WorkRequest struct {
	Token       string `json:"token"`
	MessageType string `json:"messagetype"`
	Topic       string `json:"topic"`
	Data        json.RawMessage
}

type Register struct {
	Token string `json:"token"`
	Topic string `json:"topic"`
}

func AddJob(message WorkRequest) {
	//Add message to the work queue
	WorkQueue <- message
}
