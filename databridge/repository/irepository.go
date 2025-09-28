// =============================================================================
// repository/irepository.go - Repository interfaces for DataBridge module (Import Only)
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/databridge/db/generated"

	"github.com/google/uuid"
)

// IMappingRepository defines import mapping repository interface
type IMappingRepository interface {
	GetMappingsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ImportMapping, error)
	GetMappingByID(ctx context.Context, mappingID uuid.UUID) (*db.ImportMapping, error)
	GetMappingByTableAndName(ctx context.Context, tableID uuid.UUID, mappingName string) (*db.ImportMapping, error)
	GetRecentMappingsByUser(ctx context.Context, userID string, limit int32) ([]*db.ImportMapping, error)
	CreateMapping(ctx context.Context, params *CreateMappingParams) (*db.ImportMapping, error)
	UpdateMapping(ctx context.Context, params *UpdateMappingParams) (*db.ImportMapping, error)
	UpdateMappingUsage(ctx context.Context, mappingID uuid.UUID) error
	DeactivateMapping(ctx context.Context, mappingID uuid.UUID, updatedBy string) error
	SearchMappings(ctx context.Context, query string) ([]*db.ImportMapping, error)
}

// IJobRepository defines import job repository interface
type IJobRepository interface {
	GetJobsByUser(ctx context.Context, userID string, limit, offset int32) ([]*db.GetJobsByUserRow, error)
	GetJobByID(ctx context.Context, jobID uuid.UUID) (*db.GetJobByIDRow, error)
	GetActiveJobs(ctx context.Context) ([]*db.ImportJob, error)
	GetJobsByStatus(ctx context.Context, status string, limit, offset int32) ([]*db.ImportJob, error)
	CreateJob(ctx context.Context, params *CreateJobParams) (*db.ImportJob, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error
	UpdateJobProgress(ctx context.Context, params *UpdateJobProgressParams) error
	UpdateJobError(ctx context.Context, jobID uuid.UUID, errorDetails []byte) error
	CancelJob(ctx context.Context, jobID uuid.UUID) error
	GetJobStatistics(ctx context.Context, userID string, since time.Time) (*db.GetJobStatisticsRow, error)
	CleanupOldJobs(ctx context.Context, olderThan time.Time) error
}

// Parameter structs for repository methods
type CreateMappingParams struct {
	MappingName         string
	Description         *string
	TableID             uuid.UUID
	CSVHeaders          []byte
	FieldMappings       []byte
	TransformationRules []byte
	ValidationRules     []byte
	CreatedBy           string
}

type UpdateMappingParams struct {
	ID                  uuid.UUID
	MappingName         string
	Description         *string
	CSVHeaders          []byte
	FieldMappings       []byte
	TransformationRules []byte
	ValidationRules     []byte
	UpdatedBy           string
}

type CreateJobParams struct {
	JobName   string
	MappingID uuid.UUID
	FileName  string
	FileSize  int64
	FileHash  *string
	TotalRows *int32
	CreatedBy string
}

type UpdateJobProgressParams struct {
	ID             uuid.UUID
	ProcessedRows  *int32
	SuccessfulRows *int32
	FailedRows     *int32
	ErrorDetails   []byte
}