package db

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

const SERVICEBUCKETNAME = "Services"

const ERRORFILEEXISTS = "file exists"

type BoltDB struct {
	Conn   *bolt.DB
	Bucket string
}

func NewBoltDB(bucket string) *BoltDB {
	conn := connect(bucket)
	return &BoltDB{Conn: conn, Bucket: bucket}
}

func connect(bucket string) *bolt.DB {
	err := os.Mkdir("/usr/local/bin/gateway", 0777)

	if err != nil {
		if err.Error() != ERRORFILEEXISTS {

		}
	}
	db, err := bolt.Open("/usr/local/bin/gateway/service.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	_ = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return db
}
