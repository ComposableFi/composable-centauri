#!/bin/bash

echo "kill composable node"
pkill -f picad
rm -rf mytestnet


echo "Cleanign up devnet-picasso containers..."
# The image name you want to stop containers for
IMAGE_NAME="composablefi/devnet-picasso"

# Find the container ID(s) for containers running this image
CONTAINER_IDS=$(docker ps --filter "ancestor=$IMAGE_NAME" --format "{{.ID}}")

# Stop the container(s)
for ID in $CONTAINER_IDS; do
    echo "Stopping container $ID..."
    docker stop $ID
done

echo "Cleanup complete."