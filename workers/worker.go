package workers

import (
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

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start(ApiKey string) {
	fcmClient := fcm.NewFcmClient(ApiKey)
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				//arr := []byte(work.Message)

				log.Println(work.Message)
				data := map[string]string{
					"msg": work.Message,
				}
				topic := "/topics/" + work.Topic
				log.Println("Topic:", topic)
				fcmClient.NewFcmMsgTo(topic, data)
				status, err := fcmClient.Send()
				if err == nil {
					status.PrintResults()
				} else {
					status.PrintResults()
					log.Println(err)
				}
				// Receive a work request.
				//log.Println("Job Received")
				//fmt.Printf("worker%d: Hello, %s!\n", w.ID, work.Name)

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
