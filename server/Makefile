
REGISTRY_NAME=servicehubregistry
IMG_NAME=service_hub_environment
IMG_DATE_TAG=20250228
MAIN_DIR := $(shell dirname -- $(shell dirname -- $(shell pwd)))
USER_ID := $(shell id -u):$(shell id -g)
DIR_NAME=mygreeterv3
#The config file used to populate resourceName and subscriptionId, default assumes there is a config file
#in the main directory of project.
ENVCONFIG?="env-config.yaml"

all: tidy test build

template-files:
	@if [ -f ../../$(ENVCONFIG) ]; then \
		docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) \
		./maintain_service executeTemplatesInFolder --templatePrefix "template-" --templatePath binded-data/$(DIR_NAME) --envConfig binded-data/$(ENVCONFIG) --leftDelim "{{" --rightDelim "}}"; \
		echo "Templating server files was successful"; \
	else \
		echo "../../$(ENVCONFIG) does not exist. Please make sure either the path exists or rerun the command with the flag --ENVCONFIG 'pathtoconfigfile'"; \
		exit 1; \
	fi

# (Prerequisite) To deploy the shared resources needed, go to the directory of the shared resources and run the Makefile.
# Creates all the service specific resources and generates markdown files
# TODO: Investigate how to express dependency across different directories rather than write imperative code.
deploy-resources: template-files
	cd resources; make all
	docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) ./maintain_service generateResourceFiles --svcDirName=binded-data/$(DIR_NAME); \

tidy: template-files
	go mod tidy

test:
	go test ./...

build: template-files
	go build -o bin/ dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/cmd/client
	go build -o bin/ dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/cmd/demoserver
	go build -o bin/ dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/cmd/server
	go build -o bin/ dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/cmd/async
	
build-image: template-files
	cd generated; make build-image

build-multiarch-image: template-files
	cd generated; make build-multiarch-image

build-workspace-image: template-files
	cd generated; make build-workspace-image

push-image: template-files
	cd generated; make push-image


AZURESDK_PATH := artifacts/.Azuresdk_properties_outputs.yaml

# The install and uninstall information (k8s secret) is stored in the current namespace.
install: template-files
	cd generated; make connect-cluster
	if [ -f $(AZURESDK_PATH) ]; then \
		docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) /bin/bash -c \
		"cd binded-data/$(DIR_NAME)/server; export KUBECONFIG=\$$(pwd)/generated/.kube/config; \
		cd deployments; helm install servicehubval-mygreeterv3-client . -f values-common.yaml -f values-client.yaml; \
		helm install servicehubval-mygreeterv3-server . -f values-common.yaml -f values-server.yaml -f ../$(AZURESDK_PATH); \
		helm install servicehubval-mygreeterv3-async . -f values-common.yaml -f values-async.yaml -f ../$(AZURESDK_PATH); \
		helm install servicehubval-mygreeterv3-demoserver . -f values-common.yaml -f values-demoserver.yaml;"; \
	else \
		echo ".Azuresdk_properties_outputs.yaml file does not exist. Run the make deploy-resources cmd"; \
		docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) /bin/bash -c \
		"cd binded-data/$(DIR_NAME)/server; export KUBECONFIG=\$$(pwd)/generated/.kube/config; \
		cd deployments; helm install servicehubval-mygreeterv3-client . -f values-common.yaml -f values-client.yaml; \
		helm install servicehubval-mygreeterv3-server . -f values-common.yaml -f values-server.yaml; \
		helm install servicehubval-mygreeterv3-async . -f values-common.yaml -f values-async.yaml; \
		helm install servicehubval-mygreeterv3-demoserver . -f values-common.yaml -f values-demoserver.yaml;"; \
	fi

upgrade: template-files
	cd generated; make connect-cluster
	if [ -f $(AZURESDK_PATH) ]; then \
		docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) /bin/bash -c \
		"cd binded-data/$(DIR_NAME)/server; export KUBECONFIG=\$$(pwd)/generated/.kube/config; \
		cd deployments; helm upgrade servicehubval-mygreeterv3-client . -f values-common.yaml -f values-client.yaml; \
		helm upgrade servicehubval-mygreeterv3-server . -f values-common.yaml -f values-server.yaml -f ../$(AZURESDK_PATH); \
		helm upgrade servicehubval-mygreeterv3-async . -f values-common.yaml -f values-async.yaml -f ../$(AZURESDK_PATH); \
		helm upgrade servicehubval-mygreeterv3-demoserver . -f values-common.yaml -f values-demoserver.yaml;"; \
	else \
		echo ".Azuresdk_properties_outputs.yaml file does not exist. Run the make deploy-resources cmd"; \
		docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) /bin/bash -c \
		"cd binded-data/$(DIR_NAME)/server; export KUBECONFIG=\$$(pwd)/generated/.kube/config; \
		cd deployments; helm upgrade servicehubval-mygreeterv3-client . -f values-common.yaml -f values-client.yaml; \
		helm upgrade servicehubval-mygreeterv3-server . -f values-common.yaml -f values-server.yaml; \
		helm upgrade servicehubval-mygreeterv3-async . -f values-common.yaml -f values-async.yaml; \
		helm upgrade servicehubval-mygreeterv3-demoserver . -f values-common.yaml -f values-demoserver.yaml;"; \
	fi

uninstall: template-files
	docker run --rm -u $(USER_ID) --mount src=$(MAIN_DIR),target=/app/binded-data,type=bind $(REGISTRY_NAME).azurecr.io/$(IMG_NAME):$(IMG_DATE_TAG) /bin/bash -c \
	"cd binded-data/$(DIR_NAME)/server; export KUBECONFIG=\$$(pwd)/generated/.kube/config; \
	helm uninstall servicehubval-mygreeterv3-client; \
	helm uninstall servicehubval-mygreeterv3-demoserver; \
	helm uninstall servicehubval-mygreeterv3-async; \
	helm uninstall servicehubval-mygreeterv3-server;";

