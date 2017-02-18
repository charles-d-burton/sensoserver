package workers

import (
	"encoding/json"
	"log"

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
			message, err1 = retrieveSensors(user.Token)
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
			log.Println("LAST READINGS: ", request.String())
			err := json.Unmarshal([]byte(request.String()), &device)
			sensors = append(sensors, device)
			return err
		})
		return nil
	})
	message, err := json.Marshal(sensors)
	log.Println("SENSORS: ", string(message))
	return string(message), err
}
