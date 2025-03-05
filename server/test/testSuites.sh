#!/bin/bash -x
go install github.com/onsi/ginkgo/v2/ginkgo@latest
#Add go path to bash
PATH=$PATH:$(go env GOPATH)/bin
# Run this file to create test suites for any folders under this repo that have go files but no test suites
shopt -s globstar
currentDir=$(pwd)
echo $currentDir
for d in **;
do
    if [ -d "$d" ] &&  [[ $d != *"mock"* ]];
    then
        subD=$(find "$d" -maxdepth 1 -type d | wc -l)
        subD=$((subD - 1))
        if [ $subD -eq 0 ]; then
            cd $d
            countGo=$(ls *.go 2>/dev/null | wc -l)
            if [ $countGo -ne 0 ];
            then
                suites=$(ls *_test.go 2>/dev/null | wc -l)
                if [ $suites -eq 0 ];
                then
                    echo "There were $countGo go files in $d and no test suite."
                    echo "If running locally, this will produce a test suite in the folder. However, as a part of our PR pipeline, this test will fail causing the inability to merge until it is re-run locally or test suite is created manually."
                    ginkgo bootstrap
                    exit 1
                else
                    :
                fi
            else
                :
            fi
            cd $currentDir
        else
            :
        fi
    else
        :
    fi
done
echo "All folders with go files have their required test suites"
