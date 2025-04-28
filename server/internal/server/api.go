package server

import (
	"context"
	"fmt"
	log "log/slog"
	"net/http"
	"os"

	"regexp"

	"database/sql"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/logattrs"
	oc "github.com/Azure/OperationContainer/api/v1"
	ocClient "github.com/Azure/OperationContainer/api/v1/client"
	database "github.com/Azure/aks-async/database"
	"github.com/Azure/aks-async/servicebus"
	"github.com/Azure/aks-middleware/grpc/interceptor"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"google.golang.org/grpc"

	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1/client"
	serviceHubPolicy "github.com/Azure/aks-middleware/http/client/azuresdk/policy"
)

type ServerInfo struct {
	GrpcServer *grpc.Server
	GwServer   *http.Server
	GrpcPort   int
	HttpPort   int
}

type Server struct {
	// When the UnimplementedMyGreeterServer struct is embedded,
	// the generated method/implementation in .pb file will be associated with this struct.
	// If this struct doesn't implment some methods,
	// the .pb ones will be used. If this struct implement the methods, it will override the .pb ones.
	// The reason is that anonymous field's methods are promoted to the struct.
	//
	// When this struct is NOT embedded,, all methods have to be implemented to meet the interface requirement.
	// See https://go.dev/ref/spec#Struct_types.
	pb.UnimplementedMyGreeterServer
	ResourceGroupClient      *armresources.ResourceGroupsClient
	AccountsClient           *armstorage.AccountsClient
	client                   pb.MyGreeterClient
	secondaryClient          pb.MyGreeterClient // P32e4
	operationContainerClient oc.OperationContainerClient
	serviceBusClient         servicebus.ServiceBusClientInterface
	serviceBusSender         servicebus.SenderInterface
	dbClient                 *sql.DB
	entityTableName          string
	ServerInfo
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Init(options Options) {
	var err error
	var cred azcore.TokenCredential

	logger := log.New(log.NewTextHandler(os.Stdout, nil).WithAttrs(logattrs.GetAttrs()))
	if options.JsonLog {
		logger = log.New(log.NewJSONHandler(os.Stdout, nil).WithAttrs(logattrs.GetAttrs()))
	}

	log.SetDefault(logger)
	if options.EnableAzureSDKCalls {
		armClientOptions := serviceHubPolicy.GetDefaultArmClientOptions(logger)
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
			os.Exit(1)
		}
		resourcesClientFactory, err := armresources.NewClientFactory(options.SubscriptionID, cred, armClientOptions)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}

		s.ResourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()
		s.AccountsClient, err = armstorage.NewAccountsClient(options.SubscriptionID, cred, armClientOptions)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	}

	if options.RemoteAddr != "" {
		s.client, err = client.NewClient(options.RemoteAddr, interceptor.GetClientInterceptorLogOptions(logger, logattrs.GetAttrs()))
		// logging the error for transparency, retry interceptor will handle it
		if err != nil {
			log.Error("did not connect: " + err.Error())
		}
	}

	if options.SecondaryRemoteAddr != "" { // P32e4
		s.secondaryClient, err = client.NewClient(options.SecondaryRemoteAddr, interceptor.GetClientInterceptorLogOptions(logger, logattrs.GetAttrs())) // P32e4
		if err != nil { // P837b
			log.Error("did not connect to secondary server: " + err.Error()) // P837b
		} // P837b
	} // P32e4

	if options.ServiceBusHostName != "" {
		s.serviceBusClient, err = servicebus.CreateServiceBusClient(context.Background(), options.ServiceBusHostName, nil, nil)
		if err != nil {
			logger.Error("Something went wrong creating the service bus client: " + err.Error())
			os.Exit(1)
		}
	}

	if options.ServiceBusQueueName != "" {
		s.serviceBusSender, err = s.serviceBusClient.NewServiceBusSender(context.Background(), options.ServiceBusQueueName, nil)
		if err != nil {
			logger.Error("Something went wrong creating the service bus sender: " + err.Error())
			os.Exit(1)
		}
	}

	if options.OperationContainerAddr != "" {
		s.operationContainerClient, err = ocClient.NewClient(options.OperationContainerAddr, interceptor.GetClientInterceptorLogOptions(logger, logattrs.GetAttrs()))
		if err != nil {
			logger.Error("Failed to initialize operationContainerClient at " + options.OperationContainerAddr + "with err: " + err.Error())
			os.Exit(1)
		}
	}

	dbStarted := false
	if options.DatabaseConnectionString != "" {
		s.dbClient, err = database.NewDbClientWithConnectionString(context.Background(), options.DatabaseConnectionString)
		if err != nil {
			logger.Error("Error creating connection pool: " + err.Error())
			os.Exit(1)
		}
		dbStarted = true
	} else if options.DatabaseServerUrl != "" && options.DatabaseName != "" {
		s.dbClient, err = database.NewDbClient(context.Background(), options.DatabaseServerUrl, options.DatabasePort, options.DatabaseName)
		if err != nil {
			logger.Error("Error creating connection pool: " + err.Error())
			os.Exit(1)
		}
		dbStarted = true
	}

	if dbStarted {

		if options.EntityTableName == "" {
			logger.Error("No OperationTableName set.")
			os.Exit(1)
		}

		if err := sanitizeTableName(options.EntityTableName); err != nil {
			logger.Error("Table name is not valid: " + err.Error())
			os.Exit(1)
		}

		s.entityTableName = options.EntityTableName
		//TODO(mheberling): Move this common functionality to a public repo which can also house other util functions we use.
		// Check if table exists
		entityListCheckQuery := "SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = @p1"
		rows, err := database.QueryDb(context.Background(), s.dbClient, entityListCheckQuery, options.EntityTableName)
		if err != nil {
			logger.Error("Error querying for entity table metadata: " + err.Error())
			os.Exit(1)
		}
		defer rows.Close()

		var TABLE_CATALOG, TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE string
		for rows.Next() {
			err = rows.Scan(&TABLE_CATALOG, &TABLE_SCHEMA, &TABLE_NAME, &TABLE_TYPE)
			if err != nil {
				logger.Error("Error getting the operationStatus of the current Entity: " + err.Error())
				os.Exit(1)
			}
		}

		if TABLE_CATALOG == "" && TABLE_SCHEMA == "" && TABLE_NAME == "" && TABLE_TYPE == "" {
			logger.Info(fmt.Sprintf("The table %s doesn't exist!", s.entityTableName))
			entityListCreateTableQuery := fmt.Sprintf("CREATE TABLE %s (entity_type VARCHAR(255), entity_id VARCHAR(255), last_operation_id VARCHAR(255), operation_name VARCHAR (255), operation_status VARCHAR(255))", s.entityTableName)
			_, err = database.QueryDb(context.Background(), s.dbClient, entityListCreateTableQuery)
			if err != nil {
				logger.Error("Error creating the entity table: " + err.Error())
				os.Exit(1)
			}
			logger.Info(fmt.Sprintf("The table %s has been created.", options.EntityTableName))
		}
	}
}

func sanitizeTableName(tableName string) error {
	// Use a regular expression to allow only alphanumeric characters and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validName.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	return nil
}
