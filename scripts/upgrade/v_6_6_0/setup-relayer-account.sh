#!/bin/bash
KEY=mykey1
RLY_KEY=mykey
KEYALGO="secp256k1"
KEYRING="test"
HOME_DIR="mytestnet"
BINARY=_build/old/picad
DENOM=ppica
CHAINID=centauri-dev


UPGRADE_PRPOSAL_ID=1
RLY_KEY=$($BINARY keys show $RLY_KEY -a --keyring-backend $KEYRING --home $HOME_DIR)
echo "Address of myKEY: $RLY_KEY"


$BINARY tx transmiddleware add-rly --from $KEY $RLY_KEY --keyring-backend test --home $HOME_DIR --chain-id $CHAINID --fees 100000${DENOM}  -y

sleep 5
