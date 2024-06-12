#!/bin/bash

KEY="mykey"
KEY1="mykey1"
CHAINID="centauri-dev"
MONIKER="localtestnet"
KEYALGO="secp256k1"
KEYRING="test"
LOGLEVEL="info"
BINARY=$1
# to trace evm
#TRACE="--trace"
TRACE=""

echo "runnode"

HOME_DIR=mytestnet
DENOM=ppica

# remove existing daemon
rm -rf $HOME_DIR


if [ ! -x "$(command -v $BINARY)" ]; then
    echo "Error: Binary $BINARY is not executable or not found."
    exit 1
fi


if [ "$CONTINUE" == "true" ]; then
    echo "\n ->> continuing from previous state"
    $BINARY start --home $HOME_DIR --log_level debug
    exit 0
fi


$BINARY config keyring-backend $KEYRING
$BINARY config chain-id $CHAINID


# if $KEY exists it should be deleted
echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover --home $HOME_DIR
echo "bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort" | $BINARY keys add $KEY1 --keyring-backend $KEYRING --algo $KEYALGO --recover --home $HOME_DIR

$BINARY init $CHAINID --chain-id $CHAINID --default-denom "ppica" --home $HOME_DIR >/dev/null 2>&1

update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && cp $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
}

# Allocate genesis accounts (cosmos formatted addresses)
$BINARY add-genesis-account $KEY 100000000000000000000000000ppica --keyring-backend $KEYRING --home $HOME_DIR
$BINARY add-genesis-account $KEY1 100000000000000000000000000ppica --keyring-backend $KEYRING --home $HOME_DIR

# Sign genesis transaction
$BINARY gentx $KEY 10030009994127689ppica --keyring-backend $KEYRING --chain-id $CHAINID --home $HOME_DIR --moniker val1
mv $HOME_DIR/config/priv_validator_key.json $HOME_DIR/config/priv_validator_key0.json
$BINARY gentx $KEY1 1003000999412768ppica --keyring-backend $KEYRING --chain-id $CHAINID --home $HOME_DIR --output-document $HOME_DIR/config/gentx/1.json
mv $HOME_DIR/config/priv_validator_key.json $HOME_DIR/config/priv_validator_key1.json
update_test_genesis '.app_state["gov"]["params"]["voting_period"]="20s"'
update_test_genesis '.app_state["gov"]["params"]["expedited_voting_period"]="10s"'
update_test_genesis '.app_state["stakingmiddleware"]["params"]["blocks_per_epoch"]="5"'
update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="'$DENOM'"'
update_test_genesis '.app_state["gov"]["params"]["min_deposit"]=[{"denom":"'$DENOM'","amount": "1"}]'
update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom":"'$DENOM'","amount":"1000"}'
update_test_genesis '.app_state["slashing"]["params"]["signed_blocks_window"]="4"'
update_test_genesis '.app_state["slashing"]["params"]["downtime_jail_duration"]="5s"'


# Collect genesis tx
$BINARY collect-gentxs --home $HOME_DIR

# Run this to ensure everything worked and that the genesis file is setup correctly
$BINARY validate-genesis --home $HOME_DIR

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# update request max size so that we can upload the light client
# '' -e is a must have params on mac, if use linux please delete before run
sed -i'' -e 's/max_body_bytes = /max_body_bytes = 1/g' $HOME_DIR/config/config.toml
sed -i'' -e 's/max_tx_bytes = 1048576/max_tx_bytes = 10000000/g' $HOME_DIR/config/config.toml

# Initialize directories for two validators
$BINARY init $CHAINID --chain-id $CHAINID --default-denom "ppica" --home $HOME_DIR/validator1 >/dev/null 2>&1
$BINARY init $CHAINID --chain-id $CHAINID --default-denom "ppica" --home $HOME_DIR/validator2 >/dev/null 2>&1


# Copy the genesis file to each validator's directory
cp $HOME_DIR/config/genesis.json $HOME_DIR/validator1/config/genesis.json
cp $HOME_DIR/config/genesis.json $HOME_DIR/validator2/config/genesis.json
mv $HOME_DIR/config/priv_validator_key0.json $HOME_DIR/validator1/config/priv_validator_key.json
mv $HOME_DIR/config/priv_validator_key1.json $HOME_DIR/validator2/config/priv_validator_key.json

P2PPORT_2=26658
RPCPORT_2=26659
RESTPORT_2=1316
ROSETTA_2=8081
WEB_PORT_2=9091

sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://localhost:'"$P2PPORT_2"'"#g' $HOME_DIR/validator2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://localhost:'"$RPCPORT_2"'"#g' $HOME_DIR/validator2/config/config.toml
sed -i -e 's#"tcp://localhost:26657"#"tcp://localhost:'"$RPCPORT_2"'"#g' $HOME_DIR/validator2/config/client.toml
sed -i -e 's#"tcp://localhost:1317"#"tcp://localhost:'"$RESTPORT_2"'"#g' $HOME_DIR/validator2/config/app.toml
sed -i -e 's#"localhost:9090"#"localhost:'"$WEB_PORT_2"'"#g' $HOME_DIR/validator2/config/app.toml
sed -i -e 's#"127.0.0.1:9090"#"localhost:'"$WEB_PORT_2"'"#g' $HOME_DIR/validator2/config/app.toml
sed -i -e 's#pprof_laddr = "localhost:6060"#pprof_laddr = "localhost:7070"#g' $HOME_DIR/validator2/config/config.toml
sed -i -e 's#":8080"#":'"$ROSETTA_2"'"#g' $HOME_DIR/validator2/config/app.toml


Start each validator with different ports
screen -L -dmS node1 $BINARY start --rpc.unsafe --rpc.laddr tcp://0.0.0.0:26657 --p2p.laddr tcp://0.0.0.0:26656 --home=$HOME_DIR/validator1 --log_level info --trace
NodeID=$($BINARY comet show-node-id --home=$HOME_DIR/validator1)
echo "Node ID: $NodeID"
bin/picad start --home=$HOME_DIR/validator2 --p2p.persistent_peers=$NodeID@127.0.0.1:26656