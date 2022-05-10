package main

import (
	"cs553/pkg/api"
	"cs553/pkg/config"
	"cs553/pkg/db"
	"cs553/pkg/replication"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	dbLocation  = flag.String("db-location", "", "path for the db")
	httpAddress = flag.String("http-address", "127.0.0.1:8080", "HTTP host address")
	configFile  = flag.String("config-file", "config.yaml", "config file for sharding")
	dbType      = flag.String("db-type", "bolt", "which DB to use for storage, bolt or badger")
	shardName   = flag.String("shard", "", "name of shard for data")
	replica     = flag.Bool("replica", false, "run as a read-only replica")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatalf("Must provide db-location")
	}

	if (*dbType != "bolt") && (*dbType != "badger") {
		log.Fatalf("db-type must be one of bolt or badger")
	}

	if (*dbType == "badger") && (*replica) {
		log.Fatalf("replicas are not supported with badger")
	}
}

func main() {
	// parse the input flags
	parseFlags()

	// construct the DB
	newdb, close, err := db.NewDatabase(*dbLocation, *dbType, *replica)
	if err != nil {
		log.Fatalf("NewDatabase(%q): %v", *dbLocation, err)
	}
	defer close()

	// parse config
	config, err := config.NewConfig(*configFile, *shardName)
	if err != nil {
		log.Fatalf("Could not construct config with error: %v \n", err)
	}
	if config.ShardIndex == -1 {
		log.Fatalf("Could not find shard with name %v in config \n", *shardName)
	}

	if *replica {
		masterAddress, ok := config.ShardToAddress[config.ShardIndex]
		if !ok {
			log.Fatalf("Could not address for master shard")
		}
		go replication.PropagateReplication(newdb, masterAddress)
	}

	// set up the api http server
	ws := api.NewWebServer(newdb, config)
	http.HandleFunc("/get", ws.GetHandler)
	http.HandleFunc("/put", ws.PutHandler)
	http.HandleFunc("/clean", ws.CleanHandler)
	http.HandleFunc("/get-next-replication-key", ws.GetNextReplicationKeyHandler)
	http.HandleFunc("/delete-next-replication-key", ws.DeleteReplicationKeyHandler)

	fmt.Printf("Spinning up http server at %s with shard %s at index %d \n", *httpAddress, *shardName, config.ShardIndex)
	log.Fatal(http.ListenAndServe(*httpAddress, nil))
}
