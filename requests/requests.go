package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sensoserver/workers"

	"google.golang.org/api/oauth2/v2"
)

var (
	httpClient = &http.Client{}
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
	path := "index.html"
	if r.URL.Path != "/" {
		path = r.URL.Path[1:len(r.URL.Path)]
	}
	w.Header().Add("Content-Type", getContentType(path))
	log.Println(path)
	if bs, err := Asset(path); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		var reader = bytes.NewBuffer(bs)
		io.Copy(w, reader)
	}
}

/*
Google Callback code
*/
type Token struct {
	Token string `json:token`
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	/*requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))*/
	var token Token
	err := json.NewDecoder(r.Body).Decode(&token)
	log.Println("TOKEN RECEIVED: ", token.Token)
	//token := r.FormValue("idtoken")
	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(token.Token)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		log.Println(err)
	}
	user, err := workers.GetUser(tokenInfo.UserId, tokenInfo.Email)
	log.Println("Error:", err)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	log.Println(user.Email)
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
		//errorMessage := "Device not found: " + message.Sensor.Device
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]byte(errorMessage))
		log.Println(err)
		return
	}
	log.Println("adding message to queue")
	workers.AddJob(*message)
}

//Helper to decode messages
func decoder(r *http.Request) (*workers.WorkRequest, error) {
	defer r.Body.Close()
	var message workers.WorkRequest
	err := json.NewDecoder(r.Body).Decode(&message)
	data, err := json.Marshal(message)
	log.Println("Decoded message:", string(data))
	return &message, err
}
