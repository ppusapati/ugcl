// =============================================================================
// repository/job_repository.go - Import job repository implementation
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/databridge/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// JobRepository implements IJobRepository
type JobRepository struct {
	queries *db.Queries
	db      *pgx.Conn
}

// NewJobRepository creates a new job repository instance
func NewJobRepository(database *pgx.Conn) IJobRepository {
	return &JobRepository{
		queries: db.New(database),
		db:      database,
	}
}

// GetJobsByUser retrieves jobs for a specific user with pagination
func (r *JobRepository) GetJobsByUser(ctx context.Context, userID string, limit, offset int32) ([]*db.GetJobsByUserRow, error) {
	jobs, err := r.queries.GetJobsByUser(ctx, db.GetJobsByUserParams{
		CreatedBy: userID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetJobsByUserRow, len(jobs))
	for i, job := range jobs {
		result[i] = &job
	}

	return result, nil
}

// GetJobByID retrieves a job by its ID
func (r *JobRepository) GetJobByID(ctx context.Context, jobID uuid.UUID) (*db.GetJobByIDRow, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(jobID.String()); err != nil {
		return nil, err
	}

	job, err := r.queries.GetJobByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// GetActiveJobs retrieves all currently active jobs
func (r *JobRepository) GetActiveJobs(ctx context.Context) ([]*db.ImportJob, error) {
	jobs, err := r.queries.GetActiveJobs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.ImportJob, len(jobs))
	for i, job := range jobs {
		result[i] = &job
	}

	return result, nil
}

// GetJobsByStatus retrieves jobs by status with pagination
func (r *JobRepository) GetJobsByStatus(ctx context.Context, status string, limit, offset int32) ([]*db.ImportJob, error) {
	jobs, err := r.queries.GetJobsByStatus(ctx, db.GetJobsByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*db.ImportJob, len(jobs))
	for i, job := range jobs {
		result[i] = &job
	}

	return result, nil
}

// CreateJob creates a new import job
func (r *JobRepository) CreateJob(ctx context.Context, params *CreateJobParams) (*db.ImportJob, error) {
	var mappingUUID pgtype.UUID
	if err := mappingUUID.Scan(params.MappingID.String()); err != nil {
		return nil, err
	}

	createParams := db.CreateJobParams{
		JobName:   params.JobName,
		MappingID: mappingUUID,
		FileName:  params.FileName,
		FileSize:  params.FileSize,
		CreatedBy: params.CreatedBy,
	}

	// Handle optional fields
	if params.FileHash != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.FileHash); err != nil {
			return nil, err
		}
		createParams.FileHash = pgText
	}

	if params.TotalRows != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.TotalRows); err != nil {
			return nil, err
		}
		createParams.TotalRows = pgInt4
	}

	job, err := r.queries.CreateJob(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// UpdateJobStatus updates the status of an import job
func (r *JobRepository) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(jobID.String()); err != nil {
		return err
	}

	return r.queries.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:     pgUUID,
		Status: status,
	})
}

// UpdateJobProgress updates the progress of an import job
func (r *JobRepository) UpdateJobProgress(ctx context.Context, params *UpdateJobProgressParams) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return err
	}

	updateParams := db.UpdateJobProgressParams{
		ID:           pgUUID,
		ErrorDetails: params.ErrorDetails,
	}

	// Handle optional fields
	if params.ProcessedRows != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.ProcessedRows); err != nil {
			return err
		}
		updateParams.ProcessedRows = pgInt4
	}

	if params.SuccessfulRows != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.SuccessfulRows); err != nil {
			return err
		}
		updateParams.SuccessfulRows = pgInt4
	}

	if params.FailedRows != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.FailedRows); err != nil {
			return err
		}
		updateParams.FailedRows = pgInt4
	}

	return r.queries.UpdateJobProgress(ctx, updateParams)
}

// UpdateJobError updates a job with error information and sets status to failed
func (r *JobRepository) UpdateJobError(ctx context.Context, jobID uuid.UUID, errorDetails []byte) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(jobID.String()); err != nil {
		return err
	}

	return r.queries.UpdateJobError(ctx, db.UpdateJobErrorParams{
		ID:           pgUUID,
		ErrorDetails: errorDetails,
	})
}

// CancelJob cancels an active import job
func (r *JobRepository) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(jobID.String()); err != nil {
		return err
	}

	return r.queries.CancelJob(ctx, pgUUID)
}

// GetJobStatistics retrieves job statistics for a user since a specific time
func (r *JobRepository) GetJobStatistics(ctx context.Context, userID string, since time.Time) (*db.GetJobStatisticsRow, error) {
	var pgTimestamp pgtype.Timestamptz
	if err := pgTimestamp.Scan(since); err != nil {
		return nil, err
	}

	stats, err := r.queries.GetJobStatistics(ctx, db.GetJobStatisticsParams{
		CreatedBy: userID,
		CreatedAt: pgTimestamp,
	})
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CleanupOldJobs removes old completed/failed/cancelled jobs
func (r *JobRepository) CleanupOldJobs(ctx context.Context, olderThan time.Time) error {
	var pgTimestamp pgtype.Timestamptz
	if err := pgTimestamp.Scan(olderThan); err != nil {
		return err
	}

	return r.queries.CleanupOldJobs(ctx, pgTimestamp)
}