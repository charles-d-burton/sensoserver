package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sensoserver/workers"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"os"

	"github.com/dchest/uniuri"
	uuid "github.com/satori/go.uuid"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://smoker-relay.us/callback",
		ClientID:     os.Getenv("CLIENTID"),
		ClientSecret: os.Getenv("CLIENTSECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	// Some random string, random for each request
	oauthStateString = "random"
)

/*
Handle main page requests
*/
func Index(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "GET" {
		log.Println("Method not GET")
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	log.Println("Handling Index Request: ")
	path := r.URL.Path
	if path == "" {
		path = "index.html"
	}
	if bs, err := Asset(path); err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		var reader = bytes.NewBuffer(bs)
		io.Copy(w, reader)
	}
	//fmt.Fprintf(w, assets.HtmlIndex)
}

/*
Google Login code
*/
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

/*
Google Callback code
*/

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	fmt.Fprintf(w, "Content: %s\n", contents)
}

/*
register a new user with Google Cloud messaging
*/
func Register(w http.ResponseWriter, r *http.Request) {
	registration, err := decodeRegistration(w, r)
	//log.Println(registration.TopicString)
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

//JoinTopic ... with a given topic join the token to the topic, this token will receive messages from FCM
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

//LeaveTopic ... Remove a given token from a topic
func LeaveTopic(w http.ResponseWriter, r *http.Request) {
	registration, err := decodeRegistration(w, r)
	if err == nil {
		err = registration.LeaveTopic()
	}
}

//RecoverTopic ... Will eventually return a way to recover tokens from a topic
func RecoverTopic(w http.ResponseWriter, r *http.Request) {

}

//Load a registration request into a struct
func decodeRegistration(w http.ResponseWriter, r *http.Request) (*workers.Registration, error) {
	defer r.Body.Close()
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, errors.New("Method not allowed")
	}
	var registration workers.Registration
	err := json.NewDecoder(r.Body).Decode(&registration)
	return &registration, err
}

//RefreshToken ... Replace an expired token with a new one
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var refresh workers.RefreshToken
	err := json.NewDecoder(r.Body).Decode(&refresh)
	if err == nil {
		err = refresh.Refresh()
	}
}

//Reading ... process sensor data input
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

//Helper to decode messages
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

//RegisterDevice ... Register a new sensor with a given topic
func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var device workers.Sensor
	//var topic workers.Topic
	err := json.NewDecoder(r.Body).Decode(&device)
	//Assign a unique id and a friendly name to connected device
	id := uuid.NewV4()
	device.Device = id.String()
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
