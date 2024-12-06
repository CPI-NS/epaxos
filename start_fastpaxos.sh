#!/bin/bash

all_nodes_addrs=($(echo "@.all.host@" | tr ',' ' '))

N=$(expr @.all | length@ - 1)
if [ "@.me.host@" == "@.all[0].host@" ]; then
	./lib/master -N=5
else
	./lib/server -maddr=@.me.maddr@ -mport=@.me.mport@ -addr=@.me.host@ -port=@.me.port@ -p=8 -dreply=true -e=false -exec=true
fi
