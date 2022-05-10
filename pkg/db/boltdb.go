package db

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
)

var defaultBucket = []byte("default")
var replicaBucket = []byte("replica")

type BoltDatabase struct {
	db      *bolt.DB
	replica bool
}

func NewBoltDatabase(dbPath string, replica bool) (db *BoltDatabase, closeFunc func() error, err error) {
	boltdb, err := bolt.Open(dbPath+"-boltdb", 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	// Dangerous! But will improve write throughput
	boltdb.NoSync = true

	db = &BoltDatabase{db: boltdb, replica: replica}
	closeFunc = boltdb.Close

	if err := db.createBuckets(); err != nil {
		closeFunc()
		return nil, nil, fmt.Errorf("creating buckets: %w", err)
	}

	return db, closeFunc, nil
}

func (db *BoltDatabase) createBuckets() error {
	return db.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(defaultBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(replicaBucket); err != nil {
			return err
		}
		return nil
	})
}

func (db *BoltDatabase) PutKey(key string, value []byte) error {
	if db.replica {
		return fmt.Errorf("replicas only allow read operations")
	}
	db.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(defaultBucket).Put([]byte(key), value); err != nil {
			return err
		}

		if err := tx.Bucket(replicaBucket).Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (db *BoltDatabase) PutKeyReplica(key string, value []byte) error {
	db.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(defaultBucket).Put([]byte(key), value)
	})
	return nil
}

func (db *BoltDatabase) GetKey(key string) ([]byte, error) {
	var result []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		result = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *BoltDatabase) getBulkKeys(getKey func(string) bool) ([]string, error) {
	var keys []string
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		return b.ForEach(func(key, value []byte) error {
			if getKey(string(key)) {
				keys = append(keys, string(key))
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *BoltDatabase) deleteKeys(keys []string) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		for _, key := range keys {
			if err := b.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *BoltDatabase) DeleteBulkKeys(deleteKey func(string) bool) error {
	keys, err := db.getBulkKeys(deleteKey)
	if err != nil {
		return err
	}

	if err = db.deleteKeys(keys); err != nil {
		return err
	}
	return nil
}

func copyValueIntoSlice(b []byte) []byte {
	if b == nil {
		return nil
	}
	res := make([]byte, len(b))
	copy(res, b)
	return res
}

func (db *BoltDatabase) GetKeyForReplication() (keyCopy, valueCopy []byte, err error) {
	err = db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(replicaBucket)
		key, value := b.Cursor().First()
		keyCopy = copyValueIntoSlice(key)
		valueCopy = copyValueIntoSlice(value)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return keyCopy, valueCopy, nil
}

func (db *BoltDatabase) DeleteReplicationKey(key, value []byte) (err error) {
	err = db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(replicaBucket)
		replicaValue := b.Get(key)
		if replicaValue == nil {
			return fmt.Errorf("key does not exist")
		}
		if !bytes.Equal(replicaValue, value) {
			return fmt.Errorf("values do not match.")
		}
		return b.Delete(key)
	})
	if err != nil {
		return err
	}
	return nil
}
