package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewDataBase(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name() + "-boltdb")

	db, closeFunc, err := NewBoltDatabase(f.Name(), false)
	if err != nil {
		t.Fatalf("Unexpected error with NewDatabase: %v", err)
	}
	defer closeFunc()

	// test valid Put
	err = db.PutKey("a", []byte("b"))
	if err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}

	// test valid Get
	val, err := db.GetKey("a")
	if err != nil {
		t.Fatalf("Unexpected error with GetKey: %v", err)
	}
	res := bytes.Compare(val, []byte("b"))
	if res != 0 {
		t.Errorf("Unexpected value. Got: %v Expected: %v", val, []byte("b"))
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestNewDataBaseReadOnly(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name() + "-boltdb")

	dbReadOnly, closeFunc, err := NewBoltDatabase(f.Name(), true)
	if err != nil {
		t.Fatalf("Unexpected error with NewDatabase: %v", err)
	}
	defer closeFunc()

	// test invalid Put
	err = dbReadOnly.PutKey("a", []byte("b"))
	if err == nil {
		t.Fatalf("Expected error from writing to a readonly")
	}

	// test put on default bucket
	err = dbReadOnly.PutKeyReplica("a", []byte("b"))
	if err != nil {
		t.Fatalf("Unexpected error with PutKeyReplica: %v", err)
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestGetBulkKeys(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name() + "-boltdb")

	db, closeFunc, err := NewBoltDatabase(f.Name(), false)
	if err != nil {
		t.Fatalf("Unexpected error with NewDatabase: %v", err)
	}
	defer closeFunc()

	// load keys into db
	if err := db.PutKey("key-1", []byte("value-1")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-2", []byte("value-2")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-3", []byte("value-3")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}

	// test getting keys that match a certain function
	keys, err := db.getBulkKeys(func(s string) bool {
		return s == "key-2"
	})

	if len(keys) != 1 {
		t.Fatalf("Unexpected number of keys. Got: %v", keys)
	} else {
		if keys[0] != "key-2" {
			t.Fatalf("Unexpected result, got: %s expected: %s", keys[0], "key-2")
		}
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestDeleteKeys(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name() + "-boltdb")

	db, closeFunc, err := NewBoltDatabase(f.Name(), false)
	if err != nil {
		t.Fatalf("Unexpected error with NewDatabase: %v", err)
	}
	defer closeFunc()

	// load keys into db
	if err := db.PutKey("key-1", []byte("value-1")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-2", []byte("value-2")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-3", []byte("value-3")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}

	// test deleting keys
	if err = db.deleteKeys([]string{"key-1", "key-3"}); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	keys, err := db.getBulkKeys(func(s string) bool {
		return true
	})
	if len(keys) != 1 {
		t.Fatalf("Unexpected number of keys. Got: %v", keys)
	} else {
		if keys[0] != "key-2" {
			t.Fatalf("Unexpected result, got: %s expected: %s", keys[0], "key-2")
		}
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestDeleteBulkKeys(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name() + "-boltdb")

	db, closeFunc, err := NewBoltDatabase(f.Name(), false)
	if err != nil {
		t.Fatalf("Unexpected error with NewDatabase: %v", err)
	}
	defer closeFunc()

	// load keys into db
	if err := db.PutKey("key-1", []byte("value-1")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-2", []byte("value-2")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}
	if err := db.PutKey("key-3", []byte("value-3")); err != nil {
		t.Fatalf("Unexpected error with PutKey: %v", err)
	}

	// test deleting bulk keys
	if err = db.DeleteBulkKeys(func(s string) bool {
		return s == "key-2"
	}); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	keys, err := db.getBulkKeys(func(s string) bool {
		return true
	})
	if len(keys) != 1 {
		t.Fatalf("Unexpected number of keys. Got: %v", keys)
	} else {
		if keys[0] != "key-2" {
			t.Fatalf("Unexpected result, got: %s expected: %s", keys[0], "key-2")
		}
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}
