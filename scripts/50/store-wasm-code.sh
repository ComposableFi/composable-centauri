#!/bin/bash
KEY=mykey
DENOM="ppica"
HOME_DIR="mytestnet"
CHAINID="centauri-dev"
BINARY=picad
WASM_CONTRACT_PATH="composable-ibc/target/wasm32-unknown-unknown/release/ics10_grandpa_cw.wasm"

HEX_CHECKSUM=$(sha256sum "$WASM_CONTRACT_PATH" | awk '{ print $1 }')
echo "Hex checksum is: $HEX_CHECKSUM"


# Wait for chain to start 
echo "Waiting for chain to start..."
sleep 10

exit 0
picad keys show mykey --keyring-backend test --home mytestnet

$BINARY tx ibc-wasm store-code $WASM_CONTRACT_PATH --from mykey --keyring-backend test --chain-id $CHAINID --home $HOME_DIR --gas 20002152622 --fees 20020166${DENOM}  -y

# exit 0
# # Fetch proposal id 
# sleep 6
# # $BINARY query gov proposals -o json > /tmp/proposals.json
# # PROPOSAL_ID=$(jq -r '.proposals[-1].id' /tmp/proposals.json)
# PROPOSAL_ID=3  ## fix this
# echo "Proposal ID is: $PROPOSAL_ID"

# # Validator vote yes
# $BINARY tx gov vote $PROPOSAL_ID yes --from $KEY --fees 100000${DENOM} --keyring-backend test --home $HOME_DIR --chain-id $CHAINID -y 

# #Voting time is 20s, check in localnode.sh
# sleep 20

# # Check the status 
# $BINARY query gov proposal $PROPOSAL_ID -o json > /tmp/proposal.json
# STATUS=$(jq -r '.proposal.status' /tmp/proposal.json)
# echo "Proposal status is: $STATUS"

# # Query wasm checksums 
# $BINARY query ibc-wasm checksums
