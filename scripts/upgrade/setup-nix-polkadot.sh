ROOT=$(pwd)

cd $ROOT/_build/composable

# Set the maximum number of attempts
max_attempts=30

# Initialize the attempt counter
attempt=1

while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt of $max_attempts"
    nix run .#zombienet-rococo-local-picasso-dev
    
    # Check if the command was successful
    if [ $? -eq 0 ]; then
        echo "Command executed successfully."
        break
    else
        echo "Command failed, retrying..."
        echo "attemp: $attempt"
        ((attempt++))
    fi

    # Optional: sleep for a few seconds before retrying
    sleep 2
done

# Check if all attempts failed
if [ $attempt -gt $max_attempts ]; then
    echo "All attempts failed."
fi