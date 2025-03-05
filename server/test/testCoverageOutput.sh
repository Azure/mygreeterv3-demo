#!/bin/bash -x
#Define color codes for printing to help analysis.
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
mkdir -p mygreeterv3/server/test/coverage_reports

echo "export GOPRIVATE='dev.azure.com'" >> ~/.bashrc
git config --global url."https://$READPAT@dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service".insteadOf "https://dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service"


go install github.com/onsi/ginkgo/v2/ginkgo@latest
#Add go path to bash
PATH=$PATH:$(go env GOPATH)/bin
cd mygreeterv3
fileName=mygreeterv3-coverage-report
#Perform test coverage for given folder and save output result
ginkgo -r -v --trace --coverprofile=.coverage-report.out --skip-package=mock ./...
go tool cover -html=.coverage-report.out -o server/test/coverage_reports/$fileName.html
#Extract coverage percentage as an integer and compare to required threshold.
results=$(go tool cover -func=.coverage-report.out | grep total: | awk '{print $NF}')
number="${results%\%}"
number=$(printf "%.0f" $number)
if [ $number -lt $threshold ]
then
    echo -e "${RED}$fileName results were: $results and below the required threshold of: $threshold${NC}"
    exit 1
fi
echo -e "${GREEN}$fileName results were: $results and above the required threshold of: $threshold${NC}"