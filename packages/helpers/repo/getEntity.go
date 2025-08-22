package helpers_repo

import (
	"context"
	"errors"

	"p9e.in/ugcl/packages/database/pgxpostgres/operations"
	"p9e.in/ugcl/packages/deps"
	ph "p9e.in/ugcl/packages/helpers/utils"
	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/p9log"
)

func GetByID[T any](ctx context.Context, deps deps.ServiceDeps, tableName string, identifier int64) (*T, error) {
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "GetByID"))
	// entityType := ph.GetTypeName[T]()
	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()

	dm := models.DataModel[T]{
		TableName:  tableName,
		Where:      "id = $1",
		WhereArgs:  []any{identifier},
		FieldNames: []string{"*"},
	}

	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeSelect,
	)

	if err != nil {
		// Log timeout or cancellation
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("FindByID operation timed out")
		}
		lg.Errorf("failed to find record by ID: %v", err)
		return nil, err
	}

	return &result, nil

}

func GetByUUID[T any](ctx context.Context, deps deps.ServiceDeps, tableName string, identifier string) (*T, error) {
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "GetByUUID"))
	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()

	dm := models.DataModel[T]{
		TableName:  tableName,
		Where:      "uuid = $1",
		WhereArgs:  []any{identifier},
		FieldNames: []string{"*"},
	}

	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeSelect,
	)

	if err != nil {
		// Log timeout or cancellation
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("FindByID operation timed out")
		}
		lg.Errorf("failed to find user by ID: %v", err)
		return nil, err
	}

	return &result, nil

}

func GetByField[T any](ctx context.Context, deps deps.ServiceDeps, dm models.DataModel[T]) (T, error) {
	var zero T

	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "GetByField"))

	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()
	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeSelect,
	)

	// Handle errors
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return zero, context.DeadlineExceeded // Return timeout error directly
		} else {
			lg.Errorf("Failed to find industry with %v", err)
		}
		return zero, err
	}

	return result, nil
}

func GetByIdentifier[T any](ctx context.Context, deps deps.ServiceDeps,
	tableName string, identifier *models.Identifier) (*T, error) {
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "GetByIdentifier"))
	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()
	entityType := ph.GetTypeName[T]()
	// Validate identifier
	if identifier.Id == 0 && identifier.Uuid == "" {
		lg.Errorf("%s: no valid identifier provided", entityType)
		return nil, errors.New("no valid identifier provided")
	}

	dm := models.DataModel[T]{
		TableName: tableName,
		// FieldNames: []string{"id", "username", "email", "phone", "created_at", "updated_at"},
	}

	// Prioritize Uuid if it's non-empty
	if identifier.Uuid != "" {
		dm.Where = "uuid = $1"
		dm.WhereArgs = []any{identifier.Uuid}
	} else {
		// Fallback to Id if Uuid is empty
		dm.Where = "id = $1"
		dm.WhereArgs = []any{identifier.Id}
	}

	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeSelect,
	)

	if err != nil {
		// Log timeout or cancellation
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("FindByID operation timed out")
		}
		lg.Errorf("failed to find user by ID: %v", err)
		return nil, err
	}

	return &result, nil

}
