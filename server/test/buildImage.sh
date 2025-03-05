#!/bin/bash
#Define color codes for printing to help analysis.
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
#---------
cd mygreeterv3
cd server

export READPAT=$READPAT

if [ "$WORKSPACE" = "true" ]
then
    make build-workspace-image
    if [ $? -ne 0 ]
    then
        echo -e "${RED}Docker image build failed with exit code $?${NC}"
        exit 1
    fi
else
    make build-image
    if [ $? -ne 0 ]
    then
        echo -e "${RED}Docker image build failed with exit code $?${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}Docker image build was successfull!${NC}"