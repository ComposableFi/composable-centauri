#!/bin/bash
KEY=mykey
KEYALGO="secp256k1"
KEYRING="test"
HOME_DIR="mytestnet"
BINARY=_build/old/picad
DENOM=ppica
CHAINID=centauri-dev

$BINARY tx gov submit-proposal scripts/08-wasm/ics10_grandpa_cw.wasm.json --from=$KEY --fees 100000${DENOM} --gas auto --keyring-backend test  --home $HOME_DIR  --chain-id $CHAINID -y  

sleep 5
# TODO: fetch the propsoal id dynamically 
$BINARY tx gov deposit "1" "20000000ppica" --from $KEY --fees 100000${DENOM} --gas auto --keyring-backend test --home $HOME_DIR --chain-id $CHAINID -y

sleep 5
$BINARY tx gov vote 1 yes --from $KEY --fees 100000${DENOM} --keyring-backend test --home $HOME_DIR --chain-id $CHAINID -y 


## Voting time is 20s, check in localnode.sh
sleep 20

$BINARY query 08-wasm all-wasm-code --home $HOME_DIR --chain-id $CHAINID