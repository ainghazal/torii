package share

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const (
	experimentBucket = "exp"
)

func InitDB() (*bolt.DB, error) {
	db, err := bolt.Open("data/experiments.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(experimentBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
