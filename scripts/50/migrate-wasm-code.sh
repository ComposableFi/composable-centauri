#!/bin/bash
KEY=mykey
DENOM="ppica"
HOME_DIR="mytestnet"
CHAINID="centauri-dev"
BINARY=_build/new/picad
WASM_CONTRACT_PATH="composable-ibc/target/wasm32-unknown-unknown/release/ics10_grandpa_cw.wasm"

WASM_CLIENT_ID="08-wasm-0"
HEX_CHECKSUM=$(sha256sum "$WASM_CONTRACT_PATH" | awk '{ print $1 }')
echo "Hex checksum is: $HEX_CHECKSUM"

# Convert HEX_CHECKSUM to raw bytes
RAW_CHECKSUM=$(echo "$HEX_CHECKSUM" | xxd -r -p)

# Convert raw bytes to Base64
BASE64_CHECKSUM=$(echo -n "$RAW_CHECKSUM" | base64)
echo "Base 64 check sum is: $BASE64_CHECKSUM"

$BINARY tx ibc-wasm migrate-contract $WASM_CLIENT_ID $HEX_CHECKSUM {}  --from $KEY --keyring-backend test --chain-id $CHAINID  --home $HOME_DIR --gas 20002152622 --fees 20020166${DENOM} -y

# exit 0
# sleep 6
# # $BINARY query gov proposals -o json > /tmp/proposals.json
# PROPOSAL_ID=5
# echo "Proposal ID is: $PROPOSAL_ID"

# # Validator vote yes
# $BINARY tx gov vote $PROPOSAL_ID yes --from $KEY --fees 100000${DENOM} --keyring-backend test --home $HOME_DIR --chain-id $CHAINID -y 

# #Voting time is 20s, check in localnode.sh
# sleep 20

# # Check the status 
# $BINARY query gov proposal $PROPOSAL_ID -o json > /tmp/proposal.json
# STATUS=$(jq -r '.proposal.status' /tmp/proposal.json)
# echo "Proposal status is: $STATUS"


