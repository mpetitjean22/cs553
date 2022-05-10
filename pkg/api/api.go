package api

import (
	"cs553/pkg/config"
	"cs553/pkg/db"
	"cs553/pkg/replication"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
)

type WebServer struct {
	db     db.Database
	config *config.Config
}

func NewWebServer(db db.Database, config *config.Config) *WebServer {
	return &WebServer{
		db:     db,
		config: config,
	}
}

func (ws *WebServer) redirectToCorrectShard(shardIndex int, w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://" + ws.config.ShardToAddress[shardIndex] + r.RequestURI)
	fmt.Fprintf(w, "redirecting from shard %d to shard %d \n", ws.config.ShardIndex, shardIndex)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error with redirecting request %v \n", err)
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
	return
}

func (ws *WebServer) getKeyHash(key string) int {
	h := fnv.New64()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(ws.config.TotalShards))
}

func (ws *WebServer) PutHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("value")

	shardIndex := ws.getKeyHash(key)
	if shardIndex != ws.config.ShardIndex {
		ws.redirectToCorrectShard(shardIndex, w, r)
		return
	}

	err := ws.db.PutKey(key, []byte(val))
	fmt.Fprintf(w, "Key= %q, hash = %d, Value = %q, Error = %v \n", key, shardIndex, val, err)
}

func (ws *WebServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	shardIndex := ws.getKeyHash(key)
	if shardIndex != ws.config.ShardIndex {
		ws.redirectToCorrectShard(shardIndex, w, r)
		return
	}

	val, err := ws.db.GetKey(key)
	fmt.Fprintf(w, "Value = %q, Error = %v \n", val, err)
}

func (ws *WebServer) CleanHandler(w http.ResponseWriter, r *http.Request) {
	err := ws.db.DeleteBulkKeys(func(key string) bool {
		return ws.getKeyHash(key) != ws.config.ShardIndex
	})
	fmt.Fprintf(w, "Error: %v \n", err)
}

func (ws *WebServer) GetNextReplicationKeyHandler(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	key, value, err := ws.db.GetKeyForReplication()
	encoder.Encode(&replication.ReplicateKeyValue{
		Key:   string(key),
		Value: string(value),
		Err:   err,
	})
}

func (ws *WebServer) DeleteReplicationKeyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get("value")

	err := ws.db.DeleteReplicationKey([]byte(key), []byte(value))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "recevied error: %v \n", err)
		return
	}
	fmt.Fprint(w, "ok \n")
}
