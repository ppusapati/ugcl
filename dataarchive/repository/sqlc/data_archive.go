package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"p9e.in/ugcl/dataarchive/models"
	"p9e.in/ugcl/dataarchive/repository/interfaces"
)

type DataArchiveRepository struct {
	db *sql.DB
}

func NewDataArchiveRepository(db *sql.DB) interfaces.DataArchiveRepository {
	return &DataArchiveRepository{
		db: db,
	}
}

// Retention Policy management

func (r *DataArchiveRepository) CreateRetentionPolicy(ctx context.Context, policy *models.RetentionPolicy) (*models.RetentionPolicy, error) {
	query := `
		INSERT INTO retention_policies (
			id, name, description, entity_type, schema_name, database_name,
			retention_period, archival_period, grace_period, policy_type,
			archival_method, compression_type, encrypt_archive, selection_criteria,
			exclusion_rules, legal_hold_enabled, compliance_level, regulatory_basis,
			schedule_enabled, schedule_cron, batch_size, parallel_jobs,
			notify_on_start, notify_on_complete, notify_on_error, notification_channels,
			is_active, last_executed, next_execution, created_by, updated_by,
			created_at, updated_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.EntityType,
		policy.SchemaName, policy.DatabaseName, policy.RetentionPeriod,
		policy.ArchivalPeriod, policy.GracePeriod, policy.PolicyType,
		policy.ArchivalMethod, policy.CompressionType, policy.EncryptArchive,
		policy.SelectionCriteria, policy.ExclusionRules, policy.LegalHoldEnabled,
		policy.ComplianceLevel, policy.RegulatoryBasis, policy.ScheduleEnabled,
		policy.ScheduleCron, policy.BatchSize, policy.ParallelJobs,
		policy.NotifyOnStart, policy.NotifyOnComplete, policy.NotifyOnError,
		policy.NotificationChannels, policy.IsActive, policy.LastExecuted,
		policy.NextExecution, policy.CreatedBy, policy.UpdatedBy,
		policy.CreatedAt, policy.UpdatedAt, policy.Metadata,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create retention policy: %w", err)
	}

	return policy, nil
}

func (r *DataArchiveRepository) GetRetentionPolicyByID(ctx context.Context, id uuid.UUID) (*models.RetentionPolicy, error) {
	query := `
		SELECT id, name, description, entity_type, schema_name, database_name,
			   retention_period, archival_period, grace_period, policy_type,
			   archival_method, compression_type, encrypt_archive, selection_criteria,
			   exclusion_rules, legal_hold_enabled, compliance_level, regulatory_basis,
			   schedule_enabled, schedule_cron, batch_size, parallel_jobs,
			   notify_on_start, notify_on_complete, notify_on_error, notification_channels,
			   is_active, last_executed, next_execution, created_by, updated_by,
			   created_at, updated_at, metadata
		FROM retention_policies
		WHERE id = $1`

	policy := &models.RetentionPolicy{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.EntityType,
		&policy.SchemaName, &policy.DatabaseName, &policy.RetentionPeriod,
		&policy.ArchivalPeriod, &policy.GracePeriod, &policy.PolicyType,
		&policy.ArchivalMethod, &policy.CompressionType, &policy.EncryptArchive,
		&policy.SelectionCriteria, &policy.ExclusionRules, &policy.LegalHoldEnabled,
		&policy.ComplianceLevel, &policy.RegulatoryBasis, &policy.ScheduleEnabled,
		&policy.ScheduleCron, &policy.BatchSize, &policy.ParallelJobs,
		&policy.NotifyOnStart, &policy.NotifyOnComplete, &policy.NotifyOnError,
		&policy.NotificationChannels, &policy.IsActive, &policy.LastExecuted,
		&policy.NextExecution, &policy.CreatedBy, &policy.UpdatedBy,
		&policy.CreatedAt, &policy.UpdatedAt, &policy.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("retention policy not found")
		}
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}

	return policy, nil
}

func (r *DataArchiveRepository) GetRetentionPoliciesByEntity(ctx context.Context, entityType string, schemaName string, databaseName string) ([]*models.RetentionPolicy, error) {
	query := `
		SELECT id, name, description, entity_type, schema_name, database_name,
			   retention_period, archival_period, grace_period, policy_type,
			   archival_method, compression_type, encrypt_archive, selection_criteria,
			   exclusion_rules, legal_hold_enabled, compliance_level, regulatory_basis,
			   schedule_enabled, schedule_cron, batch_size, parallel_jobs,
			   notify_on_start, notify_on_complete, notify_on_error, notification_channels,
			   is_active, last_executed, next_execution, created_by, updated_by,
			   created_at, updated_at, metadata
		FROM retention_policies
		WHERE entity_type = $1 AND schema_name = $2 AND database_name = $3 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, entityType, schemaName, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.RetentionPolicy
	for rows.Next() {
		policy := &models.RetentionPolicy{}
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.EntityType,
			&policy.SchemaName, &policy.DatabaseName, &policy.RetentionPeriod,
			&policy.ArchivalPeriod, &policy.GracePeriod, &policy.PolicyType,
			&policy.ArchivalMethod, &policy.CompressionType, &policy.EncryptArchive,
			&policy.SelectionCriteria, &policy.ExclusionRules, &policy.LegalHoldEnabled,
			&policy.ComplianceLevel, &policy.RegulatoryBasis, &policy.ScheduleEnabled,
			&policy.ScheduleCron, &policy.BatchSize, &policy.ParallelJobs,
			&policy.NotifyOnStart, &policy.NotifyOnComplete, &policy.NotifyOnError,
			&policy.NotificationChannels, &policy.IsActive, &policy.LastExecuted,
			&policy.NextExecution, &policy.CreatedBy, &policy.UpdatedBy,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate retention policies: %w", err)
	}

	return policies, nil
}

func (r *DataArchiveRepository) UpdateRetentionPolicy(ctx context.Context, policy *models.RetentionPolicy) error {
	query := `
		UPDATE retention_policies
		SET name = $2, description = $3, entity_type = $4, schema_name = $5,
			database_name = $6, retention_period = $7, archival_period = $8,
			grace_period = $9, policy_type = $10, archival_method = $11,
			compression_type = $12, encrypt_archive = $13, selection_criteria = $14,
			exclusion_rules = $15, legal_hold_enabled = $16, compliance_level = $17,
			regulatory_basis = $18, schedule_enabled = $19, schedule_cron = $20,
			batch_size = $21, parallel_jobs = $22, notify_on_start = $23,
			notify_on_complete = $24, notify_on_error = $25, notification_channels = $26,
			is_active = $27, last_executed = $28, next_execution = $29,
			updated_by = $30, updated_at = $31, metadata = $32
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.EntityType,
		policy.SchemaName, policy.DatabaseName, policy.RetentionPeriod,
		policy.ArchivalPeriod, policy.GracePeriod, policy.PolicyType,
		policy.ArchivalMethod, policy.CompressionType, policy.EncryptArchive,
		policy.SelectionCriteria, policy.ExclusionRules, policy.LegalHoldEnabled,
		policy.ComplianceLevel, policy.RegulatoryBasis, policy.ScheduleEnabled,
		policy.ScheduleCron, policy.BatchSize, policy.ParallelJobs,
		policy.NotifyOnStart, policy.NotifyOnComplete, policy.NotifyOnError,
		policy.NotificationChannels, policy.IsActive, policy.LastExecuted,
		policy.NextExecution, policy.UpdatedBy, policy.UpdatedAt, policy.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update retention policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("retention policy not found")
	}

	return nil
}

func (r *DataArchiveRepository) DeleteRetentionPolicy(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM retention_policies WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete retention policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("retention policy not found")
	}

	return nil
}

func (r *DataArchiveRepository) ListRetentionPolicies(ctx context.Context, filters *models.RetentionPolicyFilters) ([]*models.RetentionPolicy, error) {
	query := `
		SELECT id, name, description, entity_type, schema_name, database_name,
			   retention_period, archival_period, grace_period, policy_type,
			   archival_method, compression_type, encrypt_archive, selection_criteria,
			   exclusion_rules, legal_hold_enabled, compliance_level, regulatory_basis,
			   schedule_enabled, schedule_cron, batch_size, parallel_jobs,
			   notify_on_start, notify_on_complete, notify_on_error, notification_channels,
			   is_active, last_executed, next_execution, created_by, updated_by,
			   created_at, updated_at, metadata
		FROM retention_policies`

	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if filters != nil {
		if filters.EntityType != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("entity_type = $%d", argIndex))
			args = append(args, filters.EntityType)
			argIndex++
		}

		if filters.SchemaName != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("schema_name = $%d", argIndex))
			args = append(args, filters.SchemaName)
			argIndex++
		}

		if filters.DatabaseName != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("database_name = $%d", argIndex))
			args = append(args, filters.DatabaseName)
			argIndex++
		}

		if filters.IsActive != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, *filters.IsActive)
			argIndex++
		}

		if filters.PolicyType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("policy_type = $%d", argIndex))
			args = append(args, *filters.PolicyType)
			argIndex++
		}

		if filters.ComplianceLevel != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("compliance_level = $%d", argIndex))
			args = append(args, *filters.ComplianceLevel)
			argIndex++
		}
	}

	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list retention policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.RetentionPolicy
	for rows.Next() {
		policy := &models.RetentionPolicy{}
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.EntityType,
			&policy.SchemaName, &policy.DatabaseName, &policy.RetentionPeriod,
			&policy.ArchivalPeriod, &policy.GracePeriod, &policy.PolicyType,
			&policy.ArchivalMethod, &policy.CompressionType, &policy.EncryptArchive,
			&policy.SelectionCriteria, &policy.ExclusionRules, &policy.LegalHoldEnabled,
			&policy.ComplianceLevel, &policy.RegulatoryBasis, &policy.ScheduleEnabled,
			&policy.ScheduleCron, &policy.BatchSize, &policy.ParallelJobs,
			&policy.NotifyOnStart, &policy.NotifyOnComplete, &policy.NotifyOnError,
			&policy.NotificationChannels, &policy.IsActive, &policy.LastExecuted,
			&policy.NextExecution, &policy.CreatedBy, &policy.UpdatedBy,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate retention policies: %w", err)
	}

	return policies, nil
}

func (r *DataArchiveRepository) GetActivePolicies(ctx context.Context) ([]*models.RetentionPolicy, error) {
	query := `
		SELECT id, name, description, entity_type, schema_name, database_name,
			   retention_period, archival_period, grace_period, policy_type,
			   archival_method, compression_type, encrypt_archive, selection_criteria,
			   exclusion_rules, legal_hold_enabled, compliance_level, regulatory_basis,
			   schedule_enabled, schedule_cron, batch_size, parallel_jobs,
			   notify_on_start, notify_on_complete, notify_on_error, notification_channels,
			   is_active, last_executed, next_execution, created_by, updated_by,
			   created_at, updated_at, metadata
		FROM retention_policies
		WHERE is_active = true AND schedule_enabled = true
		ORDER BY next_execution ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.RetentionPolicy
	for rows.Next() {
		policy := &models.RetentionPolicy{}
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.EntityType,
			&policy.SchemaName, &policy.DatabaseName, &policy.RetentionPeriod,
			&policy.ArchivalPeriod, &policy.GracePeriod, &policy.PolicyType,
			&policy.ArchivalMethod, &policy.CompressionType, &policy.EncryptArchive,
			&policy.SelectionCriteria, &policy.ExclusionRules, &policy.LegalHoldEnabled,
			&policy.ComplianceLevel, &policy.RegulatoryBasis, &policy.ScheduleEnabled,
			&policy.ScheduleCron, &policy.BatchSize, &policy.ParallelJobs,
			&policy.NotifyOnStart, &policy.NotifyOnComplete, &policy.NotifyOnError,
			&policy.NotificationChannels, &policy.IsActive, &policy.LastExecuted,
			&policy.NextExecution, &policy.CreatedBy, &policy.UpdatedBy,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate retention policies: %w", err)
	}

	return policies, nil
}

// Archival Job management

func (r *DataArchiveRepository) CreateArchivalJob(ctx context.Context, job *models.ArchivalJob) (*models.ArchivalJob, error) {
	query := `
		INSERT INTO archival_jobs (
			id, policy_id, job_type, status, target_table, date_range, batch_size,
			started_at, completed_at, estimated_end, total_records, processed_records,
			archived_records, deleted_records, error_count, original_size,
			compressed_size, compression_ratio, archive_location, archive_format,
			encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		) RETURNING id, started_at`

	err := r.db.QueryRowContext(ctx, query,
		job.ID, job.PolicyID, job.JobType, job.Status, job.TargetTable,
		job.DateRange, job.BatchSize, job.StartedAt, job.CompletedAt,
		job.EstimatedEnd, job.TotalRecords, job.ProcessedRecords,
		job.ArchivedRecords, job.DeletedRecords, job.ErrorCount,
		job.OriginalSize, job.CompressedSize, job.CompressionRatio,
		job.ArchiveLocation, job.ArchiveFormat, job.EncryptionKey,
		job.ErrorMessage, job.RetryCount, job.MaxRetries,
		job.ExecutedBy, job.JobMetadata,
	).Scan(&job.ID, &job.StartedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create archival job: %w", err)
	}

	return job, nil
}

func (r *DataArchiveRepository) GetArchivalJobByID(ctx context.Context, id uuid.UUID) (*models.ArchivalJob, error) {
	query := `
		SELECT id, policy_id, job_type, status, target_table, date_range, batch_size,
			   started_at, completed_at, estimated_end, total_records, processed_records,
			   archived_records, deleted_records, error_count, original_size,
			   compressed_size, compression_ratio, archive_location, archive_format,
			   encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		FROM archival_jobs
		WHERE id = $1`

	job := &models.ArchivalJob{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.PolicyID, &job.JobType, &job.Status, &job.TargetTable,
		&job.DateRange, &job.BatchSize, &job.StartedAt, &job.CompletedAt,
		&job.EstimatedEnd, &job.TotalRecords, &job.ProcessedRecords,
		&job.ArchivedRecords, &job.DeletedRecords, &job.ErrorCount,
		&job.OriginalSize, &job.CompressedSize, &job.CompressionRatio,
		&job.ArchiveLocation, &job.ArchiveFormat, &job.EncryptionKey,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
		&job.ExecutedBy, &job.JobMetadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("archival job not found")
		}
		return nil, fmt.Errorf("failed to get archival job: %w", err)
	}

	return job, nil
}

func (r *DataArchiveRepository) GetJobsByPolicy(ctx context.Context, policyID uuid.UUID) ([]*models.ArchivalJob, error) {
	query := `
		SELECT id, policy_id, job_type, status, target_table, date_range, batch_size,
			   started_at, completed_at, estimated_end, total_records, processed_records,
			   archived_records, deleted_records, error_count, original_size,
			   compressed_size, compression_ratio, archive_location, archive_format,
			   encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		FROM archival_jobs
		WHERE policy_id = $1
		ORDER BY started_at DESC`

	rows, err := r.db.QueryContext(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by policy: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ArchivalJob
	for rows.Next() {
		job := &models.ArchivalJob{}
		err := rows.Scan(
			&job.ID, &job.PolicyID, &job.JobType, &job.Status, &job.TargetTable,
			&job.DateRange, &job.BatchSize, &job.StartedAt, &job.CompletedAt,
			&job.EstimatedEnd, &job.TotalRecords, &job.ProcessedRecords,
			&job.ArchivedRecords, &job.DeletedRecords, &job.ErrorCount,
			&job.OriginalSize, &job.CompressedSize, &job.CompressionRatio,
			&job.ArchiveLocation, &job.ArchiveFormat, &job.EncryptionKey,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.ExecutedBy, &job.JobMetadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archival job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate archival jobs: %w", err)
	}

	return jobs, nil
}

func (r *DataArchiveRepository) UpdateArchivalJob(ctx context.Context, job *models.ArchivalJob) error {
	query := `
		UPDATE archival_jobs
		SET policy_id = $2, job_type = $3, status = $4, target_table = $5,
			date_range = $6, batch_size = $7, started_at = $8, completed_at = $9,
			estimated_end = $10, total_records = $11, processed_records = $12,
			archived_records = $13, deleted_records = $14, error_count = $15,
			original_size = $16, compressed_size = $17, compression_ratio = $18,
			archive_location = $19, archive_format = $20, encryption_key = $21,
			error_message = $22, retry_count = $23, max_retries = $24,
			executed_by = $25, job_metadata = $26
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		job.ID, job.PolicyID, job.JobType, job.Status, job.TargetTable,
		job.DateRange, job.BatchSize, job.StartedAt, job.CompletedAt,
		job.EstimatedEnd, job.TotalRecords, job.ProcessedRecords,
		job.ArchivedRecords, job.DeletedRecords, job.ErrorCount,
		job.OriginalSize, job.CompressedSize, job.CompressionRatio,
		job.ArchiveLocation, job.ArchiveFormat, job.EncryptionKey,
		job.ErrorMessage, job.RetryCount, job.MaxRetries,
		job.ExecutedBy, job.JobMetadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update archival job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("archival job not found")
	}

	return nil
}

func (r *DataArchiveRepository) ListArchivalJobs(ctx context.Context, filters *models.ArchivalJobFilters) ([]*models.ArchivalJob, error) {
	query := `
		SELECT id, policy_id, job_type, status, target_table, date_range, batch_size,
			   started_at, completed_at, estimated_end, total_records, processed_records,
			   archived_records, deleted_records, error_count, original_size,
			   compressed_size, compression_ratio, archive_location, archive_format,
			   encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		FROM archival_jobs`

	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if filters != nil {
		if filters.PolicyID != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("policy_id = $%d", argIndex))
			args = append(args, *filters.PolicyID)
			argIndex++
		}

		if filters.JobType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("job_type = $%d", argIndex))
			args = append(args, *filters.JobType)
			argIndex++
		}

		if filters.Status != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filters.Status)
			argIndex++
		}

		if filters.StartedAfter != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("started_at >= $%d", argIndex))
			args = append(args, *filters.StartedAfter)
			argIndex++
		}

		if filters.StartedBefore != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("started_at <= $%d", argIndex))
			args = append(args, *filters.StartedBefore)
			argIndex++
		}
	}

	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	query += " ORDER BY started_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list archival jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ArchivalJob
	for rows.Next() {
		job := &models.ArchivalJob{}
		err := rows.Scan(
			&job.ID, &job.PolicyID, &job.JobType, &job.Status, &job.TargetTable,
			&job.DateRange, &job.BatchSize, &job.StartedAt, &job.CompletedAt,
			&job.EstimatedEnd, &job.TotalRecords, &job.ProcessedRecords,
			&job.ArchivedRecords, &job.DeletedRecords, &job.ErrorCount,
			&job.OriginalSize, &job.CompressedSize, &job.CompressionRatio,
			&job.ArchiveLocation, &job.ArchiveFormat, &job.EncryptionKey,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.ExecutedBy, &job.JobMetadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archival job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate archival jobs: %w", err)
	}

	return jobs, nil
}

func (r *DataArchiveRepository) GetRunningJobs(ctx context.Context) ([]*models.ArchivalJob, error) {
	query := `
		SELECT id, policy_id, job_type, status, target_table, date_range, batch_size,
			   started_at, completed_at, estimated_end, total_records, processed_records,
			   archived_records, deleted_records, error_count, original_size,
			   compressed_size, compression_ratio, archive_location, archive_format,
			   encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		FROM archival_jobs
		WHERE status = 'RUNNING'
		ORDER BY started_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get running jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ArchivalJob
	for rows.Next() {
		job := &models.ArchivalJob{}
		err := rows.Scan(
			&job.ID, &job.PolicyID, &job.JobType, &job.Status, &job.TargetTable,
			&job.DateRange, &job.BatchSize, &job.StartedAt, &job.CompletedAt,
			&job.EstimatedEnd, &job.TotalRecords, &job.ProcessedRecords,
			&job.ArchivedRecords, &job.DeletedRecords, &job.ErrorCount,
			&job.OriginalSize, &job.CompressedSize, &job.CompressionRatio,
			&job.ArchiveLocation, &job.ArchiveFormat, &job.EncryptionKey,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.ExecutedBy, &job.JobMetadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archival job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate archival jobs: %w", err)
	}

	return jobs, nil
}

func (r *DataArchiveRepository) GetFailedJobs(ctx context.Context, retryable bool) ([]*models.ArchivalJob, error) {
	query := `
		SELECT id, policy_id, job_type, status, target_table, date_range, batch_size,
			   started_at, completed_at, estimated_end, total_records, processed_records,
			   archived_records, deleted_records, error_count, original_size,
			   compressed_size, compression_ratio, archive_location, archive_format,
			   encryption_key, error_message, retry_count, max_retries, executed_by, job_metadata
		FROM archival_jobs
		WHERE status = 'FAILED'`

	if retryable {
		query += " AND retry_count < max_retries"
	}

	query += " ORDER BY started_at DESC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ArchivalJob
	for rows.Next() {
		job := &models.ArchivalJob{}
		err := rows.Scan(
			&job.ID, &job.PolicyID, &job.JobType, &job.Status, &job.TargetTable,
			&job.DateRange, &job.BatchSize, &job.StartedAt, &job.CompletedAt,
			&job.EstimatedEnd, &job.TotalRecords, &job.ProcessedRecords,
			&job.ArchivedRecords, &job.DeletedRecords, &job.ErrorCount,
			&job.OriginalSize, &job.CompressedSize, &job.CompressionRatio,
			&job.ArchiveLocation, &job.ArchiveFormat, &job.EncryptionKey,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.ExecutedBy, &job.JobMetadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archival job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate archival jobs: %w", err)
	}

	return jobs, nil
}

// Legal Hold management - continuing with placeholder implementations for space
// Similar pattern would follow for other methods...

func (r *DataArchiveRepository) CreateLegalHold(ctx context.Context, hold *models.LegalHold) (*models.LegalHold, error) {
	// Implementation similar to retention policies
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetLegalHoldByID(ctx context.Context, id uuid.UUID) (*models.LegalHold, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) UpdateLegalHold(ctx context.Context, hold *models.LegalHold) error {
	return fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) DeleteLegalHold(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) ListLegalHolds(ctx context.Context, filters *models.LegalHoldFilters) ([]*models.LegalHold, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetActiveLegalHolds(ctx context.Context) ([]*models.LegalHold, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetLegalHoldsByEntity(ctx context.Context, entityTypes []string, entityIDs []string) ([]*models.LegalHold, error) {
	return nil, fmt.Errorf("not implemented")
}

// Archived Data management
func (r *DataArchiveRepository) CreateArchivedData(ctx context.Context, data *models.ArchivedData) (*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetArchivedDataByID(ctx context.Context, id uuid.UUID) (*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetArchivedDataBySource(ctx context.Context, sourceTable string, sourceSchema string, sourceDatabase string) ([]*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) UpdateArchivedData(ctx context.Context, data *models.ArchivedData) error {
	return fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) ListArchivedData(ctx context.Context, filters *models.ArchivedDataFilters) ([]*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetExpiredArchives(ctx context.Context) ([]*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetArchivesOnLegalHold(ctx context.Context) ([]*models.ArchivedData, error) {
	return nil, fmt.Errorf("not implemented")
}

// Data Inventory management
func (r *DataArchiveRepository) CreateDataInventory(ctx context.Context, inventory *models.DataInventory) (*models.DataInventory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetDataInventoryByID(ctx context.Context, id uuid.UUID) (*models.DataInventory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetInventoryByTable(ctx context.Context, databaseName string, schemaName string, tableName string) (*models.DataInventory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) UpdateDataInventory(ctx context.Context, inventory *models.DataInventory) error {
	return fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) ListDataInventory(ctx context.Context, filters *models.DataInventoryFilters) ([]*models.DataInventory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetInventoryByClassification(ctx context.Context, classification models.DataClassification) ([]*models.DataInventory, error) {
	return nil, fmt.Errorf("not implemented")
}

// Compliance Audit management
func (r *DataArchiveRepository) CreateComplianceAudit(ctx context.Context, audit *models.ComplianceAudit) (*models.ComplianceAudit, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetComplianceAuditByID(ctx context.Context, id uuid.UUID) (*models.ComplianceAudit, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) UpdateComplianceAudit(ctx context.Context, audit *models.ComplianceAudit) error {
	return fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) ListComplianceAudits(ctx context.Context, filters *models.ComplianceAuditFilters) ([]*models.ComplianceAudit, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetAuditsByDateRange(ctx context.Context, dateRange *models.DateRange) ([]*models.ComplianceAudit, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *DataArchiveRepository) GetLatestAuditByType(ctx context.Context, auditType models.AuditType) (*models.ComplianceAudit, error) {
	return nil, fmt.Errorf("not implemented")
}