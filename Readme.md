# Simple Distributed KV Store in Go
- [Simple Distributed KV Store in Go](#simple-distributed-kv-store-in-go)
  - [Getting Started](#getting-started)
  - [Configurations](#configurations)
  - [Demo](#demo)
    - [Simple BoltDB](#simple-boltdb)
    - [Simple BadgerDB](#simple-badgerdb)
    - [Replicas](#replicas)
    - [Adding More Nodes](#adding-more-nodes)
  - [Benchmarks](#benchmarks)


## Getting Started 
We first need to make sure that we can build and run the go program. For convience we have a make file so we simply need to verify that we can do that following: 
```sh
$ make kvstore 
# there should be no errors, but if you receive something like
# "command not found" then you need to add the file to your path.
# You can use the following command to fix this: 
$ export PATH="$PATH:$(dirname $(go list -f '{{.Target}}' cs553/cmd/... | head -n 1))"
```
Once we can run `make` without any issues, then we can spin up a node with the following: 
```sh
$ kvstore -db-location=db0.db -http-address=127.0.0.2:8080 -config-file=config.yaml -shard=shard0
```
Other flags that we can set include `-db-type` which specifies whether to use badgerDB or BoltDB, by default it is BoltDB. We can also use the `-replica` flag to indicate if a shard is a replica, for this shards the shard name needs to be the same as the master node that it is a replica of. To get the full list of flags and their descriptions, run: 
``` sh
$ kvstore -h
```
When modifying flags, you may need to adjust the `config.yaml` file. For more information about how to do this go to the Configurations section. 

## Configurations 
In order to make the program easier to use, the `config.yaml` file allows the user to specify configurations:
``` yaml
Shard: 
  Name: shard0
  Index: 0
  Address: 127.0.0.2:8080
  Replicas: [127.0.0.22:8080]
---
Shard: 
  Name: shard1
  Index: 1
  Address: 127.0.0.3:8080
  Replicas: [127.0.0.33:8080]
```
Each Shard object is seperated by `---` and is assigned a name and a unique index. The indices must be unique and no indices can be skipped, cannot have 0 and 2 for example, but they may be out of order within the config file. The Address must match the `-http-address` flag when spinning up the http server with `kvstore`. 

## Demo
### Simple BoltDB 
We have several scripts to make it easier to demo this program. We can start with running:
``` sh
$ ./bolt_demo.sh
```
If there is a problem like, `listen tcp 127.0.0.6:8080: bind: can't assign requested address` then you may need to run `sudo ifconfig lo0 alias 127.0.0.6`. Do this for each IP address where this problem comes up for. 

This script will create 4 nodes with no replicas using boltDB. It will then add 100 key/value paris to nodes 1 and 2. Once the nodes are spun up and the key/value pairs are done being added then we can send requests to the nodes. For example, in a seperate terminal we can do the following: 
``` sh 
# put request to node 0, hosted on 127.0.0.2:8080
$ curl 'http://127.0.0.2:8080/put?key=key-1&value=value-1'

# get request to node 3, hosted on 127.0.0.4:8080
$ curl 'http://127.0.0.5:8080/get?key=key-1'
```
If the key of the request is not in that shard, then the request will automatically be redirected to the correct shard for both put and get requests. 

To shut down the http servers, we simply need to do `^C` in the terminal that is running them. We can then run the script: 
``` sh
$ ./clean.sh
```
which will kill any open processes running this program and will also remove the db files. 

### Simple BadgerDB 
To run a demo that uses BadgerDB instead of BoltDB then we have the following script 
``` sh
$ ./badger_demo.sh
```
From the user perspective, the scripts are exactly the same except we use the `-db-type=badger` flag. The way that the client interacts is the same as above. 

### Replicas 
To run a demo that also has replicas then we have the following script: 
``` sh
$ ./replica_demo.sh
``` 
It is important to note that the `-replica` flag is only supported when using BoltDB. To play with the replicas we see that:
``` sh 
# add a key/value pair to node 0 
$ curl 'http://127.0.0.2:8080/put?key=key-1&value=value-1'
Key= "key-1", hash = 0, Value = "value-1", Error = <nil> 

# verify that is in node 0 master node
$ curl 'http://127.0.0.2:8080/get?key=key-1'
Value = "value-1", Error = <nil> 

# verify that is in the node 0 replica node 
$ curl 'http://127.0.0.22:8080/get?key=key-1'
Value = "value-1", Error = <nil> 

# try and update the key/value pair at the replica node and verify error
$ curl 'http://127.0.0.22:8080/put?key=key-1&value=value-2'
Key= "key-1", hash = 0, Value = "value-2", Error = replicas only allow read operations 

# update key/value pair by sending request to node 3 
$ curl 'http://127.0.0.5:8080/put?key=key-1&value=value-2'
redirecting from shard 3 to shard 0 
Key= "key-1", hash = 0, Value = "value-2", Error = <nil> 

# verify it is updated in node 0 replica 
$ curl 'http://127.0.0.22:8080/get?key=key-1'
Value = "value-2", Error = <nil> 
```
In this small demo we can see that when we update a key/value pair by sending a request to any node, even if does not contain the key itself, it is redirected to the correct master node and is subsequently automatically updated in the replica, which we are able to verify. 

### Adding More Nodes 
In order to demo how to add additional nodes, we will have to modify some of the exisiting `bolt_demo.sh` script and the `config.yaml` file. First we begin by spinning up two nodes. We will want to comment out the last two `kvstore` commands in `bolt_demo.sh` and the last two shard objects in `config.yaml`. 

The middle sections of `bolt_demo.sh` should look like: 
``` sh
kvstore -db-location=db0.db -http-address=127.0.0.2:8080 -config-file=config.yaml -shard=shard0 &
kvstore -db-location=db1.db -http-address=127.0.0.3:8080 -config-file=config.yaml -shard=shard1 &
# kvstore -db-location=db2.db -http-address=127.0.0.4:8080 -config-file=config.yaml -shard=shard2 &
# kvstore -db-location=db3.db -http-address=127.0.0.5:8080 -config-file=config.yaml -shard=shard3 &
```
and `config.yaml` should look like: 
``` sh
Shard: 
  Name: shard0
  Index: 0
  Address: 127.0.0.2:8080
---
Shard: 
  Name: shard1
  Index: 1
  Address: 127.0.0.3:8080
#---
#Shard: 
#  Name: shard2
#  Index: 2
#  Address: 127.0.0.4:8080
#---
#Shard: 
#  Name: shard3 
#  Index: 3
#  Address: 127.0.0.5:8080
```
Running the script: 
``` sh
$ ./bolt_demo.sh
```
will spin up two nodes and will create 100 key/value pairs which will be added to either node 0 or node 1. From the list of key/value pairs written, we can find one that was written to node 0 and one that was written to node 1. For example,
``` sh 
# stored in node 0 
$ curl 'http://127.0.0.2:8080/get?key=key-17007'
Value = "value-30045", Error = <nil> 

# stored in node 1 
$ curl 'http://127.0.0.3:8080/get?key=key-68'
Value = "value-26119", Error = <nil> 
```
We then introduce enough nodes to get to the next power of 2, which would be 4. To do this, we stop the script with `^C` and uncomment the portions that we commented out in `bolt_demo.sh` and `config.yaml`. Now we can run: 
``` sh
$ ./reshard.sh
```
This script will do several things. First it will kill all of the current running processes of `kvstore`. It will then copy over the db from node 0 to node 2, and from node 1 to node 3, read the write up to understand why we do this. We then spin up the 4 nodes by running `./bolt_demo.sh`. Finally we make a request to `/clean` on all 4 nodes which will remove any keys which do not belong to this shard. Once this is complete then we can verify that: 
``` sh
# make get request for key-17007 which we verified earlier belonged to 
# node 1 but now it has been resharded to node 2, and is no longer 
# on node 1. 
$ curl 'http://127.0.0.2:8080/get?key=key-17007'
redirecting from shard 0 to shard 2 
Value = "value-30045", Error = <nil> 
```
## Benchmarks
We also have a small program which will run read and write benchmarks. In order to use it, we first have to spin up some nodes, we can do this easily with: 
``` sh
$ ./bolt_demo.sh
# or 
$ ./badger_demo.sh
```
We can run the benchmark with the following: 
``` sh 
$ make perf 
$ perf -address=127.0.0.2:8080 -write-iterations=100 -read-iterations=100 -concurrency=2
Running with 100 write iterations and concurency of 2
Func write took 230.838µs avg, 4332.0 QPS, 1.597126ms max, 106.049µs min
Func write took 240.104µs avg, 4164.8 QPS, 1.734227ms max, 108.821µs min
Total write QPS: 8496.9 and Number of Set Keys: 200 
Running with 100 read iterations and concurency of 2
Func read took 205.137µs avg, 4874.8 QPS, 400.233µs max, 98.161µs min
Func read took 209.923µs avg, 4763.6 QPS, 401.233µs max, 92.171µs min
Total Read QPS: 9638.4 
```
The `-address` flag specifies which node we are benchmarking and `-write-iterations` and `-read-iterations` flags specifies the number of random read/writes that we want to do. Finally concurrency specifies the number of goroutines that we run in paraellel. 