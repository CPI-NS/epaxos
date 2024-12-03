#!/bin/bash
if [ "$(hostname)" == "node7" ]; then
  export LEADER=0
  ./lib/ClientDriverGo -maddr=server1 -mport=7087 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node8" ]; then
  export LEADER=1
  ./lib/ClientDriverGo -maddr=server1 -mport=8086 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node9" ]; then
  export LEADER=2
  ./lib/ClientDriverGo -maddr=server1 -mport=9085 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node10" ]; then
  export LEADER=3
  ./lib/ClientDriverGo -maddr=server1 -mport=10084 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node11" ]; then
  export LEADER=4
  ./lib/ClientDriverGo -maddr=server1 -mport=11083 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node12" ]; then
  export LEADER=0
  ./lib/ClientDriverGo -maddr=server1 -mport=7087 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node13" ]; then
  export LEADER=1
  ./lib/ClientDriverGo -maddr=server1 -mport=8086 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node14" ]; then
  export LEADER=2
  ./lib/ClientDriverGo -maddr=server1 -mport=9085 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node15" ]; then
  export LEADER=3
  ./lib/ClientDriverGo -maddr=server1 -mport=10084 -p=8 -e=false -smp=true
elif [ "$(hostname)" == "node16" ]; then
  export LEADER=4
  ./lib/ClientDriverGo -maddr=server1 -mport=11083 -p=8 -e=false -smp=true
else 
  echo "Invalid client script (check node addresses in script?)"
fi
