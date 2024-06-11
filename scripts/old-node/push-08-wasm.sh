#!/bin/bash

KEY="mykey"
KEYALGO="secp256k1"
KEYRING="test"
HOME_DIR="mytestnet"
# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

./_build/old/centaurid tx 08-wasm push-wasm contracts/ics10_grandpa_cw.wasm --from=mykey --gas 10002152622 --fees 10020166upica --keyring-backend test --chain-id=localpica -y  --home $HOME_DIR

sleep 5

./_build/old/centaurid query 08-wasm all-wasm-code --home $HOME_DIR