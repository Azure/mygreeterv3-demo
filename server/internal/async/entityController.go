package async

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Azure/aks-async/database"
	opbus "github.com/Azure/aks-async/operationsbus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
)

type EntityFactoryFunc func(string) opbus.Entity

var _ opbus.EntityController = &EntityController{}

type EntityController struct {
	dbClient        *sql.DB
	entityTableName string
	matcher         *opbus.Matcher
}

func NewEntityController(ctx context.Context, options Options, matcher *opbus.Matcher, dbClient *sql.DB) (*EntityController, error) {
	logger := ctxlogger.GetLogger(ctx)

	if options.EntityTableName == "" {
		logger.Error("No EntityTableName provided.")
		return nil, errors.New("No EntityTableName provided.")
	}

	if matcher == nil {
		logger.Error("No matcher provided.")
		return nil, errors.New("No matcher provided.")
	}

	if dbClient == nil {
		logger.Error("No dbClient provided.")
		return nil, errors.New("No dbClient provided.")
	}

	newEntityController := &EntityController{
		dbClient:        dbClient,
		entityTableName: options.EntityTableName,
		matcher:         matcher,
	}

	return newEntityController, nil
}

func (e *EntityController) GetEntity(ctx context.Context, opReq opbus.OperationRequest) (opbus.Entity, error) {
	logger := ctxlogger.GetLogger(ctx)
	logger.Info("Getting entity with id: " + opReq.EntityId)

	queryEntity := fmt.Sprintf("SELECT last_operation_id FROM %s WHERE entity_id = @p1", e.entityTableName)
	rows, err := database.QueryDb(ctx, e.dbClient, queryEntity, opReq.EntityId)
	if err != nil {
		logger.Error("Error executing query: " + err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var lastOperationId string
	if rows.Next() {
		err := rows.Scan(&lastOperationId)
		if err != nil {
			logger.Info("Error scanning row: " + err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		logger.Error("No rows returned for entityId: " + opReq.EntityId)
		return nil, status.Error(codes.NotFound, "EntityId not found in database.")
	}

	entity, err := e.matcher.CreateEntityInstance(opReq.OperationName, lastOperationId)
	if err != nil {
		logger.Error("Something went wrong creating the entity instance: " + err.Error())
		return nil, err
	}

	return entity, nil
}
