package db

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

type BadgerDatabase struct {
	db *badger.DB
}

func NewBadgerDatabase(dbPath string) (db *BadgerDatabase, closeFunc func() error, err error) {
	badgerdb, err := badger.Open(badger.DefaultOptions("badgerdb-" + dbPath))
	if err != nil {
		return nil, nil, err
	}
	db = &BadgerDatabase{db: badgerdb}
	closeFunc = badgerdb.Close
	return db, closeFunc, nil
}

func (db *BadgerDatabase) PutKey(key string, value []byte) error {
	db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	return nil
}

func (db *BadgerDatabase) PutKeyReplica(key string, value []byte) error {
	return fmt.Errorf("not implemented")
}

func (db *BadgerDatabase) GetKey(key string) ([]byte, error) {
	var result []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		result, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *BadgerDatabase) getBulkKeys(getKey func(string) bool) ([]string, error) {
	var keys []string
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if getKey(string(item.Key())) {
				keys = append(keys, string(item.Key()))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *BadgerDatabase) deleteKeys(keys []string) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Delete([]byte(key)); err != nil {
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

func (db *BadgerDatabase) DeleteBulkKeys(deleteKey func(string) bool) error {
	keys, err := db.getBulkKeys(deleteKey)
	if err != nil {
		return err
	}
	fmt.Println(keys)

	if err = db.deleteKeys(keys); err != nil {
		return err
	}
	return nil
}

func (db *BadgerDatabase) GetKeyForReplication() (keyCopy, valueCopy []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (db *BadgerDatabase) DeleteReplicationKey(key, value []byte) (err error) {
	return fmt.Errorf("not implemented")
}
