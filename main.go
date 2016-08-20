package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"tempserver/structs"

	"rsc.io/letsencrypt"
)

func main() {
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
