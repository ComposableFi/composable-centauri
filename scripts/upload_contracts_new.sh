#!/bin/bash

KEY="mykey"
KEYALGO="secp256k1"
KEYRING="test"

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

centaurid tx ibc-wasm store-code contracts/ics10_grandpa_cw.wasm --from=mykey --title test --summary test --gas 10002152622 --fees 10020166ppica --deposit 100000000ppica --keyring-backend test --chain-id=localpica -y
sleep 5
centaurid tx gov vote 1 yes --from=mykey --gas 10002152622 --fees 10020166ppica --keyring-backend test --chain-id=localpica -y