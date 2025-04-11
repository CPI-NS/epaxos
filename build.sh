#
#/bin/bash
#
# Script to build the epaxos parts
# 
# Usage: `./build.sh` to build just the core epaxos stuff
#        `./build.sh --eaas` to build the eaas client
# 


echo "Building Master"
go build -o bin/master ./src/master
echo "Building Server"
go build -o bin/server ./src/server
echo "Building Client"
go build -o bin/client ./src/client
echo "Building Open Loop Client"
go build -o bin/client-ol ./src/client-ol-lat

if [ "$1" == "--eaas" ]; then
  echo "Building EaaS"
  go build -o bin/eaasclient ./src/eaasclient

elif [ "$1" != "" ]; then
  echo "Invalid argument: $1"
fi


## TO RUN: (from the ./src readme)
#
# bin/master &
# bin/server -port 7070 &
# bin/server -port 7071 &
# bin/server -port 7072 &
# bin/client
