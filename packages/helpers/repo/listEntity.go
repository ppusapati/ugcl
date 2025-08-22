package helpers_repo

import (
	"context"
	"strings"
	"time"

	"p9e.in/ugcl/packages/database/pgxpostgres/builder"
	"p9e.in/ugcl/packages/database/pgxpostgres/operations"
	"p9e.in/ugcl/packages/deps"
	ph "p9e.in/ugcl/packages/helpers/utils"
	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/p9log"
)

func ListEntity[T any](ctx context.Context, deps deps.ServiceDeps, tableName string, search *models.SearchCriteria,
	extractFunc func([]string) ([]string, error)) ([]*T, error) {

	var zero []*T
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "ListEntity"))
	entityType := ph.GetTypeName[T]()

	// Apply timeout, potentially a long-running query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()
	// Start tracing the operation
	lg.Infof("Starting operation: %s", "Create Repo"+entityType)
	startTime := time.Now()
	// Build where clause dynamically based on search criteria
	whereClause, args := builder.BuildWhereClause(search)

	// Determine which fields to select based on FieldMask
	// If no field mask is provided, select all fields
	//TODO: There is a bug in the field mask, it should be a slice of strings, not a single string
	var fieldNames []string
	if len(fieldNames) == 0 { // No field names specified, so select all columns
		fieldNames, _ = extractFunc(search.FieldMask.GetPaths())
	} else {
		// If no field mask, select all fields
		fieldNames = []string{"*"}
	}
	dm := models.DataModel[T]{
		TableName:  tableName,
		FieldNames: fieldNames,
		Where:      whereClause,
		WhereArgs:  args,
		OrderBy:    strings.Join(search.Sort, ", "),
		Limit:      search.PageSize,
		Offset:     search.PageOffset,
	}
	result, err := operations.ExecuteQuerySlice(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeSelect,
	)
	// End tracing the operation
	elapsedTime := time.Since(startTime)
	lg.Infof("Completed operation: %s in %s", entityType, elapsedTime)

	// Collect metrics
	deps.Metrics.RecordDBOperation(entityType, elapsedTime, true)
	if err != nil {
		// Log timeout or cancellation
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("SearchUsers operation timed out")
		}
		lg.Errorf("failed to search users: %v", err)
		return zero, err
	}
	// Convert result to []*T
	convertedResult := make([]*T, len(result))
	for i := range result {
		convertedResult[i] = &result[i]
	}
	return convertedResult, nil
}
