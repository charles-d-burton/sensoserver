package workers

import (
	"bytes"
	"fmt"
	"log"

	fcm "github.com/NaySoftware/go-fcm"
)

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

var key string

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start(ApiKey string) {
	key = ApiKey
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				handleWork(&work)
			case <-w.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}

func handleWork(work *WorkRequest) {
	log.Println(work.MessageType)
	switch work.MessageType {
	case "register":
		registerClient(work)
	case "reading":
		publishToFirebase(work)
	}

}

//Send the reading data to Firebase Cloud Messaging
func publishToFirebase(work *WorkRequest) {
	fcmClient := fcm.NewFcmClient(key)

	//Use a buffer to concat strings, it's much faster
	buffer := bytes.NewBuffer(make([]byte, 0, 32))
	buffer.WriteString("/topics/")
	buffer.WriteString(work.Topic)
	topic := buffer.String()

	log.Println("Topic:", topic)
	fcmClient.NewFcmMsgTo(topic, work.Data)
	status, err := fcmClient.Send()
	if err == nil {
		status.PrintResults()
	} else {
		status.PrintResults()
		log.Println(err)
	}
}

//Register a new client and topic
func registerClient(work *WorkRequest) {

}

//Join a new Device to a topic
func joinTopic(work *WorkRequest) {

}

//Invalidate and Refresh a Token
func refreshToken(work *WorkRequest) {

}
