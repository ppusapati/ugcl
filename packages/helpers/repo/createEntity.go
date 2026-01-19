package helpers_repo

import (
	"context"
	"time"

	"p9e.in/ugcl/packages/database/pgxpostgres/operations"
	"p9e.in/ugcl/packages/deps"
	ph "p9e.in/ugcl/packages/helpers/utils"
	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/p9log"
)

func CreateEntity[T any](ctx context.Context, deps deps.ServiceDeps,
	tableName string, entity T, fieldNames []string, conflictColumns []string) (T, error) {
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "Create Entity"))
	var zero T
	entityType := ph.GetTypeName[T]()
	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()
	// Start tracing the operation
	lg.Infof("Starting operation: %s", "Create Repo"+entityType)
	startTime := time.Now()

	dm := models.DataModel[T]{
		TableName:       tableName,
		FieldNames:      fieldNames,
		Values:          []T{entity},
		ConflictColumns: conflictColumns,
	}
	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeInsert,
	)
	// End tracing the operation
	elapsedTime := time.Since(startTime)
	lg.Infof("Completed operation: %s in %s", entityType, elapsedTime)

	// Collect metrics
	deps.Metrics.RecordDBOperation(entityType, elapsedTime, true)
	if err != nil {
		// Log timeout or cancellation
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("CreateUser operation timed out")
		}
		lg.Errorf("failed to create user: %v", err)
		return zero, err
	}

	return result, nil
}
