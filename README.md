# mygreeterv3

## Prerequisites

- Follow the instructions in [../README/README.md](../README/README.md) to setup the generated service and make sure it can run. It involves creating developer's own environment that is isolated from other developers, provisioning the shared resources, provisioning the service specific resources, and initializing the service.

## Code Overview

This directory stores a complete microservice. It has two Go modules.

### api

This module stores the microservice's API definition.

- The API is defined through [protobuf](https://grpc.io/docs/languages/go/quickstart/). The gRPC method is annotated to provide the following features.
  - option (google.api.http) and option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation). By declaring the HTTP mapping and OpenAPI documentation, the [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) module can generate a reverse proxy that expose the gRPC microservice as an HTTP service. It also generate a swagger.json OpenAPI document. Through the [swagger-codegen-cli](https://swagger.io/docs/open-source-tools/swagger-codegen/), the Makefile further generates an HTTP client library (restsdk) based on the swagger.json.
  - buf.validate. By declaring the field constraints, the [protovalidate](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/main/interceptors/protovalidate) middleware will automatically enforce the field constraints.
  - servicehub.fieldoptions.loggable. By declaring if a field is false for logging, the [aks-middleware](https://github.com/Azure/aks-middleware/blob/main/ctxlogger/ctxlogger.go) will not log this field. By default, all fields will be logged automatically.
- client. It provides a function to create a gRPC client to talk with this microservice. The client enabled the following [default interceptors](https://github.com/Azure/aks-middleware/blob/main/interceptor/interceptor.go)
  - Auto retry. Automatically retry the failed request if the error is retryable.
  - metadata forwarding. It forwards information such as request id so that we can correlate requests across microservices.
  - Request logging. Each request will be logged once. QoS dashboard uses the logs to show qps, latency, error ratio, etc.
- mock. This is auto generated. All mock functions of the API are generated so that users can use the mock directly in their unit test rather than each user implements their own mock.
- restsdk. This is auto generated. Users can use this SDK to access the HTTP API instead of making raw HTTP requests directly.

### server

This module stores the implementation of the microservice.

- source code. The Go source code are in cmd and internal. They follow the guideline in [Go project layout](https://github.com/golang-standards/project-layout). Two binaries will be built out of the source code. See [Run Service Locally](#run-service-locally) for details.
  - client. The client binary demonstrated how to use the microservice's client library (gRPC client and restsdk) to call the gRPC service and HTTP service.
  - server. The server binary is the key microservice. It demonstrates how a full-blown microservice works.
    - Accepting gRPC calls and HTTP/REST calls.
    - Making calls to its dependency to fulfill incoming calls. The dependency can be a gRPC service, an HTTP service such as Azure, or something else. When it is Azure, it demonstrates how the code can assume an Azure identity and gain access to Azure.
    - The server is configured with both server interceptors (for incoming calls) and client interceptors (for outgoing calls). See details at [default interceptors](https://github.com/Azure/aks-middleware/blob/main/interceptor/interceptor.go). The buf.validate and servicehub.fieldoptions.loggable annotations in the protobuf need to work with the interceptors to be effective.
  - async. The async binary processes asynchronous operations (which are typically long-running) by using a processor with handlers provided by the [aks-async](https://github.com/Azure/aks-async) library.
    - This component does not receive gRPC calls, rather grabs operations directly from a connected Azure Service Bus resource in order to process them accordingly.
    - Async also utilizes an Azure SQL Server created by the service specific resources earlier, and it uses the url or connection string with the name of the specific database to connect to it.
      - The database is created by the bicep files and deployed in the deployment of service specific resources. The entityTableName might not be created yet (since the table is created by the server and async and server should initialize simultaneously) but that doesn't matter because if the entityTable hasn't been created, it means that the server hasn't started and async should not be receiving any messages through the service bus to process.
- deployments. The deployments are via [Helm](https://helm.sh/).
  - The two binaries are deployed as k8s deployments. The server is exposed as a k8s service (ClusterIP).
  - To grant Azure managed identity to the server microservice, [AKS workload identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster) is used. It involves multiple components.
    - Shared resource: The AKS cluster needs to enable this feature.
    - Service resource: Managed identity needs to trust the AKS cluster as an OIDC issuer. The managed identity output its client ID.
    - K8s service account: Annotate with the managed identity's client id so that the service account will assume the managed identity when talking with Azure.
    - K8s pod spec: Run with the identity of the above k8s service account.
    - Shared resource deployment and service resource deployment are handled separately. Helm deployment only handles k8s resources.
  - [Istio](https://istio.io/latest/) is used to enforce mutual TLS (PeerAuthentication) and microservice method level access control (AuthorizationPolicy).
- resources. This directory stores service resources only.
  - The managed identity and its role assignment. The managed identity is used by the microservice.
  - Each method's QoS (throughput, error ratio, latency) alert. They are logAnalyticsWorkspace scheduled-query-rule.
  - For shared resources and resources in general, see [../README/README.md](../README/README.md)
- test. This is the integration test. Multiple scripts are defined to finish the build, release, test cycle.
- monitoring. This folder stores the QoS dashboard. Please follow the README.md in the directory to create your dashboard. The dashboard will only show data from your instance defined in env-config.yaml.
- Ev2. See [../README/Ev2_README.md](../README/Ev2_README.md)

## Middleware

The service leverages multiple middleware for features such as logging, retry, and input validation. To learn more, please visit the [middleware repo](https://github.com/Azure/aks-middleware/tree/main).

## Modify the API

Whenever the API is changed, you need to run the following command to regenerate the code.

```bash
cd api/v1
make service
```

## Run Service Locally

Deploying the changed service to Azure such as an AKS cluster for a complete modify-test cycle is slow. You can run the service on your local machine to speed up the development cycle.

### Server

```bash
go run dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/cmd/server start
```

