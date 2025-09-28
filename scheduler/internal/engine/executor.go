package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"p9e.in/ugcl/scheduler/models"
)

// ServiceExecutor handles job execution by calling target services
type ServiceExecutor struct {
	httpClient    *http.Client
	grpcConns     map[string]*grpc.ClientConn
	repositories  *RepositoryManager
	hostname      string
}

// RepositoryManager interface for accessing job data
type RepositoryManager interface {
	CreateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error)
	UpdateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error)
	UpdateJobStats(ctx context.Context, jobID uuid.UUID, status string) error
}

// NewServiceExecutor creates a new service executor
func NewServiceExecutor(repos RepositoryManager) *ServiceExecutor {
	hostname, _ := os.Hostname()

	return &ServiceExecutor{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		grpcConns:    make(map[string]*grpc.ClientConn),
		repositories: repos,
		hostname:     hostname,
	}
}

// ExecuteJob executes a job by calling the target service
func (e *ServiceExecutor) ExecuteJob(ctx context.Context, job *models.Job) error {
	// Create job execution record
	execution := &models.JobExecution{
		ID:             uuid.New(),
		JobID:          job.ID,
		Status:         "running",
		StartedAt:      time.Now(),
		RequestPayload: job.TargetPayload,
		TriggeredBy:    "scheduler",
		HostName:       e.hostname,
		RetryAttempt:   0,
		CreatedAt:      time.Now(),
	}

	// Save execution record
	execution, err := e.repositories.CreateJobExecution(ctx, execution)
	if err != nil {
		log.Printf("Failed to create job execution record: %v", err)
		return err
	}

	// Execute the job with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(job.TimeoutSeconds)*time.Second)
	defer cancel()

	// Execute based on target service
	var execErr error
	switch job.TargetService {
	case "insightviewer":
		execErr = e.executeInsightViewerJob(ctx, job, execution)
	case "masters":
		execErr = e.executeMastersJob(ctx, job, execution)
	default:
		execErr = fmt.Errorf("unknown target service: %s", job.TargetService)
	}

	// Update execution status
	now := time.Now()
	execution.CompletedAt = &now
	durationMs := int32(now.Sub(execution.StartedAt).Milliseconds())
	execution.DurationMs = &durationMs

	if execErr != nil {
		execution.Status = "failed"
		errMsg := execErr.Error()
		execution.ErrorMessage = &errMsg
		log.Printf("Job %s execution failed: %v", job.Name, execErr)
	} else {
		execution.Status = "success"
		log.Printf("Job %s executed successfully in %dms", job.Name, durationMs)
	}

	// Update execution record
	_, err = e.repositories.UpdateJobExecution(ctx, execution)
	if err != nil {
		log.Printf("Failed to update job execution record: %v", err)
	}

	// Update job statistics
	err = e.repositories.UpdateJobStats(ctx, job.ID, execution.Status)
	if err != nil {
		log.Printf("Failed to update job stats: %v", err)
	}

	// Send notifications if configured
	if (execution.Status == "success" && job.NotifyOnSuccess) ||
		(execution.Status == "failed" && job.NotifyOnFailure) {
		go e.sendJobNotification(job, execution)
	}

	return execErr
}

// executeInsightViewerJob executes a job targeting InsightViewer service
func (e *ServiceExecutor) executeInsightViewerJob(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	switch job.TargetMethod {
	case "ExecuteReport":
		return e.executeReportJob(ctx, job, execution)
	case "CleanupCache":
		return e.cleanupCacheJob(ctx, job, execution)
	default:
		return fmt.Errorf("unknown InsightViewer method: %s", job.TargetMethod)
	}
}

// executeReportJob executes a scheduled report
func (e *ServiceExecutor) executeReportJob(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	// Extract report parameters
	reportID, ok := job.TargetPayload["report_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid report_id in job payload")
	}

	userID, _ := job.TargetPayload["user_id"].(string)
	if userID == "" {
		userID = job.CreatedBy
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"report_id":    reportID,
		"user_id":      userID,
		"triggered_by": "scheduler",
		"parameters":   job.TargetPayload["parameters"],
	}

	// Call InsightViewer service
	return e.callHTTPService(ctx, "insightviewer", "execute-report", payload, execution)
}

// cleanupCacheJob executes cache cleanup
func (e *ServiceExecutor) cleanupCacheJob(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	payload := map[string]interface{}{
		"max_age_hours": job.TargetPayload["max_age_hours"],
	}

	return e.callHTTPService(ctx, "insightviewer", "cleanup-cache", payload, execution)
}

// executeMastersJob executes a job targeting Masters service
func (e *ServiceExecutor) executeMastersJob(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	switch job.TargetMethod {
	case "SyncMetadata":
		return e.syncMetadataJob(ctx, job, execution)
	default:
		return fmt.Errorf("unknown Masters method: %s", job.TargetMethod)
	}
}

// syncMetadataJob executes metadata synchronization
func (e *ServiceExecutor) syncMetadataJob(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	payload := job.TargetPayload
	return e.callHTTPService(ctx, "masters", "sync-metadata", payload, execution)
}

// callHTTPService makes HTTP call to target service
func (e *ServiceExecutor) callHTTPService(ctx context.Context, service, method string, payload map[string]interface{}, execution *models.JobExecution) error {
	// Build service URL (this should be configurable)
	baseURL := e.getServiceURL(service)
	url := fmt.Sprintf("%s/api/v1/%s", baseURL, method)

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "scheduler/1.0")
	req.Header.Set("X-Execution-ID", execution.ID.String())

	// Make request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Parse response (optional)
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err == nil {
		execution.ResponseData = responseData
	}

	return nil
}

// getServiceURL returns the base URL for a service
func (e *ServiceExecutor) getServiceURL(service string) string {
	// This should be loaded from configuration
	serviceURLs := map[string]string{
		"insightviewer": "http://localhost:8082",
		"masters":       "http://localhost:8080",
	}

	if url, exists := serviceURLs[service]; exists {
		return url
	}

	return fmt.Sprintf("http://localhost:8080") // fallback
}

// sendJobNotification sends notification about job execution
func (e *ServiceExecutor) sendJobNotification(job *models.Job, execution *models.JobExecution) {
	// This would call the notification service
	// For now, just log
	log.Printf("Job notification: %s - %s (Duration: %dms)",
		job.Name, execution.Status, *execution.DurationMs)

	// TODO: Call notification service
	// notificationPayload := map[string]interface{}{
	//     "template": "job_execution",
	//     "job_name": job.Name,
	//     "status": execution.Status,
	//     "duration_ms": *execution.DurationMs,
	//     "error_message": execution.ErrorMessage,
	//     "recipients": job.NotificationRecipients,
	//     "channels": job.NotificationChannels,
	// }
}

// Close closes all gRPC connections
func (e *ServiceExecutor) Close() error {
	for service, conn := range e.grpcConns {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection to %s: %v", service, err)
		}
	}
	return nil
}