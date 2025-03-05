#!/bin/bash

## Network interface is
ping -c 1 nodex | awk 'FNR == 1 {print $3}' 
iproute get result from above | awk 5
