#!/bin/bash

# the upgrade is a fork, "true" otherwise
FORK=${FORK:-"false"}

OLD_VERSION=kien-v6.5.0-tweak
UPGRADE_WAIT=${UPGRADE_WAIT:-20}
HOME=mytestnet
ROOT=$(pwd)
DENOM=ppica
CHAIN_ID=localpica
SOFTWARE_UPGRADE_NAME="v7_0_1"
ADDITIONAL_PRE_SCRIPTS="./scripts/upgrade/v6_to_7/pre_08_wasm.sh"
ADDITIONAL_AFTER_SCRIPTS="./scripts/upgrade/v6_to_7/post_08_wasm.sh"

SLEEP_TIME=1

KEY="mykey"
KEY1="test1"
KEY2="test2"

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
    wget -c "https://github.com/notional-labs/composable-cosmos/archive/refs/tags/${OLD_VERSION}.zip" -O _build/${OLD_VERSION}.zip
    unzip _build/${OLD_VERSION}.zip -d _build
fi


# reinstall old binary
if [ $# -eq 1 ] && [ $1 == "--reinstall-old" ] || ! command -v _build/old/centaurid &> /dev/null; then
    cd ./_build/composable-cosmos-${OLD_VERSION}
    GOBIN="$ROOT/_build/old" go install -mod=readonly ./...
    cd ../..
fi


# install new binary
FORCE_BUILD=${FORCE_BUILD:-"false"}
if ! command -v _build/new/centaurid &> /dev/null || [[ "$FORCE_BUILD" == "true" ]]; then
    echo "Building the new binary..."
    mkdir -p _build/new
    GOBIN="$ROOT/_build/new" go install -mod=readonly ./...
fi

# run old node
# if [[ "$OSTYPE" == "darwin"* ]]; then
  
# else
#     screen node1 bash scripts/localnode.sh _build/old/centaurid $DENOM
# fi
echo "running old node"
screen -dms old-node bash scripts/localnode.sh _build/old/centaurid $DENOM

sleep 2 # wait for note to start 

# execute additional pre scripts
if [ ! -z "$ADDITIONAL_PRE_SCRIPTS" ]; then
    # slice ADDITIONAL_SCRIPTS by ,
    SCRIPTS=($(echo "$ADDITIONAL_PRE_SCRIPTS" | tr ',' ' '))
    for SCRIPT in "${SCRIPTS[@]}"; do
         # check if SCRIPT is a file
        if [ -f "$SCRIPT" ]; then
            echo "executing additional pre scripts from $SCRIPT"
            source $SCRIPT 
            sleep 2
        else
            echo "$SCRIPT is not a file"
        fi
    done
fi


run_fork () {
    echo "forking"

    while true; do
        BLOCK_HEIGHT=$(./_build/old/centaurid status | jq '.SyncInfo.latest_block_height' -r)
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
    STATUS_INFO=($(./_build/old/centaurid status --home $HOME | jq -r '.NodeInfo.network,.SyncInfo.latest_block_height'))
    UPGRADE_HEIGHT=$((STATUS_INFO[1] + 18))
    echo "UPGRADE_HEIGHT = $UPGRADE_HEIGHT"

    tar -cf ./_build/new/centaurid.tar -C ./_build/new centaurid
    SUM=$(shasum -a 256 ./_build/new/centaurid.tar | cut -d ' ' -f1)
    UPGRADE_INFO=$(jq -n '
    {
        "binaries": {
            "linux/amd64": "file://'$(pwd)'/_build/new/centaurid.tar?checksum=sha256:'"$SUM"'",
        }
    }')


    ./_build/old/centaurid tx gov submit-legacy-proposal software-upgrade "$SOFTWARE_UPGRADE_NAME" --upgrade-height $UPGRADE_HEIGHT --upgrade-info "$UPGRADE_INFO" --title "upgrade" --description "upgrade"  --from $KEY --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    ./_build/old/centaurid tx gov deposit 1 "20000000${DENOM}" --from $KEY --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    ./_build/old/centaurid tx gov vote 1 yes --from $KEY --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    # ./_build/old/centaurid tx gov vote 1 yes --from $KEY2 --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    # sleep $SLEEP_TIME

    # determine block_height to halt
    while true; do
        BLOCK_HEIGHT=$(./_build/old/centaurid status | jq '.SyncInfo.latest_block_height' -r)
        if [ $BLOCK_HEIGHT = "$UPGRADE_HEIGHT" ]; then
            # assuming running only 1 centaurid
            echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, KILLING OLD ONE"
            pkill centaurid
            break
        else
            ./_build/old/centaurid q gov proposal 1 --output=json | jq ".status"
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
CONTINUE="true" screen -dms new-node bash scripts/localnode.sh _build/new/centaurid $DENOM 

sleep 5


# execute additional after scripts
if [ ! -z "$ADDITIONAL_AFTER_SCRIPTS" ]; then
    # slice ADDITIONAL_SCRIPTS by ,
    SCRIPTS=($(echo "$ADDITIONAL_AFTER_SCRIPTS" | tr ',' ' '))
    for SCRIPT in "${SCRIPTS[@]}"; do
         # check if SCRIPT is a file
        if [ -f "$SCRIPT" ]; then
            echo "executing additional after scripts from $SCRIPT"
            source $SCRIPT _build/new/centaurid
            sleep 5
        else
            echo "$SCRIPT is not a file"
        fi
    done
fi