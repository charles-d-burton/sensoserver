package workers

import (
	"encoding/json"
	"log"
	"net/http"
)

var (
	MaxWorker = 4
	MaxQueue  = 100
	//MaxWorker = os.Getenv("MAX_WORKERS")
	//MaxQueue  = os.Getenv("MAX_QUEUE")
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

type WorkRequest struct {
	Token   string `json:"token"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func AddJob(w http.ResponseWriter, r *http.Request) {
	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var message WorkRequest
	err := decoder.Decode(&message)
	if err != nil {
		log.Println(err)
	} else {
		WorkQueue <- message
	}
}
