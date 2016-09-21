package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"tempserver/workers"

	"rsc.io/letsencrypt"

	"github.com/NaySoftware/go-fcm"
)

var nWorkers = runtime.NumCPU()

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//Start the work dispatcher
	//key, err := getAPIKey()
	key := os.Getenv("APIKEY")

	data := map[string]string{
		"msg": "test",
		"sum": "something",
	}

	c := fcm.NewFcmClient(key)
	c.NewFcmMsgTo("/topics/test", data)
	status, err := c.Send()

	if err == nil {
		status.PrintResults()
	} else {
		log.Println(err)
	}

	workers.StartDispatcher(nWorkers, strings.Trim(key, " "))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS!\n")
	})

	http.HandleFunc("/reading", workers.AddJob)

	var m letsencrypt.Manager
	if err := m.CacheFile("letsencrypt.cache"); err != nil {
		log.Fatal(err)
	}
	log.Fatal(m.Serve())
}
