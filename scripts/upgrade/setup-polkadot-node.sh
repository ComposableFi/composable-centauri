ROOT=$(pwd)

cd $ROOT/_build/composable

# This start the node
nix run .#zombienet-rococo-local-picasso-dev
