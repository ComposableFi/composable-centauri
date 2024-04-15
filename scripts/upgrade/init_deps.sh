#!/bin/bash


OLD_VERSION=kien-devnet-651
SOFTWARE_UPGRADE_NAME="v6_6_0"
ROOT=$(pwd)

COMPOSABLE_VERSION="branchfortestingpfmfix"

mkdir -p _build/


# Check if the directory exists and is a Git repository
# TODO: using git, since nix in composable repo requires something with git
# Consider using submodule, or firgure this out
if [ ! -d "_build/composable/.git" ]; then
    cd _build/
    git clone https://github.com/ComposableFi/composable.git composable
    cd composable
    git checkout "$COMPOSABLE_VERSION"
    cd ../.././
fi


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
if ! command -v _build/new/picad &> /dev/null
then
    mkdir -p _build/new
    GOBIN="$ROOT/_build/new" go install -mod=readonly ./...
fi

