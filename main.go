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
)

var nWorkers = runtime.NumCPU()

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//Start the work dispatcher
	key := os.Getenv("APIKEY")

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

func register(w http.ResponseWriter, r *http.Request) {

}
