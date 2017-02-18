package workers

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/Jeffail/gabs"
	"github.com/boltdb/bolt"
	"github.com/tidwall/gjson"
)

func GetToken(message string) {

	log.Println(message)
}

func GetLastReadings(id string) (string, error) {
	//json.NewEncoder(os.Stderr).Encode(boltDB.Stats())
	var user User
	var message = ""
	err := boltDB.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(id))
		if userBucket == nil {
			log.Println("User not found")
			return nil
		} else {
			log.Println("User found")
			data := userBucket.Get([]byte("userData"))
			log.Println(string(data))
			err1 := json.Unmarshal(data, &user)
			message, err1 = retrieveLastReadings(user.Token)
			log.Println(message)
			return err1
		}
	})
	log.Println("Returning from GetData")
	return message, err
}

func retrieveLastReadings(token string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("Your current sensor readings. ")
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(token))

		data := b.Bucket([]byte("lastReadings"))
		data.ForEach(func(k, v []byte) error {
			request := gjson.GetBytes(v, "sensor")
			reading := gjson.GetBytes(v, "data")
			log.Println("LAST READING: ", string(v))
			sentence, err := generateSentence(request, reading)
			if sentence != "" {
				buffer.WriteString(sentence)
			}
			return err
		})
		return nil
	})
	response, err := generateAlexaResponse(buffer.String())
	return response, err
}

func generateSentence(sensor, reading gjson.Result) (string, error) {
	var buffer bytes.Buffer
	log.Println("Generating Sentence")
	if reading.Get("type").String() == "temperature" {
		log.Println("Found type Temperature")
		temp := reading.Get("tempF").String()
		name := sensor.Get("name").String()
		if name == "" {
			name = sensor.Get("device").String()
		}
		buffer.WriteString(name)
		buffer.WriteString(" ")
		buffer.WriteString(temp)
		buffer.WriteString(" degrees.")
		return buffer.String(), nil
	} else {
		return "", nil
	}

}

func generateAlexaResponse(messageString string) (string, error) {
	response := gabs.New()
	response.Set("1.0", "version")
	response.Set("PlainText", "response", "outputSpeech", "type")
	response.Set(messageString, "response", "outputSpeech", "text")
	return response.String(), nil
}
