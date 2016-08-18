package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"rsc.io/letsencrypt"
)

type Token struct {
	Token string `json:"token"`
}

type Reading struct {
	Reading string `json:"reading"`
	Time    string `json:"time"`
	Topic   string `json:"topic"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS!\n")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var token Token
		err := decoder.Decode(&token)
		if err != nil {
			panic(err)
		}
		go handleRegistration(token)
	})

	http.HandleFunc("/reading", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var reading Reading
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

func handleRegistration(token Token) {
	log.Println(token.Token)
}

func handleReading(reading Reading) {
	log.Println(reading.Reading)
}
