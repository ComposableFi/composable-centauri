ADDITIONAL_SCRIPTS=(
    "./scripts/upgrade/v_6_4_8/setup-08-wasm.sh"
)

for SCRIPT in "${ADDITIONAL_SCRIPTS[@]}"; do
    if [ -f "$SCRIPT" ]; then
        echo "Running additional script: $SCRIPT"
        source $SCRIPT
    else
        echo "Additional script $SCRIPT does not exist."
    fi
done

