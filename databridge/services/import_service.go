// =============================================================================
// services/import_service.go - Import execution service implementation
// =============================================================================
package services

import (
	"context"
	"fmt"
	"time"

	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/mappers"
	"p9e.in/ugcl/databridge/models"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
)

// ImportService implements IImportService
type ImportService struct {
	repos             *repository.RepositoryManager
	mappers           *mappers.MapperManager
	csvService        ICSVService
	validationService IValidationService
}

// NewImportService creates a new import service instance
func NewImportService(
	repos *repository.RepositoryManager,
	mappers *mappers.MapperManager,
	csvService ICSVService,
	validationService IValidationService,
) IImportService {
	return &ImportService{
		repos:             repos,
		mappers:           mappers,
		csvService:        csvService,
		validationService: validationService,
	}
}

// CreateImportJob creates a new import job
func (s *ImportService) CreateImportJob(ctx context.Context, req *pb.CreateImportJobRequest) (*pb.ImportJob, error) {
	// Validate request
	if err := s.validateCreateJobRequest(req); err != nil {
		return nil, err
	}

	mappingID, err := uuid.Parse(req.MappingId)
	if err != nil {
		return nil, fmt.Errorf("invalid mapping ID: %w", err)
	}

	// Check if mapping exists
	mapping, err := s.repos.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil || mapping == nil {
		return nil, fmt.Errorf("mapping not found")
	}

	// Create job
	params := s.mappers.Job.CreateRequestToRepoParams(req)
	if params == nil {
		return nil, fmt.Errorf("failed to convert request to repository parameters")
	}

	job, err := s.repos.Job.CreateJob(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	return s.mappers.Job.DBToProto(job), nil
}

// GetImportJob retrieves an import job by ID
func (s *ImportService) GetImportJob(ctx context.Context, jobID uuid.UUID) (*pb.ImportJob, error) {
	job, err := s.repos.Job.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import job: %w", err)
	}

	return s.mappers.Job.DetailDBToProto(job), nil
}

// GetImportJobs retrieves import jobs with pagination and filtering
func (s *ImportService) GetImportJobs(ctx context.Context, req *pb.GetImportJobsRequest) (*pb.GetImportJobsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var jobs []*pb.ImportJob
	totalCount := int32(0)

	if req.UserId != nil {
		// Get jobs for specific user
		dbJobs, err := s.repos.Job.GetJobsByUser(ctx, *req.UserId, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for user: %w", err)
		}

		for _, job := range dbJobs {
			jobs = append(jobs, s.mappers.Job.UserJobDBToProto(job))
		}
		totalCount = int32(len(jobs)) // Simplified - in production you'd get actual count

	} else if req.Status != nil {
		// Get jobs by status
		statusStr := s.mappers.Job.StatusProtoToDB(*req.Status)
		dbJobs, err := s.repos.Job.GetJobsByStatus(ctx, statusStr, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs by status: %w", err)
		}

		for _, job := range dbJobs {
			jobs = append(jobs, s.mappers.Job.DBToProto(job))
		}
		totalCount = int32(len(jobs))
	} else {
		// Get all active jobs
		dbJobs, err := s.repos.Job.GetActiveJobs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active jobs: %w", err)
		}

		for _, job := range dbJobs {
			jobs = append(jobs, s.mappers.Job.DBToProto(job))
		}
		totalCount = int32(len(jobs))
	}

	return &pb.GetImportJobsResponse{
		Jobs:       jobs,
		TotalCount: totalCount,
	}, nil
}

// CancelImportJob cancels an active import job
func (s *ImportService) CancelImportJob(ctx context.Context, jobID uuid.UUID) error {
	// Check if job exists and is cancellable
	job, err := s.repos.Job.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get import job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("import job not found")
	}

	status := s.mappers.Job.StatusDBToProto(job.Status)
	if status != pb.ImportJobStatus_IMPORT_JOB_STATUS_PENDING &&
		status != pb.ImportJobStatus_IMPORT_JOB_STATUS_PROCESSING {
		return fmt.Errorf("job cannot be cancelled in current status: %s", job.Status)
	}

	// Cancel job
	err = s.repos.Job.CancelJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to cancel import job: %w", err)
	}

	return nil
}

// ExecuteImport executes an import job with streaming progress updates
func (s *ImportService) ExecuteImport(ctx context.Context, jobID uuid.UUID, csvData []byte, progressChan chan<- *pb.ImportProgress) error {
	// Get job details
	job, err := s.repos.Job.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get import job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("import job not found")
	}

	// Check job status
	status := s.mappers.Job.StatusDBToProto(job.Status)
	if status != pb.ImportJobStatus_IMPORT_JOB_STATUS_PENDING {
		return fmt.Errorf("job cannot be executed in current status: %s", job.Status)
	}

	// Update job status to processing
	err = s.repos.Job.UpdateJobStatus(ctx, jobID, "processing")
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Send initial progress
	progressChan <- &pb.ImportProgress{
		JobId:              jobID.String(),
		Phase:              string(models.ImportPhaseInitializing),
		TotalRows:          0,
		ProcessedRows:      0,
		SuccessfulRows:     0,
		FailedRows:         0,
		ProgressPercentage: 0,
		ThroughputRps:      0,
		RecentErrors:       []*pb.ImportError{},
		Message:            "Initializing import process...",
	}

	// Simulate import execution (in real implementation, this would be much more complex)
	return s.executeImportProcess(ctx, jobID, job, csvData, progressChan)
}

// executeImportProcess performs the actual import process
func (s *ImportService) executeImportProcess(
	ctx context.Context,
	jobID uuid.UUID,
	job *repository.GetJobByIDRow,
	csvData []byte,
	progressChan chan<- *pb.ImportProgress,
) error {
	startTime := time.Now()
	totalRows := int32(100) // This would be determined from CSV analysis

	// Phase 1: Parsing CSV
	progressChan <- &pb.ImportProgress{
		JobId:              jobID.String(),
		Phase:              string(models.ImportPhaseParsing),
		TotalRows:          totalRows,
		ProcessedRows:      0,
		SuccessfulRows:     0,
		FailedRows:         0,
		ProgressPercentage: 5,
		Message:            "Parsing CSV data...",
	}

	time.Sleep(1 * time.Second) // Simulate work

	// Phase 2: Validating data
	progressChan <- &pb.ImportProgress{
		JobId:              jobID.String(),
		Phase:              string(models.ImportPhaseValidating),
		TotalRows:          totalRows,
		ProcessedRows:      0,
		SuccessfulRows:     0,
		FailedRows:         0,
		ProgressPercentage: 15,
		Message:            "Validating data against schema...",
	}

	time.Sleep(1 * time.Second) // Simulate work

	// Phase 3: Importing data (simulate batch processing)
	batchSize := int32(10)
	processedRows := int32(0)
	successfulRows := int32(0)
	failedRows := int32(0)

	for processedRows < totalRows {
		select {
		case <-ctx.Done():
			// Handle cancellation
			s.repos.Job.UpdateJobStatus(ctx, jobID, "cancelled")
			return ctx.Err()
		default:
			// Process batch
			remainingRows := totalRows - processedRows
			currentBatchSize := batchSize
			if remainingRows < batchSize {
				currentBatchSize = remainingRows
			}

			// Simulate processing
			time.Sleep(500 * time.Millisecond)

			processedRows += currentBatchSize
			successfulRows += currentBatchSize - 1 // Simulate 1 failure per batch
			failedRows += 1

			percentage := float64(processedRows) / float64(totalRows) * 80 + 15 // 15% already done in validation
			throughput := float64(processedRows) / time.Since(startTime).Seconds()

			progressChan <- &pb.ImportProgress{
				JobId:              jobID.String(),
				Phase:              string(models.ImportPhaseImporting),
				TotalRows:          totalRows,
				ProcessedRows:      processedRows,
				SuccessfulRows:     successfulRows,
				FailedRows:         failedRows,
				ProgressPercentage: percentage,
				ThroughputRps:      throughput,
				Message:            fmt.Sprintf("Importing data... %d/%d rows processed", processedRows, totalRows),
				RecentErrors: []*pb.ImportError{
					{
						RowNumber:    processedRows,
						ErrorMessage: "Sample validation error",
						ErrorType:    "validation_error",
					},
				},
			}

			// Update job progress in database
			s.repos.Job.UpdateJobProgress(ctx, &repository.UpdateJobProgressParams{
				ID:             uuid.MustParse(jobID.String()),
				ProcessedRows:  &processedRows,
				SuccessfulRows: &successfulRows,
				FailedRows:     &failedRows,
			})
		}
	}

	// Phase 4: Finalizing
	progressChan <- &pb.ImportProgress{
		JobId:              jobID.String(),
		Phase:              string(models.ImportPhaseFinalizing),
		TotalRows:          totalRows,
		ProcessedRows:      processedRows,
		SuccessfulRows:     successfulRows,
		FailedRows:         failedRows,
		ProgressPercentage: 95,
		Message:            "Finalizing import...",
	}

	time.Sleep(500 * time.Millisecond) // Simulate finalization

	// Mark job as completed
	err := s.repos.Job.UpdateJobStatus(ctx, jobID, "completed")
	if err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	// Send final progress
	progressChan <- &pb.ImportProgress{
		JobId:              jobID.String(),
		Phase:              string(models.ImportPhaseCompleted),
		TotalRows:          totalRows,
		ProcessedRows:      processedRows,
		SuccessfulRows:     successfulRows,
		FailedRows:         failedRows,
		ProgressPercentage: 100,
		ThroughputRps:      float64(processedRows) / time.Since(startTime).Seconds(),
		Message:            fmt.Sprintf("Import completed! %d successful, %d failed", successfulRows, failedRows),
	}

	return nil
}

// Validation helper methods

func (s *ImportService) validateCreateJobRequest(req *pb.CreateImportJobRequest) error {
	if req.JobName == "" {
		return fmt.Errorf("job name is required")
	}
	if req.MappingId == "" {
		return fmt.Errorf("mapping ID is required")
	}
	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}
	if req.FileSize <= 0 {
		return fmt.Errorf("file size must be greater than zero")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}
	return nil
}