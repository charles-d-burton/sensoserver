package requests

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"tempserver/workers"
	"time"

	"github.com/dchest/uniuri"
	petname "github.com/dustinkirkland/golang-petname"
	uuid "github.com/satori/go.uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
		errorMessage := "Device not found: " + message.Sensor.Device
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorMessage))
		log.Println(err)
		return
	}
	workers.AddJob(*message)
}

func decoder(r *http.Request) (*workers.WorkRequest, error) {
	defer r.Body.Close()
	var message workers.WorkRequest
	err := json.NewDecoder(r.Body).Decode(&message)
	if message.Sensor.Exists() {
		log.Println("Found the sensor")
		return &message, err
	} else {
		return &message, errors.New("Device Not Found")
	}
}

func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var device workers.Sensor
	//var topic workers.Topic
	err := json.NewDecoder(r.Body).Decode(&device)
	//Assign a unique id and a friendly name to connected device
	id := uuid.NewV4()
	device.Device = id.String()
	device.Name = petname.Generate(2, "-")
	if err != nil || device.Topic.TopicString == "" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("☄ HTTP status code returned!"))
		return
	}
	device.Register()
	data, err := json.Marshal(&device)
	if err == nil {
		log.Println("Device Registered: ")
		log.Println("Device ", device.Device)
		log.Println("Name ", device.Name)
		io.WriteString(w, string(data))
	} else {
		log.Println("An error occurred")
		log.Println(err)
	}

}
