package requests

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"tempserver/workers"

	"github.com/dchest/uniuri"
)

/*
register a new user with Google Cloud messaging
*/
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	message, err := decoder(r)
	if err != nil {
		log.Println(err)
		return
	}

	registration := new(workers.Register)
	registration.Token = message.Token
	uniqueTopic := uniuri.New()
	registration.Topic = uniqueTopic
	data, err := json.Marshal(&registration)
	log.Println("Registration Fulfilled")
	log.Println(string(data))

	message.Data = data
	message.Topic = uniqueTopic
	io.WriteString(w, string(data))
	workers.AddJob(message)
}

func JoinTopic(w http.ResponseWriter, r *http.Request) {

}

func RefreshToken(w http.ResponseWriter, r *http.Request) {

}

func RecoverTopic(w http.ResponseWriter, r *http.Request) {

}

func Reading(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	message, err := decoder(r)
	if err != nil {
		log.Println(err)
		return
	}
	workers.AddJob(message)
}

func decoder(r *http.Request) (workers.WorkRequest, error) {
	defer r.Body.Close()
	var message workers.WorkRequest
	err := json.NewDecoder(r.Body).Decode(&message)
	return message, err
}
