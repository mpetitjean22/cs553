#!/bin/bash

make 

kvstore -db-location=db0.db -http-address=127.0.0.2:8080 -config-file=config.yaml -shard=shard0 &
kvstore -db-location=db0-r.db -http-address=127.0.0.22:8080 -config-file=config.yaml -shard=shard0 -replica &

kvstore -db-location=db1.db -http-address=127.0.0.3:8080 -config-file=config.yaml -shard=shard1 &
kvstore -db-location=db1-r.db -http-address=127.0.0.33:8080 -config-file=config.yaml -shard=shard1 -replica &

kvstore -db-location=db2.db -http-address=127.0.0.4:8080 -config-file=config.yaml -shard=shard2 &
kvstore -db-location=db2-r.db -http-address=127.0.0.44:8080 -config-file=config.yaml -shard=shard2 -replica &

kvstore -db-location=db3.db -http-address=127.0.0.5:8080 -config-file=config.yaml -shard=shard3 &
kvstore -db-location=db3-r.db -http-address=127.0.0.55:8080 -config-file=config.yaml -shard=shard3 -replica &

sleep 2

for shard in 127.0.0.2:8080 127.0.0.3:8080; do 
    for i in {1..100}; do 
        curl "http://$shard/put?key=key-$RANDOM&value=value-$RANDOM"
    done
done

wait