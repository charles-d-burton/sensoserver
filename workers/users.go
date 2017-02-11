package workers

import (
	"encoding/json"
	"log"
	"sensoserver/helpers"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

type User struct {
	Email string `json:"email"`
	Token string `json:"token"`
	Id    string `json:"user_id"`
}

type FirebaseKeys struct {
	Keys []string `json:"firebase"`
}

type DeviceObject struct {
	Device string `json:"device"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

func GetUser(user_id, email, firebase string) (User, error) {
	var user User

	err := boltDB.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(user_id))
		if userBucket == nil {
			log.Println("User not found")
			user.Email = email
			user.Token = uuid.NewV4().String()
			user.Id = user_id

			//First create the user
			userData, err := json.Marshal(user)
			log.Println("Creating User: ", string(email))
			userBucket, err = tx.CreateBucketIfNotExists([]byte(user_id))
			userBucket.Put([]byte("userData"), userData)
			userBucket.Put([]byte("api-key"), []byte(user.Token))

			//Next create the place to store api data
			firebaseKeys := FirebaseKeys{[]string{firebase}}
			fireBaseData, err := json.Marshal(firebaseKeys)
			log.Println("FIREBASE JSON ARRAY: ", fireBaseData)
			apiBucket, err := tx.CreateBucketIfNotExists([]byte(user.Token))
			apiBucket.Put([]byte("firebase-keys"), fireBaseData)

			log.Println("User created", err)

			//TODO:  Put this through the worker queue
			//go registerAPIKey(user.Token, &firebaseKeys)
			return err
		} else {
			log.Println("User found")
			userJson := userBucket.Get([]byte("userData"))
			err1 := json.Unmarshal(userJson, &user)
			if err1 == nil {
				//TODO:  Run this through the worker pool
				addToken(user.Token, firebase)
			}
			return err1
		}
	})
	log.Println("Returning from GetUser")
	return user, err
}

type APIToken struct {
	Token string `json:"token"`
}

/*func registerAPIKey(token string, keys *FirebaseKeys) error {
	keyData, err := json.Marshal(keys)
	if err != nil {
		err = boltDB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(token))
			//v := b.Get([]byte(key))
			if b == nil {

				log.Println("API Key not registered, registering new API Key")
				err := b.Put([]byte(token), keyData)
				return err
			}
			return nil
		})
	}

	return err
}*/

func addToken(key, firebase string) error {
	log.Println("Running ADDTOKEN")
	err := boltDB.Update(func(tx *bolt.Tx) error {
		apiBucket := tx.Bucket([]byte(key))
		if apiBucket != nil {
			var keys FirebaseKeys
			apiKeys := apiBucket.Get([]byte("firebase-keys"))
			err := json.Unmarshal(apiKeys, &keys)
			if err != nil {
				keys.Keys = helpers.AppendStringIfMissing(keys.Keys, firebase)
				keyData, err := json.Marshal(keys)
				log.Println("Json of Keydata: \n", string(keyData))
				err = apiBucket.Put([]byte("firebase-keys"), keyData)
				return err
			}
		} else {
			log.Println("No keys registered for: ", key)
			apiBucket, err := tx.CreateBucketIfNotExists([]byte(key))
			var keys FirebaseKeys
			keys.Keys = helpers.AppendStringIfMissing(keys.Keys, firebase)
			keyData, err := json.Marshal(keys)
			log.Println("Added Key! ", string(keyData))
			err = apiBucket.Put([]byte("firebase-keys"), keyData)
			return err
		}
		return nil
	})
	//TODO: Add to the worker pool
	go replayLastReadings(key)
	return err
}

func replayLastReadings(token string) error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		apiBucket := tx.Bucket([]byte(token))

		data := apiBucket.Bucket([]byte("lastReadings"))
		if data != nil {
			data.ForEach(func(k, v []byte) error {
				var message WorkRequest
				message.Token = token
				err := json.Unmarshal(v, &message.Data)
				data, err := json.Marshal(message)
				log.Println("Retrieved value: ", string(data))
				AddJob(message)
				return err
			})
		}
		return nil
	})
	return err
}

func GetData(id string) (string, error) {
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
			message, err1 = retrieveLastReading(user.Token)
			log.Println(message)
			return err1
		}
	})
	log.Println("Returning from GetData")
	return message, err
}

func retrieveLastReading(token string) (string, error) {
	var sensors []DeviceObject
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(token))

		data := b.Bucket([]byte("lastReadings"))
		data.ForEach(func(k, v []byte) error {
			var device DeviceObject
			request := gjson.GetBytes(v, "sensor")
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
