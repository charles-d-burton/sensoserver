package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"tempserver/structs"
	"tempserver/workers"

	"rsc.io/letsencrypt"
)

var nWorkers = 4

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//Start the work dispatcher
	workers.StartDispatcher(nWorkers)
	for i := 0; i < 1000000; i++ {
		work := workers.WorkRequest{strconv.Itoa(i), "Message"}
		workers.AddJob(work)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS!\n")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var token structs.Token
		err := decoder.Decode(&token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go handleRegistration(token)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/reading", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		decoder := json.NewDecoder(r.Body)
		var reading structs.Reading
		err := decoder.Decode(&reading)
		if err != nil {
			panic(err)
		}
		go handleReading(reading)

	})
	var m letsencrypt.Manager
	if err := m.CacheFile("letsencrypt.cache"); err != nil {
		log.Fatal(err)
	}
	log.Fatal(m.Serve())
}

func handleRegistration(token structs.Token) {
	log.Println(token.Token)
}

func handleReading(reading structs.Reading) {
	log.Println(reading.Reading)
}

func payloadHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read the body into a string for json decoding
	/*var content = &PayloadCollection{}
	#err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Go through each payload and queue items individually to be posted to S3
	for _, payload := range content.Payloads {

		// let's create a job with the payload
		work := Job{Payload: payload}

		// Push the work onto the queue.
		JobQueue <- work
	}*/

	w.WriteHeader(http.StatusOK)
}
