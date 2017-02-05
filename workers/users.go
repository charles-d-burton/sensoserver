package workers

import (
	"encoding/json"
	"log"
	"sensoserver/helpers"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
)

type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

type FirebaseKeys struct {
	Keys []string `json:"firebase"`
}

func GetUser(user_id, email, firebase string) (User, error) {
	var user User

	//json.NewEncoder(os.Stderr).Encode(boltDB.Stats())
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		v := b.Get([]byte(user_id))
		if v == nil {
			log.Println("User not found")
			user.Email = email
			user.UserID = user_id
			user.Token = uuid.NewV4().String()
			data, err := json.Marshal(user)
			log.Println("Creating User: ", string(data))
			err = b.Put([]byte(user.UserID), data)
			log.Println("User created", err)

			firebaseKeys := FirebaseKeys{[]string{firebase}}
			go registerAPIKey(user.Token, &firebaseKeys)
			return err
		} else {
			log.Println("User found")
			err1 := json.Unmarshal(v, &user)
			if err1 == nil {
				go addToken(user.Token, firebase)
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

func registerAPIKey(token string, keys *FirebaseKeys) error {
	keyData, err := json.Marshal(keys)
	if err != nil {
		err = boltDB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(apiBucket))
			v := b.Get([]byte(key))
			if v == nil {

				log.Println("API Key not registered, registering new API Key")
				err := b.Put([]byte(token), keyData)
				return err
			}
			return nil
		})
	}

	return err
}

func addToken(key, firebase string) error {
	log.Println("Running ADDTOKEN")
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(apiBucket))
		dataBucket := tx.Bucket([]byte(dataBucket))
		dataBucket.CreateBucketIfNotExists([]byte(key))
		v := b.Get([]byte(key))
		if v != nil {
			var keys FirebaseKeys
			err := json.Unmarshal(v, &keys)
			if err != nil {
				keys.Keys = helpers.AppendStringIfMissing(keys.Keys, firebase)
				keyData, err := json.Marshal(keys)
				log.Println("Json of Keydata: \n", string(keyData))
				err = b.Put([]byte(key), keyData)
				return err
			}
		} else {
			log.Println("No keys registered for: ", key)
			var keys FirebaseKeys
			keys.Keys = helpers.AppendStringIfMissing(keys.Keys, firebase)
			keyData, err := json.Marshal(keys)
			log.Println("Added Key! ", string(keyData))
			err = b.Put([]byte(key), keyData)
			return err
		}
		return nil
	})
	go replayLastReadings(key)
	return err
}

func replayLastReadings(token string) error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dataBucket))

		data := b.Bucket([]byte(token))
		data.ForEach(func(k, v []byte) error {
			var message WorkRequest
			message.Token = token
			log.Println("Raw Retrieved Value: ", string(v))
			err := json.Unmarshal(v, &message.Data)
			data, err := json.Marshal(message)
			log.Println("Retrieved value: ", string(data))
			//AddJob(&message)
			return err
		})
		return nil
	})
	return err
}
