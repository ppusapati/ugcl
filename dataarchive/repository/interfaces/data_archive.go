package interfaces

import (
	"context"

	"github.com/google/uuid"
	"p9e.in/ugcl/dataarchive/models"
)

type DataArchiveRepository interface {
	// Retention Policy management
	CreateRetentionPolicy(ctx context.Context, policy *models.RetentionPolicy) (*models.RetentionPolicy, error)
	GetRetentionPolicyByID(ctx context.Context, id uuid.UUID) (*models.RetentionPolicy, error)
	GetRetentionPoliciesByEntity(ctx context.Context, entityType string, schemaName string, databaseName string) ([]*models.RetentionPolicy, error)
	UpdateRetentionPolicy(ctx context.Context, policy *models.RetentionPolicy) error
	DeleteRetentionPolicy(ctx context.Context, id uuid.UUID) error
	ListRetentionPolicies(ctx context.Context, filters *models.RetentionPolicyFilters) ([]*models.RetentionPolicy, error)
	GetActivePolicies(ctx context.Context) ([]*models.RetentionPolicy, error)

	// Archival Job management
	CreateArchivalJob(ctx context.Context, job *models.ArchivalJob) (*models.ArchivalJob, error)
	GetArchivalJobByID(ctx context.Context, id uuid.UUID) (*models.ArchivalJob, error)
	GetJobsByPolicy(ctx context.Context, policyID uuid.UUID) ([]*models.ArchivalJob, error)
	UpdateArchivalJob(ctx context.Context, job *models.ArchivalJob) error
	ListArchivalJobs(ctx context.Context, filters *models.ArchivalJobFilters) ([]*models.ArchivalJob, error)
	GetRunningJobs(ctx context.Context) ([]*models.ArchivalJob, error)
	GetFailedJobs(ctx context.Context, retryable bool) ([]*models.ArchivalJob, error)

	// Legal Hold management
	CreateLegalHold(ctx context.Context, hold *models.LegalHold) (*models.LegalHold, error)
	GetLegalHoldByID(ctx context.Context, id uuid.UUID) (*models.LegalHold, error)
	UpdateLegalHold(ctx context.Context, hold *models.LegalHold) error
	DeleteLegalHold(ctx context.Context, id uuid.UUID) error
	ListLegalHolds(ctx context.Context, filters *models.LegalHoldFilters) ([]*models.LegalHold, error)
	GetActiveLegalHolds(ctx context.Context) ([]*models.LegalHold, error)
	GetLegalHoldsByEntity(ctx context.Context, entityTypes []string, entityIDs []string) ([]*models.LegalHold, error)

	// Archived Data management
	CreateArchivedData(ctx context.Context, data *models.ArchivedData) (*models.ArchivedData, error)
	GetArchivedDataByID(ctx context.Context, id uuid.UUID) (*models.ArchivedData, error)
	GetArchivedDataBySource(ctx context.Context, sourceTable string, sourceSchema string, sourceDatabase string) ([]*models.ArchivedData, error)
	UpdateArchivedData(ctx context.Context, data *models.ArchivedData) error
	ListArchivedData(ctx context.Context, filters *models.ArchivedDataFilters) ([]*models.ArchivedData, error)
	GetExpiredArchives(ctx context.Context) ([]*models.ArchivedData, error)
	GetArchivesOnLegalHold(ctx context.Context) ([]*models.ArchivedData, error)

	// Data Inventory management
	CreateDataInventory(ctx context.Context, inventory *models.DataInventory) (*models.DataInventory, error)
	GetDataInventoryByID(ctx context.Context, id uuid.UUID) (*models.DataInventory, error)
	GetInventoryByTable(ctx context.Context, databaseName string, schemaName string, tableName string) (*models.DataInventory, error)
	UpdateDataInventory(ctx context.Context, inventory *models.DataInventory) error
	ListDataInventory(ctx context.Context, filters *models.DataInventoryFilters) ([]*models.DataInventory, error)
	GetInventoryByClassification(ctx context.Context, classification models.DataClassification) ([]*models.DataInventory, error)

	// Compliance Audit management
	CreateComplianceAudit(ctx context.Context, audit *models.ComplianceAudit) (*models.ComplianceAudit, error)
	GetComplianceAuditByID(ctx context.Context, id uuid.UUID) (*models.ComplianceAudit, error)
	UpdateComplianceAudit(ctx context.Context, audit *models.ComplianceAudit) error
	ListComplianceAudits(ctx context.Context, filters *models.ComplianceAuditFilters) ([]*models.ComplianceAudit, error)
	GetAuditsByDateRange(ctx context.Context, dateRange *models.DateRange) ([]*models.ComplianceAudit, error)
	GetLatestAuditByType(ctx context.Context, auditType models.AuditType) (*models.ComplianceAudit, error)
}