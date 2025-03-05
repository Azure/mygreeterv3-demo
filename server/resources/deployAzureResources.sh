#!/bin/bash

# ARGS
TEMPL_DIR=$1 # Set the path to your Bicep template files
SAVE_OUTPUTS=$2 # Specifies whether or not we want to save the outputs

set -e
declare -a pids=()

# Get a list of all .json files in the specified directory
TEMPL_FILES=("$TEMPL_DIR"/*.bicep)

# Function to deploy a bicep template
deploy_template() {
    local TEMPL_FILE="$1"
    # Base filename is extracted and everything after the first period is removed
    # i.e. alert-rules/CreateResourceGroup.ServiceResources.Template.bicep -> CreateResourceGroup
    BASE_FILENAME="$(basename "${TEMPL_FILE}" | cut -d. -f1)"
    RESOURCES_NAME=$(jq -r '.parameters.resourcesName.value' ServiceResources.Parameters.json); \
    az deployment sub create --name "${BASE_FILENAME}-${RESOURCES_NAME}-mygreeterv3-deploy" --location eastus --template-file "$TEMPL_FILE" --parameters ServiceResources.Parameters.json -o json > $TEMPL_DIR/.${BASE_FILENAME}_tmp.json
    if $SAVE_OUTPUTS; then \
    if cat $TEMPL_DIR/.${BASE_FILENAME}_tmp.json | grep '"provisioningState": "Succeeded"' > /dev/null 2>&1; then \
		echo "${BASE_FILENAME} resource provisioning succeeded."; \
		cp $TEMPL_DIR/.${BASE_FILENAME}_tmp.json $TEMPL_DIR/.${BASE_FILENAME}_output.json; \
        if $SAVE_OUTPUTS; then \
            jq '.properties.outputs' $TEMPL_DIR/.${BASE_FILENAME}_output.json > $TEMPL_DIR/.${BASE_FILENAME}_properties_outputs.json; \
            if [ -e "$TEMPL_DIR/saveOutputs.sh" ]; then \
                ./$TEMPL_DIR/saveOutputs.sh $TEMPL_DIR/.${BASE_FILENAME}_properties_outputs.json ../artifacts/.${BASE_FILENAME}_properties_outputs.yaml; \
            fi; \
        fi; \
	else \
		echo "$TEMPL_DIR/${BASE_FILENAME} resource provisioning did not succeed."; \
        exit 1; \
	fi
	rm $TEMPL_DIR/.${BASE_FILENAME}_tmp.json
    fi
}

# Deploy templates without waiting for previous template
for TEMPL_FILE in "${TEMPL_FILES[@]}"; do
    echo $TEMPL_FILE
    deploy_template "$TEMPL_FILE" &
    pids+=($!)
done

# Wait for all background processes to finish
wait "${pids[@]}"

# Check exit status of each process
for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
        echo "Error: Process with PID $pid failed."
        exit 1
    fi
done

echo "${TEMPL_DIR} resource provisioning has completed."