package requests

import (
	"encoding/json"
	"errors"
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
	registration, err := decodeRegistration(w, r)
	if err == nil {
		registration.Topic.TopicString = uniuri.New()
		registration.Register()
		data, err := json.Marshal(&registration)
		if err == nil {
			log.Println("Registration Fulfilled: ", registration.Topic.TopicString)
			log.Println(string(data))
			io.WriteString(w, string(data))
		} else {
			log.Println(err)
		}
	}
}

func JoinTopic(w http.ResponseWriter, r *http.Request) {
	registration, err := decodeRegistration(w, r)
	if err == nil {
		err = registration.JoinTopic()
		data, err := json.Marshal(&registration)
		if err == nil {
			log.Println("Topic Joined")
			io.WriteString(w, string(data))
		} else {
			log.Println(err)
		}
	}
}

func LeaveTopic(w http.ResponseWriter, r *http.Request) {
	registration, err := decodeRegistration(w, r)
	if err == nil {
		err = registration.LeaveTopic()
	}
}

func RecoverTopic(w http.ResponseWriter, r *http.Request) {

}

func decodeRegistration(w http.ResponseWriter, r *http.Request) (*workers.Registration, error) {
	defer r.Body.Close()
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, errors.New("Method not allowed")
	}
	var registration workers.Registration
	err := json.NewDecoder(r.Body).Decode(&registration)
	return &registration, err
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var refresh workers.RefreshToken
	err := json.NewDecoder(r.Body).Decode(&refresh)
	if err == nil {
		err = refresh.Refresh()
	}
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

func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {

	}
}

func SetDeviceIdFriendlyName(w http.ResponseWriter, r *http.Request) {

}
