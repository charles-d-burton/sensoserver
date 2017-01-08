package workers

import (
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	topicsBucket = "topics"
	usersBucket  = "users"
	apiBucket    = "api"
)

var (
	boltDB    *bolt.DB
	words     = 2
	separator = "-"
)

func StartBolt(boltDir string, boltPerms os.FileMode) error {
	db, err := bolt.Open(boltDir, boltPerms, nil)
	if err != nil {
		panic(err)
	}
	boltDB = db
	err = boltDB.Update(func(tx *bolt.Tx) error {
		create, err := tx.CreateBucketIfNotExists([]byte(topicsBucket))
		create, err = tx.CreateBucketIfNotExists([]byte(usersBucket))
		create, err = tx.CreateBucketIfNotExists([]byte(apiBucket))
		log.Println("Stats: ", create.Stats().Depth)
		if err != nil {
			log.Println(err)
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
