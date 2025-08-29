// =============================================================================
// internal/repository/analytics_repository.go
// =============================================================================
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

type analyticsRepository struct {
	queries *db.Queries
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(queries *db.Queries) IAnalyticsRepository {
	return &analyticsRepository{
		queries: queries,
	}
}

func (r *analyticsRepository) GetFormStatistics(ctx context.Context, formID uuid.UUID) (*db.GetFormStatisticsRow, error) {
	stats, err := r.queries.GetFormStatistics(ctx, db.GetFormStatisticsParams{ID: formID})
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *analyticsRepository) GetUserWorkload(ctx context.Context) ([]*db.GetUserWorkloadRow, error) {
	workloads, err := r.queries.GetUserWorkload(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetUserWorkloadRow, len(workloads))
	for i, workload := range workloads {
		result[i] = workload
	}
	return result, nil
}

func (r *analyticsRepository) GetOverdueInstances(ctx context.Context) ([]*db.GetOverdueInstancesRow, error) {
	instances, err := r.queries.GetOverdueInstances(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetOverdueInstancesRow, len(instances))
	for i, instance := range instances {
		result[i] = instance
	}
	return result, nil
}

func (r *analyticsRepository) GetInstancesRequiringEscalation(ctx context.Context) ([]*db.GetInstancesRequiringEscalationRow, error) {
	instances, err := r.queries.GetInstancesRequiringEscalation(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetInstancesRequiringEscalationRow, len(instances))
	for i, instance := range instances {
		result[i] = instance
	}
	return result, nil
}

func (r *analyticsRepository) GetRecentActivity(ctx context.Context, since time.Time, limit, offset int32) ([]*db.GetRecentActivityRow, error) {
	params := db.GetRecentActivityParams{
		Timestamp: since,
		Limit:     limit,
		Offset:    offset,
	}

	activities, err := r.queries.GetRecentActivity(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetRecentActivityRow, len(activities))
	for i, activity := range activities {
		result[i] = activity
	}
	return result, nil
}
