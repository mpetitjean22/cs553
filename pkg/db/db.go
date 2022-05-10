package db

import "fmt"

type Database interface {
	PutKey(key string, value []byte) error
	PutKeyReplica(key string, value []byte) error
	GetKey(key string) ([]byte, error)
	DeleteBulkKeys(deleteKey func(string) bool) error
	GetKeyForReplication() (keyCopy, valueCopy []byte, err error)
	DeleteReplicationKey(key, value []byte) (err error)
}

func NewDatabase(dbPath string, dbType string, replica bool) (db Database, closeFunc func() error, err error) {
	if dbType == "bolt" {
		return NewBoltDatabase(dbPath, replica)
	}
	if dbType == "badger" {
		return NewBadgerDatabase(dbPath)
	}
	return nil, nil, fmt.Errorf("Invalid dbType")
}
