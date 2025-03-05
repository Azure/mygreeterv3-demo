#!/bin/bash

# Define the namespaces
NAMESPACES=("servicehubval-mygreeterv3-server" "servicehubval-mygreeterv3-client" "servicehubval-mygreeterv3-demoserver" "servicehubval-mygreeterv3-async")

# define wait time before exit
POD_WAIT_TIME="120s"
NS_WAIT_TIME="5s"
ENV_CONFIG_PATH=$(dirname -- $(dirname -- $(dirname -- $(pwd))))/env-config.yaml
if ! command -v yq &> /dev/null; then
    echo "ERROR: yq command not found. Please install yq to proceed."
    exit 1
fi
IMAGE_TAG=$(yq -r '.serviceImageTag' $ENV_CONFIG_PATH)

echo "Checking pod status..."

# Loop through each namespace
for NAMESPACE in "${NAMESPACES[@]}"
do
    echo "Checking namespace: $NAMESPACE"
    kubectl wait --for jsonpath='{.status.phase}=Active' --timeout="$NS_WAIT_TIME" namespace/$NAMESPACE > /dev/null
    if [ $? -ne 0 ]
    then
        echo "ERROR: $NAMESPACE does not exist."   
        exit 1
    fi
    PODS=$(kubectl get pods -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}')
    if [ -z "$PODS" ]; then
      echo "ERROR: No pods are in this namespace."
      exit 1
    fi
    kubectl wait --for=jsonpath='{.status.phase}'=Running pods --all --namespace $NAMESPACE --timeout=$POD_WAIT_TIME > /dev/null
    if [ $? -ne 0 ]
    then
        echo "ERROR: $NAMESPACE pods did not run successfully."   
        exit 1
    fi
    kubectl wait --for=condition=ready pods --all --namespace $NAMESPACE --timeout=$POD_WAIT_TIME > /dev/null
    if [ $? -ne 0 ]
    then
        echo "ERROR: $NAMESPACE pods are not ready."   
        exit 1
    fi
    echo "Pods in $NAMESPACE are running."
    echo "Checking image tags in: $NAMESPACE"
    IMAGES=$(kubectl get pods -n $NAMESPACE -o jsonpath='{range .items[*]}{.spec.containers[*].image}{"\n"}{end}')
    for IMAGE in $IMAGES; do
        if [[ $IMAGE != *":"$IMAGE_TAG ]]; then
            echo "ERROR: Image used for pod in $NAMESPACE is $IMAGE and is not using the expected tag $IMAGE_TAG."
            exit 1
        fi
    done
    echo "All images in $NAMESPACE are using the expected tag $IMAGE_TAG."
done
echo "All pods are running."