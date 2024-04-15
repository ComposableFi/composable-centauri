
ROOT=$(pwd)

cd $ROOT/_build/composable/

# init clients
nix run  .#picasso-centauri-ibc-init
sleep 1 

# init connection 
nix run  .#picasso-centauri-ibc-connection-init
sleep 1

# init channel 
nix run  .#picasso-centauri-ibc-channels-init
sleep 1
 
 # run relayer 
nix run  .#picasso-centauri-ibc-relay
sleep 1