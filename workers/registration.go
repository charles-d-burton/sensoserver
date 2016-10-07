package workers

import (
	"encoding/json"
	"errors"
	"log"
	"tempserver/helpers"

	"github.com/boltdb/bolt"
)

var (
	boltDB  *bolt.DB
	boltDir = "./senso.db"
)

const (
	bucketTopics = "topics"
)

func init() {
	db, err := bolt.Open(boltDir, 0600, nil)
	if err != nil {
		panic(err)
	}
	boltDB = db
	boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketTopics))
		if err != nil {
			log.Println(err)
		}
		return nil
	})
}

type Registration struct {
	Topic
	Token string `json:"token"`
}

type Tokens struct {
	Tokens []string `json:"tokens"`
}

//Register ... Register a new endpoint device
func (register Registration) Register() error {
	token := make([]string, 0)
	token = append(token, register.Token)
	tokens := Tokens{
		Tokens: token,
	}

	data, err := json.Marshal(tokens)
	err = boltDB.Update(func(tx *bolt.Tx) error {
		topicBucket := tx.Bucket([]byte(bucketTopics))
		err := topicBucket.Put([]byte(register.Topic.TopicString), data)
		return err
	})
	return err
}

//JoinTopic ... Join a new endpoint device to an already created topic
func (register Registration) JoinTopic() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		data := b.Get([]byte(register.Topic.TopicString))
		if data != nil {
			var tokens Tokens
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.AppendStringIfMissing(tokens.Tokens, register.Token)
				data, err := json.Marshal(tokens)
				log.Println("Tokens joined!", string(data))
				//Update in the background
				go updateBolt(register.Topic.TopicString, data, bucketTopics)
				return err
			}
		} else {
			return errors.New("Topic Not Found")
		}
		return nil
	})
	return err
}

//LeaveTopic ... Remove a device token from a topic
func (register Registration) LeaveTopic() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		data := b.Get([]byte(register.Topic.TopicString))
		if data != nil {
			var tokens Tokens
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.RemoveStringByValue(tokens.Tokens, register.Token)
				log.Println("Token removed: ", register.Token, tokens.Tokens)
				go updateBolt(register.Topic.TopicString, data, bucketTopics)
				return err
			}

		}
		return nil
	})
	return err
}

//RefreshToken ... Describes a struct to update a device token with a refreshed token
type RefreshToken struct {
	NewToken string `json:"newtoken"`
	OldToken string `json:"oldtoken"`
	Topic
}

//Refresh ... Logic to replace one token with another
func (refreshToken RefreshToken) Refresh() error {
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		data := b.Get([]byte(refreshToken.Topic.TopicString))
		if data != nil {
			var tokens Tokens
			err := json.Unmarshal(data, &tokens)
			if err == nil {
				tokens.Tokens = helpers.ReplaceStringByValue(tokens.Tokens, refreshToken.NewToken, refreshToken.OldToken)
				data, err := json.Marshal(tokens)
				go updateBolt(refreshToken.Topic.TopicString, data, bucketTopics)
				return err
			}
		}
		return nil
	})
	return err
}

//Short and sweet function to update bolt
func updateBolt(topic string, message []byte, bucket string) {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(topic), message)
		return err
	})
	log.Println(err)
}

//Discover if the token is already in a topic
func findToken(token string) (bool, string) {
	found := false
	topic := ""
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketTopics))
		c := b.Cursor()
		for key, value := c.First(); key != nil; key, value = c.Next() {
			var tokenStruct Tokens
			err := json.Unmarshal(value, &tokenStruct)
			if err != nil {
				for _, tVal := range tokenStruct.Tokens {
					if tVal == token {
						found = true
						topic = string(key)
					}
				}
			}
		}
		return nil
	})
	return found, topic
}
