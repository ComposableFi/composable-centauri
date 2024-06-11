#!/bin/bash
BINARY=$1
KEY="mykey"
CHAINID="localpica"
MONIKER="localtestnet"
KEYALGO="secp256k1"
KEYRING="test"
LOGLEVEL="info"
CONTINUE=${CONTINUE:-"false"}
# to trace evm
#TRACE="--trace"
TRACE=""

HOME_DIR=mytestnet
DENOM=${2:-ppica}


if [ "$CONTINUE" == "true" ]; then
    echo "\n ->> continuing from previous state"
    $BINARY start --home $HOME_DIR --log_level debug
    exit 0
fi


# remove existing daemon
rm -rf $HOME_DIR

# centaurid config keyring-backend $KEYRING
# centaurid config chain-id $CHAINID

# if $KEY exists it should be deleted

echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover --home $HOME_DIR

update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
}

# Allocate genesis accounts (cosmos formatted addresses)
$BINARY add-genesis-account $KEY 100000000000000000000000000$DENOM --keyring-backend $KEYRING --home $HOME_DIR


# Sign genesis transaction
$BINARY  gentx $KEY 1000000000000000000000$DENOM --keyring-backend $KEYRING --chain-id $CHAINID --home $HOME_DIR

update_test_genesis '.app_state["gov"]["params"]["voting_period"]="5s"'
update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="'$DENOM'"'
update_test_genesis '.app_state["gov"]["params"]["min_deposit"]=[{"denom":"'$DENOM'","amount": "1000000"}]'
update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom":"'$DENOM'","amount":"1000"}'
update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="'$DENOM'"'

# sed -i 's/timeout_commit = "5s"/timeout_commit = "500ms"/' $HOME_DIR/config/config.toml

echo "updating.."
sed -i '' 's/timeout_commit = "5s"/timeout_commit = "500ms"/' $HOME_DIR/config/config.toml

