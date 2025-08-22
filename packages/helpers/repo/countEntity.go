package helpers_repo

import (
	"context"

	"p9e.in/ugcl/packages/database/pgxpostgres/operations"
	"p9e.in/ugcl/packages/deps"
	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/p9log"
)

func CountQuery[T any](ctx context.Context, deps deps.ServiceDeps, tableName string) (T, error) {
	// Create a contextual logger for this function
	lg := p9log.NewHelper(p9log.With(deps.Log, "CountQuery"))
	// Apply timeout, not considered a long query
	tctx, cancel := deps.Tp.ApplyTimeout(ctx, false)
	defer cancel()
	dm := models.DataModel[T]{
		TableName: tableName,
	}

	result, err := operations.ExecuteQuery(
		tctx,
		deps.Pool,
		&dm,
		operations.QueryTypeCount,
	)

	if err != nil {
		if tctx.Err() == context.DeadlineExceeded {
			lg.Errorf("Count operation for table %s timed out", tableName)
		} else {
			lg.Errorf("failed to count for table %s : %v", tableName, err.Error())
		}
		var zero T
		return zero, err
	}
	return result, nil
}
