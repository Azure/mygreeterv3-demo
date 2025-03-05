#!/bin/bash -x

# The api version is fixed based on the value of the SERVICEHUB_APIV1_VERSION variable.
# It must be specified in double quotes
# The automated package versioning logic bumps the PATCH version only
SERVICEHUB_APIV1_VERSION="0.0.46"
SERVICEHUB_AKSMIDDLEWARE_VERSION="0.0.31"

# Execute this file under the service directory.

# `go work use`` must be after the module's go.mod exists.
# Once the go.work file exist, the module where you want to run `go build ./...`
# must be in go.work's use list.

# For the api.
cd api
go mod init dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api
go mod edit -require github.com/Azure/aks-middleware@v$SERVICEHUB_AKSMIDDLEWARE_VERSION
go get google.golang.org/genproto@latest
cd ..
cd api
cd v1 
mkdir mock # Current workaround until we find a cleaner way to add the directory to the template files.
make service
if [ $? -ne 0 ]
then
    echo "Make service failed."
    exit 1
fi
cd ..
go build ./...
if [ $? -ne 0 ]
then
    echo "Building the api module failed."
    exit 1
fi
go test ./...
if [ $? -ne 0 ]
then
    echo "Testing the api module failed."
    exit 1
fi
gofmt -s -w .
cd ..

cat << EOM

If your goModuleNamePrefix has . (dot) inside, you have to create the module
in the repo. Otherwise Go will complain that the module cannot be found
even if you use go.work.

Use the following commands:

git add ./api
git commit -m "api module v0.0.1"
git tag mygreeterv3/api/v0.0.1
git push
git push origin mygreeterv3/api/v0.0.1

Then you come back here to run this script again.
After git push, the "module cannot be found" message may still persist.
Wait a couple of minutes and the git repo will be able to return the module.

EOM

echo -----------------------

# For the server
cd server
go mod init dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server
go mod edit -require dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api@v$SERVICEHUB_APIV1_VERSION
go mod edit -require github.com/Azure/aks-middleware@v$SERVICEHUB_AKSMIDDLEWARE_VERSION
go mod edit -require github.com/Azure/aks-async@v0.0.39
go mod edit -require github.com/Azure/OperationContainer@v0.0.11
go get google.golang.org/genproto@latest
go mod tidy
# The following command must be run AFTER go mod tidy. If ran before, building the server module
# will fail as go mod tidy removes the indirect dependency with google.golang.org/genproto
# and go work sum will pull in an older version that causes an ambiguous import error.
# For more information refer to: https://github.com/googleapis/go-genproto/issues/1015
go get google.golang.org/genproto@latest
cd ..

go work init
go work use ./api
go work use ./server

cd server
go build ./...
if [ $? -ne 0 ]
then
    echo "Building the server module failed."
    echo "If downloading the server module failed, you might have to wait for the api module to be available or the tag to settle then rerun again"
    exit 1
fi
go test -tags=testcontainers ./...
if [ $? -ne 0 ]
then
    echo "Testing the server module failed."
    exit 1
fi
gofmt -s -w .
cd ..

