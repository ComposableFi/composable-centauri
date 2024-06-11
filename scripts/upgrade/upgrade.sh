# the upgrade is a fork, "true" otherwise
FORK=${FORK:-"false"}

UPGRADE_WAIT=${UPGRADE_WAIT:-20}
HOME=mytestnet
ROOT=$(pwd)
DENOM=ppica
CHAIN_ID=centauri-dev
SOFTWARE_UPGRADE_NAME="v7_0_1"
ADDITIONAL_PRE_SCRIPTS="./scripts/upgrade/v_6_6_0/pre-script.sh"
ADDITIONAL_AFTER_SCRIPTS="./scripts/upgrade/v_6_6_0/post-script.sh"
KEY="mykey"
KEY1="mykey1"

SLEEP_TIME=5


UPGRADE_PROPOSAL_ID=2
run_upgrade () {
    echo -e "\n\n=> =>start upgrading"

    # Get upgrade height, 12 block after (6s)
    STATUS_INFO=($(./_build/old/picad status --home $HOME | jq -r '.NodeInfo.network,.SyncInfo.latest_block_height'))
    UPGRADE_HEIGHT=$((STATUS_INFO[1] + 8))
    echo "UPGRADE_HEIGHT = $UPGRADE_HEIGHT"

    tar -cf ./_build/new/picad.tar -C ./_build/new picad
    SUM=$(shasum -a 256 ./_build/new/picad.tar | cut -d ' ' -f1)
    UPGRADE_INFO=$(jq -n '
    {
        "binaries": {
            "linux/amd64": "file://'$(pwd)'/_build/new/picad.tar?checksum=sha256:'"$SUM"'",
        }
    }')


    ./_build/old/picad tx gov submit-legacy-proposal software-upgrade "$SOFTWARE_UPGRADE_NAME" --upgrade-height $UPGRADE_HEIGHT --upgrade-info "$UPGRADE_INFO" --title "upgrade" --description "upgrade"  --from $KEY --fees 100000${DENOM} --keyring-backend test --chain-id $CHAIN_ID --home $HOME -y > /dev/null

    sleep $SLEEP_TIME

    ./_build/old/picad tx gov deposit $UPGRADE_PROPOSAL_ID "20000000${DENOM}" --from $KEY --keyring-backend test --fees 100000${DENOM} --chain-id $CHAIN_ID --home $HOME -y

    sleep $SLEEP_TIME

    ./_build/old/picad tx gov vote $UPGRADE_PROPOSAL_ID yes --from $KEY --keyring-backend test --fees 100000${DENOM} --chain-id $CHAIN_ID --home $HOME -y

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
            ./_build/old/picad q gov proposal $UPGRADE_PROPOSAL_ID --output=json | jq ".status"
            echo "BLOCK_HEIGHT = $BLOCK_HEIGHT"
            sleep $SLEEP_TIME
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
CONTINUE="true" screen -L -dmS picad bash scripts/localnode.sh _build/new/picad $DENOM
#CONTINUE="true" bash scripts/localnode.sh _build/new/picad $DENOM

