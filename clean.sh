#!/bin/bash

# remove any running processes
set -e
trap 'killall kvstore' SIGINT
cd $(dirname $0)
killall kvstore || true
sleep 0.1

# remove bolt db files 
rm -rf ./db*

# remove badger db folders 
rm -r ./badgerdb-*