#!/bin/bash

make  

kvstore -db-location=db0.db -db-type=badger -http-address=127.0.0.2:8080 -config-file=config.yaml -shard=shard0 &
kvstore -db-location=db1.db -db-type=badger -http-address=127.0.0.3:8080 -config-file=config.yaml -shard=shard1 &
kvstore -db-location=db2.db -db-type=badger -http-address=127.0.0.4:8080 -config-file=config.yaml -shard=shard2 &
kvstore -db-location=db3.db -db-type=badger -http-address=127.0.0.5:8080 -config-file=config.yaml -shard=shard3 &

sleep 2

for shard in 127.0.0.2:8080 127.0.0.3:8080; do 
    for i in {1..100}; do 
        curl "http://$shard/put?key=key-$RANDOM&value=value-$RANDOM"
    done
done

wait