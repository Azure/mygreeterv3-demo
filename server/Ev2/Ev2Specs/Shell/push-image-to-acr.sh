#!/bin/bash
# TODO: Investigate if each service needs to take care of image push by themselves.
set -e

if [ -z ${DESTINATION_ACR_NAME+x} ]; then
    echo "DESTINATION_ACR_NAME is unset, unable to continue"
    exit 1;
fi

if [ -z ${TARBALL_IMAGE_FILE_SAS+x} ]; then
    echo "TARBALL_IMAGE_FILE_SAS is unset, unable to continue"
    exit 1;
fi

if [ -z ${IMAGE_NAME+x} ]; then
    echo "IMAGE_NAME is unset, unable to continue"
    exit 1;
fi

if [ -z ${TAG_NAME+x} ]; then
    echo "TAG_NAME is unset, unable to continue"
    exit 1;
fi

if [ -z ${DESTINATION_FILE_NAME+x} ]; then
    echo "DESTINATION_FILE_NAME is unset, unable to continue"
    exit 1;
fi

echo "Folder Contents"
ls

echo "Login cli using managed identity"
az login --identity

if command -v oras &> /dev/null; then
  echo "ORAS is already installed"
else
  VERSION="1.2.0"
  wget "https://github.com/oras-project/oras/releases/download/v${VERSION}/oras_${VERSION}_linux_amd64.tar.gz" -q
  mkdir -p oras-install
  tar -zxf oras_${VERSION}_linux_amd64.tar.gz -C oras-install
  mv oras-install/oras /usr/local/bin/
  rm -rf oras_${VERSION}_linux_amd64.tar.gz oras-install
  echo "ORAS has been installed"
fi

oras version
TMP_FOLDER=$(mktemp -d)
cd $TMP_FOLDER

echo "Downloading docker tarball image from $TARBALL_IMAGE_FILE_SAS"
wget -O $DESTINATION_FILE_NAME "$TARBALL_IMAGE_FILE_SAS"

echo "Getting acr credentials"
TOKEN_QUERY_RES=$(az acr login -n "$DESTINATION_ACR_NAME" -t)
TOKEN=$(echo "$TOKEN_QUERY_RES" | jq -r '.accessToken')
DESTINATION_ACR=$(echo "$TOKEN_QUERY_RES" | jq -r '.loginServer')

echo "Successfully retrieved credentials. Destination ACR: $DESTINATION_ACR. Authenticating oras..."
oras login "$DESTINATION_ACR" -u "00000000-0000-0000-0000-000000000000" -p "$TOKEN"
echo "Oras successfully authenticated."

DEST_IMAGE_FULL_NAME="$DESTINATION_ACR_NAME.azurecr.io/$IMAGE_NAME:$TAG_NAME"

if [[ "$DESTINATION_FILE_NAME" == *".gz"* ]]; then
  gunzip $DESTINATION_FILE_NAME
  echo "$DESTINATION_FILE_NAME has been decompressed."

  DESTINATION_FILE_NAME="${DESTINATION_FILE_NAME%.gz}"
  echo "The decompressed file is: $DESTINATION_FILE_NAME"
else
  echo "$DESTINATION_FILE_NAME is not a .gz file."
fi
ls

echo "Pushing file $DESTINATION_FILE_NAME to $DEST_IMAGE_FULL_NAME"
oras cp --recursive --from-oci-layout "$DESTINATION_FILE_NAME:$TAG_NAME" $DEST_IMAGE_FULL_NAME