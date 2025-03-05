### How the build and push Makefile works
This makefile is templated similar to the other files as it requires (.subscriptionId) and the unique id that is used to create the resourcs (.resourcesName). 

The docker image is built with the following arguments, depending on whether or not we are accessing a GoModuleNamePrefix specific to a user's repository or a MSFT internal module. Since the dockerfile will run go mod tidy, it will try to download the ./api module from the goModuleNamePrefix that was established, and we need to provide the correct vars to access it.

For user's repositories. (i.e accessing dev.azure.com)
- The build requires the environment variables READPAT (the personal authentication token associated with the user's repository that has read access.)

For aks hosted modules. (i.e accessing go.goms.io)
- The build requires the environment variables AKS_GOPROXY_TOKEN, GOPROXY, GOPRIVATE, GONOPROXY (the vars associated with setting up goproxy for go.goms.io)
