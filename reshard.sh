#!/bin/bash

# remove any running processes
set -e
trap 'killall kvstore' SIGINT
cd $(dirname $0)
killall kvstore || true
sleep 0.1

# copy node 0 db to node 2 
cp ./db0.db-boltdb ./db2.db-boltdb

# copy node 1 db to node 3
cp ./db1.db-boltdb ./db3.db-boltdb

./bolt_demo.sh

# run /clean on all of the nodes 
curl 'http://127.0.0.2:8080/clean'
curl 'http://127.0.0.3:8080/clean'
curl 'http://127.0.0.4:8080/clean'
curl 'http://127.0.0.5:8080/clean'