package replication

import (
	"bytes"
	"cs553/pkg/db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ReplicateKeyValue struct {
	Key   string
	Value string
	Err   error
}

type ReplicationClient struct {
	db            db.Database
	masterAddress string
}

func PropagateReplication(db db.Database, masterAddress string) {
	rc := &ReplicationClient{
		db:            db,
		masterAddress: masterAddress,
	}
	for {
		backlog, err := rc.replicationLoop()
		if err != nil {
			log.Printf("eror with replicationLoop: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if !backlog {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (rc *ReplicationClient) replicationLoop() (bool, error) {
	resp, err := http.Get("http://" + rc.masterAddress + "/get-next-replication-key")
	if err != nil {
		return false, err
	}
	var repKV ReplicateKeyValue
	if err := json.NewDecoder(resp.Body).Decode(&repKV); err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	if repKV.Err != nil {
		return false, repKV.Err
	}
	if repKV.Key == "" {
		return false, nil
	}

	if err := rc.db.PutKeyReplica(repKV.Key, []byte(repKV.Value)); err != nil {
		return false, err
	}

	if err := rc.deleteFromQueue(repKV.Key, repKV.Value); err != nil {
		log.Printf("deleteFromQueue failed with: %v", err)
	}

	log.Printf("Next key value %+v", repKV)
	return true, nil
}

func (rc *ReplicationClient) deleteFromQueue(key, value string) error {
	u := url.Values{}
	u.Set("key", key)
	u.Set("value", value)

	log.Printf("Deleting key=%q, value=%q from replication queue on %q", key, value, rc.masterAddress)

	resp, err := http.Get("http://" + rc.masterAddress + "/delete-next-replication-key?" + u.Encode())
	if err != nil {
		return err
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(out, []byte("ok \n")) {
		return fmt.Errorf(string(out))
	}
	return nil
}
