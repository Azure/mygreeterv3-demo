// Auto generated. Don't modify.
package async

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	log "log/slog"
	"regexp"

	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/async/operations"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/async/operations/longRunningOperation"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/logattrs"
	"github.com/Azure/aks-middleware/grpc/interceptor"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	_ "github.com/microsoft/go-mssqldb"

	oc "github.com/Azure/OperationContainer/api/v1"
	ocClient "github.com/Azure/OperationContainer/api/v1/client"
	"github.com/Azure/aks-async/database"
	opbus "github.com/Azure/aks-async/operationsbus"
	"github.com/Azure/aks-async/servicebus"
	"github.com/Azure/go-shuttle/v2"
)

type Async struct {
	Processor       *shuttle.Processor
	entityTableName string
}

// TODO(mheberling): Add unit test for async.
func NewAsync(ctx context.Context, options Options) (*Async, error) {
	var err error
	var cred azcore.TokenCredential

	logger := ctxlogger.GetLogger(ctx)
	// logger := log.New(log.NewTextHandler(os.Stdout, nil).WithAttrs(logattrs.GetAttrs()))
	// if options.JsonLog {
	// 	logger = log.New(log.NewJSONHandler(os.Stdout, nil).WithAttrs(logattrs.GetAttrs()))
	// }

	// Use MSI in Standalone E2E env for credential
	if options.IdentityResourceID != "" {
		resourceID := azidentity.ResourceID(options.IdentityResourceID)
		opts := azidentity.ManagedIdentityCredentialOptions{ID: resourceID}
		cred, err = azidentity.NewManagedIdentityCredential(&opts)
	} else {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	}

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if options.EntityTableName == "" {
		logger.Error("No EntityTableName set.")
		return nil, errors.New("No EntityTableName set.")
	}

	if err := sanitizeTableName(options.EntityTableName); err != nil {
		logger.Error("Table name is not valid: " + err.Error())
		return nil, err
	}

	var operationContainerClient oc.OperationContainerClient
	if options.OperationContainerAddr != "" {
		operationContainerClient, err = ocClient.NewClient(options.OperationContainerAddr, interceptor.GetClientInterceptorLogOptions(logger, logattrs.GetAttrs()))
		if err != nil {
			log.Error("did not connect to operationContainerClient: " + err.Error())
			return nil, err
		}
	}

	var serviceBusClient servicebus.ServiceBusClientInterface
	if options.ServiceBusHostName != "" {
		serviceBusClient, err = servicebus.CreateServiceBusClient(ctx, options.ServiceBusHostName, cred, nil)
		if err != nil {
			log.Error("Something went wrong creating the service bus client: " + err.Error())
			return nil, err
		}
	}

	var receiver servicebus.ReceiverInterface
	if options.ServiceBusQueueName != "" {
		receiver, err = serviceBusClient.NewServiceBusReceiver(ctx, options.ServiceBusQueueName, nil)
		if err != nil {
			log.Error("Something went wrong creating the service bus receiver: " + err.Error())
			return nil, err
		}
	}

	// Verify that some db information was provided
	if options.DatabaseServerUrl == "" && options.DatabaseName == "" && options.DatabaseConnectionString == "" {
		logger.Error("No database information provided.")
		return nil, errors.New("No database information provided.")
	}

	// The database is created by the bicep files and deployed in the deployment of service specific resources. The entityTableName
	// might not be created yet (since the table is created by the server and async and server should initialize simultaneously)
	// but that doesn't matter because if the entityTable hasn't been created, it means that the server hasn't started and async
	// should not be receiving any messages through the service bus to process.
	var dbClient *sql.DB
	if options.DatabaseConnectionString != "" {
		dbClient, err = database.NewDbClientWithConnectionString(ctx, options.DatabaseConnectionString)
		if err != nil {
			logger.Error("Error creating connection pool: " + err.Error())
			return nil, err
		}
	} else if options.DatabaseServerUrl != "" && options.DatabaseName != "" {
		dbClient, err = database.NewDbClient(ctx, options.DatabaseServerUrl, options.DatabasePort, options.DatabaseName)
		if err != nil {
			logger.Error("Error creating connection pool: " + err.Error())
			return nil, err
		}
	}

	// Instantiate a matcher. Here we would add all of our operation types.
	matcher := opbus.NewMatcher()
	lro := &longRunningOperation.LongRunningOperation{}
	matcher.Register(operations.LroName, lro)
	matcher.RegisterEntity(operations.LroName, longRunningOperation.CreateLroEntityFunc)

	entityController, err := NewEntityController(ctx, options, matcher, dbClient)
	if err != nil {
		log.Error("Something went wrong creating the entity controller: " + err.Error())
		return nil, err
	}

	operationStatusHook := &OperationStatusHook{
		dbClient:        dbClient,
		EntityTableName: options.EntityTableName,
	}
	hooks := []opbus.BaseOperationHooksInterface{operationStatusHook}

	// Add hooks from hooks.go
	processor, err := opbus.CreateProcessor(receiver, matcher, operationContainerClient, entityController, nil, nil, nil, hooks)
	if err != nil {
		return nil, err
	}

	async := &Async{
		Processor: processor,
	}

	return async, nil
}

func sanitizeTableName(tableName string) error {
	// Use a regular expression to allow only alphanumeric characters and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validName.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	return nil
}
