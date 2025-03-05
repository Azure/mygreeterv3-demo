#!/bin/bash

# Check if the correct number of arguments is provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <input_json_file> <output_yaml_file>"
    exit 1
fi

# Check if the input JSON file exists
if [ ! -e "$1" ]; then
    echo "Error: Input JSON file '$1' does not exist."
    exit 1
fi

# Read the clientId value from the JSON file
CLIENT_ID=$(jq -r '.clientId.value' "$1")

# Generate the YAML configuration directly
YAML_CONFIG="serviceAccount:
  annotations:
    azure.workload.identity/client-id: \"$CLIENT_ID\""

# Write the YAML configuration to the specified output file
echo "$YAML_CONFIG" > "$2"

echo "$2 generated successfully!"
