#!/bin/bash

# the upgrade is a fork, "true" otherwise
FORK=${FORK:-"false"}

OLD_VERSION=v7_0_2
UPGRADE_WAIT=${UPGRADE_WAIT:-20}
HOME=mytestnet
ROOT=$(pwd)
DENOM=ppica
CHAIN_ID=localpica
SOFTWARE_UPGRADE_NAME="v7_0_3"

SLEEP_TIME=1

if [[ "$FORK" == "true" ]]; then
    export PICA_HALT_HEIGHT=20
fi

# underscore so that go tool will not take gocache into account
mkdir -p _build/gocache
export GOMODCACHE=$ROOT/_build/gocache

# install old binary if not exist
if [ ! -f "_build/$OLD_VERSION.zip" ] &> /dev/null
then
    mkdir -p _build/old
    wget -c "https://github.com/ComposableFi/composable-cosmos/archive/refs/tags/${OLD_VERSION}.zip" -O _build/${OLD_VERSION}.zip
    unzip _build/${OLD_VERSION}.zip -d _build
fi

# reinstall old binary
if [ $# -eq 1 ] && [ $1 == "--reinstall-old" ] || ! command -v _build/old/picad &> /dev/null; then
    cd ./_build/composable-cosmos-${OLD_VERSION:1}
    GOBIN="$ROOT/_build/old" go install -mod=readonly ./...
    cd ../..
fi

# install new binary
if ! command -v _build/new/picad &> /dev/null
then
    mkdir -p _build/new
    GOBIN="$ROOT/_build/new" go install -mod=readonly ./...
fi

# run old node
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "running old node"
    screen -L -dmS node1 bash scripts/localnode-upgrade.sh _build/old/picad $DENOM --Logfile $HOME/log-screen.txt
else
    screen -L -Logfile $HOME/log-screen.txt -dmS node1 bash scripts/localnode-upgrade.sh _build/old/picad $DENOM
fi

sleep 5 # wait for note to start 

# execute additional pre scripts
if [ ! -z "$ADDITIONAL_PRE_SCRIPTS" ]; then
    # slice ADDITIONAL_SCRIPTS by ,
    SCRIPTS=($(echo "$ADDITIONAL_PRE_SCRIPTS" | tr ',' ' '))
    for SCRIPT in "${SCRIPTS[@]}"; do
         # check if SCRIPT is a file
        if [ -f "$SCRIPT" ]; then
            echo "executing additional pre scripts from $SCRIPT"
            source $SCRIPT _build/old/picad
            echo "CONTRACT_ADDRESS = $CONTRACT_ADDRESS"
            sleep 5
        else
            echo "$SCRIPT is not a file"
        fi
    done
fi

run_fork () {
    echo "forking"

    while true; do
        BLOCK_HEIGHT=$(./_build/old/picad status | jq '.SyncInfo.latest_block_height' -r)
        # if BLOCK_HEIGHT is not empty
        if [ ! -z "$BLOCK_HEIGHT" ]; then
            echo "BLOCK_HEIGHT = $BLOCK_HEIGHT"
            sleep 10
        else
            echo "BLOCK_HEIGHT is empty, forking"
            break
        fi
    done
}

run_upgrade () {
    echo -e "\n\n=> =>start upgrading"

    # Get upgrade height, 12 block after (6s)
    STATUS_INFO=($(./_build/old/picad status --home $HOME | jq -r '.NodeInfo.network,.SyncInfo.latest_block_height'))
    UPGRADE_HEIGHT=$((STATUS_INFO[1] + 20))
    echo "UPGRADE_HEIGHT = $UPGRADE_HEIGHT"

    tar -cf ./_build/new/picad.tar -C ./_build/new picad
    SUM=$(shasum -a 256 ./_build/new/picad.tar | cut -d ' ' -f1)
    UPGRADE_INFO=$(jq -n '
    {
        "binaries": {
            "linux/amd64": "file://'$(pwd)'/_build/new/picad.tar?checksum=sha256:'"$SUM"'",
        }
    }')

   echo "submit upgrade"

    ./_build/old/picad tx gov submit-legacy-proposal software-upgrade "$SOFTWARE_UPGRADE_NAME" --upgrade-height $UPGRADE_HEIGHT --upgrade-info "$UPGRADE_INFO" --title "upgrade" --description "upgrade"  --from=test1 --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y

    sleep $SLEEP_TIME

    ./_build/old/picad tx gov deposit 1 "20000000${DENOM}" --from test1 --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    ./_build/old/picad tx gov vote 1 yes --from test0 --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    ./_build/old/picad tx gov vote 1 yes --from test1 --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    # determine block_height to halt
    while true; do
        BLOCK_HEIGHT=$(./_build/old/picad status | jq '.SyncInfo.latest_block_height' -r)
        if [ $BLOCK_HEIGHT = "$UPGRADE_HEIGHT" ]; then
            # assuming running only 1 picad
            echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, KILLING OLD ONE"
            pkill picad
            break
        else
            ./_build/old/picad q gov proposal 1 --output=json | jq ".status"
            echo "BLOCK_HEIGHT = $BLOCK_HEIGHT"
            sleep 1 
        fi
    done
}

# if FORK = true
if [[ "$FORK" == "true" ]]; then
    run_fork
    unset PICA_HALT_HEIGHT
else
    run_upgrade
fi

sleep 1

# run new node
echo -e "\n\n=> =>continue running nodes after upgrade"   
if [[ "$OSTYPE" == "darwin"* ]]; then
    CONTINUE="true" screen -L -dmS picad bash scripts/localnode-upgrade.sh _build/new/picad $DENOM
else
    CONTINUE="true" screen -L -dmS picad bash scripts/localnode-upgrade.sh _build/new/picad $DENOM
fi

sleep 5


# execute additional after scripts
if [ ! -z "$ADDITIONAL_AFTER_SCRIPTS" ]; then
    # slice ADDITIONAL_SCRIPTS by ,
    SCRIPTS=($(echo "$ADDITIONAL_AFTER_SCRIPTS" | tr ',' ' '))
    for SCRIPT in "${SCRIPTS[@]}"; do
         # check if SCRIPT is a file
        if [ -f "$SCRIPT" ]; then
            echo "executing additional after scripts from $SCRIPT"
            source $SCRIPT _build/new/picad
            sleep 5
        else
            echo "$SCRIPT is not a file"
        fi
    done
fi