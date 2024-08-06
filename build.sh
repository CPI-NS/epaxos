#! /bin/bash
#
# Script to build the epaxos parts
#
# ###################################


go build -o bin/master ./src/master
go build -o bin/server ./src/server
go build -o bin/client ./src/client



## TO RUN: (from the ./src readme)
#
# bin/master &
# bin/server -port 7070 &
# bin/server -port 7071 &
# bin/server -port 7072 &
# bin/client
