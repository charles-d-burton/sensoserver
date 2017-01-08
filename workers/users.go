package workers

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
)

type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func GetUser(user_id, email string) (User, error) {
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
			go registerAPIKey(user.Token)
			return err
		} else {
			log.Println("User found")
			err1 := json.Unmarshal(v, &user)
			return err1
		}
	})
	log.Println("Returning from GetUser")
	return user, err
}

type APIToken struct {
	Token string `json:"token"`
}

func registerAPIKey(key string) error {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(apiBucket))
		v := b.Get([]byte(key))
		if v == nil {
			log.Println("API Key not registered, registering new API Key")
			err := b.Put([]byte(key), []byte("valid"))
			return err
		}
		return nil
	})
	return err
}
