#!/bin/bash

# the upgrade is a fork, "true" otherwise
FORK=${FORK:-"false"}


BINARY=_build/old/picad
HOME=mytestnet
ROOT=$(pwd)
DENOM=ppica

ADDITIONAL_PRE_SCRIPTS="./scripts/upgrade/old-node-scripts.sh"

SLEEP_TIME=1


screen -L -dmS node1 bash scripts/upgrade/localnode.sh $BINARY $DENOM --Logfile $HOME/log-screen.txt

#screen -L -dmS bash scripts/upgrade/localnode.sh $BINARY $DENOM --Logfile $HOME/log-screen.txt
#scripts/localnode.sh $BINARY

sleep 10 # wait for note to start

# execute additional pre scripts
source $ADDITIONAL_PRE_SCRIPTS 

