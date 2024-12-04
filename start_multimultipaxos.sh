#!/bin/bash

all_nodes_addrs=($(echo "@.all.host@" | tr ',' ' '))

N=$(expr @.all | length@ - 1)
if [ "@.me.host@" == "@.all[0].host@" ]; then
	./lib/master -N=5 -port=7087 &
  ./lib/master -N=5 -port=8086 &
#  ./lib/master -N=5 -port=9085 &
#  ./lib/master -N=5 -port=10084 &
#  ./lib/master -N=5 -port=11083 &
else
	./lib/server -maddr=@.me.maddr@ -mport=7087 -addr=@.me.host@ -port=7070 -p=8 -dreply=true -e=false -exec=true &
	./lib/server -maddr=@.me.maddr@ -mport=8086 -addr=@.me.host@ -port=8071 -p=8 -dreply=true -e=false -exec=true &
#	./lib/server -maddr=@.me.maddr@ -mport=9085 -addr=@.me.host@ -port=9072 -p=8 -dreply=true -e=false -exec=true &
#	./lib/server -maddr=@.me.maddr@ -mport=10084 -addr=@.me.host@ -port=5073 -p=8 -dreply=true -e=false -exec=true &
#	./lib/server -maddr=@.me.maddr@ -mport=11083 -addr=@.me.host@ -port=4074 -p=8 -dreply=true -e=false -exec=true &
fi
