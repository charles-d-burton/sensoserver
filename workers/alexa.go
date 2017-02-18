package workers

import (
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
	var sensors []DeviceObject
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(token))

		data := b.Bucket([]byte("lastReadings"))
		data.ForEach(func(k, v []byte) error {
			var device DeviceObject
			request := gjson.GetBytes(v, "sensor")
			log.Println("LAST READINGS: ", string(v))
			err := json.Unmarshal([]byte(request.String()), &device)
			sensors = append(sensors, device)
			return err
		})
		return nil
	})
	message, err := json.Marshal(sensors)
	log.Println("SENSORS: ", string(message))
	response, err := generateAlexaResponse("")
	return response, err
}

func generateAlexaResponse(reading string) (string, error) {
	response := gabs.New()
	response.Set("1.0", "version")
	response.Set("PlainText", "response", "outputSpeech", "type")
	response.Set("Something", "response", "outputSpeech", "text")
	return response.String(), nil
}
