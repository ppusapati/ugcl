package events

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/documentviewer/models"
)

// Event types for document operations
const (
	DocumentCreatedEvent    = "document.created"
	DocumentUpdatedEvent    = "document.updated"
	DocumentDeletedEvent    = "document.deleted"
	DocumentProcessedEvent  = "document.processed"
	DocumentAccessedEvent   = "document.accessed"
	WatermarkAppliedEvent   = "document.watermark.applied"
	ProcessingStartedEvent  = "document.processing.started"
	ProcessingCompletedEvent = "document.processing.completed"
	ProcessingFailedEvent   = "document.processing.failed"
)

// DocumentEventPublisher publishes document-related events
type DocumentEventPublisher struct {
	publisher domain.EventPublisher
	logger    *zap.Logger
}

// NewDocumentEventPublisher creates a new document event publisher
func NewDocumentEventPublisher(publisher domain.EventPublisher, logger *zap.Logger) *DocumentEventPublisher {
	return &DocumentEventPublisher{
		publisher: publisher,
		logger:    logger,
	}
}

// DocumentCreatedEventData represents data for document created event
type DocumentCreatedEventData struct {
	DocumentID    string                 `json:"document_id"`
	FileName      string                 `json:"file_name"`
	Title         string                 `json:"title"`
	MimeType      string                 `json:"mime_type"`
	SizeBytes     int64                  `json:"size_bytes"`
	UploadedBy    string                 `json:"uploaded_by"`
	ExtractedText string                 `json:"extracted_text,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Categories    []string               `json:"categories,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	StoragePath   string                 `json:"storage_path"`
}

// DocumentUpdatedEventData represents data for document updated event
type DocumentUpdatedEventData struct {
	DocumentID    string                 `json:"document_id"`
	FileName      string                 `json:"file_name"`
	Title         string                 `json:"title"`
	UpdatedFields []string               `json:"updated_fields"`
	UpdatedBy     string                 `json:"updated_by"`
	PreviousData  map[string]interface{} `json:"previous_data,omitempty"`
	NewData       map[string]interface{} `json:"new_data,omitempty"`
}

// DocumentDeletedEventData represents data for document deleted event
type DocumentDeletedEventData struct {
	DocumentID  string `json:"document_id"`
	FileName    string `json:"file_name"`
	DeletedBy   string `json:"deleted_by"`
	HardDelete  bool   `json:"hard_delete"`
	StoragePath string `json:"storage_path"`
}

// DocumentProcessedEventData represents data for document processed event
type DocumentProcessedEventData struct {
	DocumentID      string                 `json:"document_id"`
	ProcessingType  string                 `json:"processing_type"` // "ocr", "thumbnail", "compression"
	Status          string                 `json:"status"`          // "completed", "failed"
	ExtractedText   string                 `json:"extracted_text,omitempty"`
	ThumbnailPath   string                 `json:"thumbnail_path,omitempty"`
	CompressedPath  string                 `json:"compressed_path,omitempty"`
	ProcessingTime  int64                  `json:"processing_time_ms"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentAccessedEventData represents data for document access event
type DocumentAccessedEventData struct {
	DocumentID       string                 `json:"document_id"`
	UserID           string                 `json:"user_id"`
	Action           string                 `json:"action"` // "view", "download", "print"
	IPAddress        string                 `json:"ip_address,omitempty"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	WatermarkApplied bool                   `json:"watermark_applied"`
	WatermarkConfig  string                 `json:"watermark_config_id,omitempty"`
	Duration         int64                  `json:"duration_ms"`
	Success          bool                   `json:"success"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// WatermarkAppliedEventData represents data for watermark applied event
type WatermarkAppliedEventData struct {
	DocumentID      string                 `json:"document_id"`
	WatermarkConfigID string               `json:"watermark_config_id"`
	UserID          string                 `json:"user_id"`
	Action          string                 `json:"action"` // "preview", "download", "print"
	DynamicData     map[string]interface{} `json:"dynamic_data"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	Success         bool                   `json:"success"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

// ProcessingJobEventData represents data for processing job events
type ProcessingJobEventData struct {
	JobID        string                 `json:"job_id"`
	DocumentID   string                 `json:"document_id"`
	JobType      string                 `json:"job_type"`
	Status       string                 `json:"status"`
	Progress     int                    `json:"progress_percentage"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Result       map[string]interface{} `json:"result,omitempty"`
}

// PublishDocumentCreated publishes a document created event
func (p *DocumentEventPublisher) PublishDocumentCreated(ctx context.Context, document *models.Document) error {
	eventData := &DocumentCreatedEventData{
		DocumentID:    document.ID.String(),
		FileName:      document.FileName,
		Title:         document.Title,
		MimeType:      document.MimeType,
		SizeBytes:     document.SizeBytes,
		UploadedBy:    document.UploadedBy,
		ExtractedText: document.ExtractedText,
		StoragePath:   document.StoragePath,
	}

	// Parse tags and categories from JSONB
	if document.Tags != nil {
		var tags []string
		if err := json.Unmarshal(document.Tags, &tags); err == nil {
			eventData.Tags = tags
		}
	}

	if document.Categories != nil {
		var categories []string
		if err := json.Unmarshal(document.Categories, &categories); err == nil {
			eventData.Categories = categories
		}
	}

	if document.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(document.Metadata, &metadata); err == nil {
			eventData.Metadata = metadata
		}
	}

	event := &domain.Event{
		Type:      DocumentCreatedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id": document.ID.String(),
			"mime_type":   document.MimeType,
			"uploaded_by": document.UploadedBy,
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish document created event",
			zap.String("document_id", document.ID.String()),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published document created event",
		zap.String("document_id", document.ID.String()),
		zap.String("file_name", document.FileName))

	return nil
}

// PublishDocumentUpdated publishes a document updated event
func (p *DocumentEventPublisher) PublishDocumentUpdated(ctx context.Context, documentID, fileName, updatedBy string, updatedFields []string, previousData, newData map[string]interface{}) error {
	eventData := &DocumentUpdatedEventData{
		DocumentID:    documentID,
		FileName:      fileName,
		UpdatedBy:     updatedBy,
		UpdatedFields: updatedFields,
		PreviousData:  previousData,
		NewData:       newData,
	}

	event := &domain.Event{
		Type:      DocumentUpdatedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id":     documentID,
			"updated_by":      updatedBy,
			"updated_fields":  fmt.Sprintf("%v", updatedFields),
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish document updated event",
			zap.String("document_id", documentID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published document updated event",
		zap.String("document_id", documentID),
		zap.Strings("updated_fields", updatedFields))

	return nil
}

// PublishDocumentDeleted publishes a document deleted event
func (p *DocumentEventPublisher) PublishDocumentDeleted(ctx context.Context, documentID, fileName, deletedBy, storagePath string, hardDelete bool) error {
	eventData := &DocumentDeletedEventData{
		DocumentID:  documentID,
		FileName:    fileName,
		DeletedBy:   deletedBy,
		HardDelete:  hardDelete,
		StoragePath: storagePath,
	}

	event := &domain.Event{
		Type:      DocumentDeletedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id": documentID,
			"deleted_by":  deletedBy,
			"hard_delete": fmt.Sprintf("%t", hardDelete),
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish document deleted event",
			zap.String("document_id", documentID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published document deleted event",
		zap.String("document_id", documentID),
		zap.String("deleted_by", deletedBy))

	return nil
}

// PublishDocumentProcessed publishes a document processed event
func (p *DocumentEventPublisher) PublishDocumentProcessed(ctx context.Context, job *models.ProcessingJob) error {
	eventData := &DocumentProcessedEventData{
		DocumentID:     job.DocumentID.String(),
		ProcessingType: job.JobType,
		Status:         job.Status,
		ProcessingTime: job.CompletedAt.Sub(*job.StartedAt).Milliseconds(),
		ErrorMessage:   job.ErrorMessage,
	}

	// Parse result data
	if job.Result != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(job.Result, &result); err == nil {
			eventData.Metadata = result

			// Extract specific fields based on processing type
			switch job.JobType {
			case "ocr":
				if text, ok := result["extracted_text"].(string); ok {
					eventData.ExtractedText = text
				}
			case "thumbnail":
				if path, ok := result["thumbnail_path"].(string); ok {
					eventData.ThumbnailPath = path
				}
			case "compression":
				if path, ok := result["compressed_path"].(string); ok {
					eventData.CompressedPath = path
				}
			}
		}
	}

	event := &domain.Event{
		Type:      DocumentProcessedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id":      job.DocumentID.String(),
			"processing_type":  job.JobType,
			"status":           job.Status,
			"job_id":           job.ID.String(),
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish document processed event",
			zap.String("document_id", job.DocumentID.String()),
			zap.String("job_type", job.JobType),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published document processed event",
		zap.String("document_id", job.DocumentID.String()),
		zap.String("job_type", job.JobType),
		zap.String("status", job.Status))

	return nil
}

// PublishDocumentAccessed publishes a document accessed event
func (p *DocumentEventPublisher) PublishDocumentAccessed(ctx context.Context, accessLog *models.DocumentAccessLog) error {
	eventData := &DocumentAccessedEventData{
		DocumentID:       accessLog.DocumentID.String(),
		UserID:           accessLog.UserID,
		Action:           accessLog.Action,
		IPAddress:        accessLog.IPAddress.String(),
		UserAgent:        accessLog.UserAgent,
		WatermarkApplied: accessLog.WatermarkApplied,
		Duration:         accessLog.Duration,
		Success:          accessLog.Success,
		ErrorMessage:     accessLog.ErrorMessage,
	}

	// Parse watermark data
	if accessLog.WatermarkData != nil {
		var watermarkData map[string]interface{}
		if err := json.Unmarshal(accessLog.WatermarkData, &watermarkData); err == nil {
			eventData.Metadata = watermarkData
			if configID, ok := watermarkData["config_id"].(string); ok {
				eventData.WatermarkConfig = configID
			}
		}
	}

	event := &domain.Event{
		Type:      DocumentAccessedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id": accessLog.DocumentID.String(),
			"user_id":     accessLog.UserID,
			"action":      accessLog.Action,
			"success":     fmt.Sprintf("%t", accessLog.Success),
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish document accessed event",
			zap.String("document_id", accessLog.DocumentID.String()),
			zap.String("user_id", accessLog.UserID),
			zap.Error(err))
		return err
	}

	p.logger.Debug("Published document accessed event",
		zap.String("document_id", accessLog.DocumentID.String()),
		zap.String("user_id", accessLog.UserID),
		zap.String("action", accessLog.Action))

	return nil
}

// PublishWatermarkApplied publishes a watermark applied event
func (p *DocumentEventPublisher) PublishWatermarkApplied(ctx context.Context, documentID, watermarkConfigID, userID, action string, dynamicData map[string]interface{}, success bool, errorMessage string) error {
	eventData := &WatermarkAppliedEventData{
		DocumentID:        documentID,
		WatermarkConfigID: watermarkConfigID,
		UserID:            userID,
		Action:            action,
		DynamicData:       dynamicData,
		Success:           success,
		ErrorMessage:      errorMessage,
	}

	event := &domain.Event{
		Type:      WatermarkAppliedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"document_id":          documentID,
			"watermark_config_id":  watermarkConfigID,
			"user_id":              userID,
			"action":               action,
			"success":              fmt.Sprintf("%t", success),
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish watermark applied event",
			zap.String("document_id", documentID),
			zap.String("watermark_config_id", watermarkConfigID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published watermark applied event",
		zap.String("document_id", documentID),
		zap.String("watermark_config_id", watermarkConfigID),
		zap.String("action", action))

	return nil
}

// PublishProcessingJobStarted publishes a processing job started event
func (p *DocumentEventPublisher) PublishProcessingJobStarted(ctx context.Context, job *models.ProcessingJob) error {
	eventData := &ProcessingJobEventData{
		JobID:      job.ID.String(),
		DocumentID: job.DocumentID.String(),
		JobType:    job.JobType,
		Status:     job.Status,
		Progress:   job.Progress,
		StartedAt:  job.StartedAt,
	}

	event := &domain.Event{
		Type:      ProcessingStartedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"job_id":      job.ID.String(),
			"document_id": job.DocumentID.String(),
			"job_type":    job.JobType,
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish processing started event",
			zap.String("job_id", job.ID.String()),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published processing started event",
		zap.String("job_id", job.ID.String()),
		zap.String("job_type", job.JobType))

	return nil
}

// PublishProcessingJobCompleted publishes a processing job completed event
func (p *DocumentEventPublisher) PublishProcessingJobCompleted(ctx context.Context, job *models.ProcessingJob) error {
	eventData := &ProcessingJobEventData{
		JobID:       job.ID.String(),
		DocumentID:  job.DocumentID.String(),
		JobType:     job.JobType,
		Status:      job.Status,
		Progress:    job.Progress,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}

	// Parse result data
	if job.Result != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(job.Result, &result); err == nil {
			eventData.Result = result
		}
	}

	event := &domain.Event{
		Type:      ProcessingCompletedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"job_id":      job.ID.String(),
			"document_id": job.DocumentID.String(),
			"job_type":    job.JobType,
			"status":      job.Status,
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish processing completed event",
			zap.String("job_id", job.ID.String()),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published processing completed event",
		zap.String("job_id", job.ID.String()),
		zap.String("job_type", job.JobType),
		zap.String("status", job.Status))

	return nil
}

// PublishProcessingJobFailed publishes a processing job failed event
func (p *DocumentEventPublisher) PublishProcessingJobFailed(ctx context.Context, job *models.ProcessingJob) error {
	eventData := &ProcessingJobEventData{
		JobID:        job.ID.String(),
		DocumentID:   job.DocumentID.String(),
		JobType:      job.JobType,
		Status:       job.Status,
		Progress:     job.Progress,
		StartedAt:    job.StartedAt,
		CompletedAt:  job.CompletedAt,
		ErrorMessage: job.ErrorMessage,
	}

	event := &domain.Event{
		Type:      ProcessingFailedEvent,
		Source:    "documentviewer",
		Data:      eventData,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"job_id":      job.ID.String(),
			"document_id": job.DocumentID.String(),
			"job_type":    job.JobType,
			"error":       job.ErrorMessage,
		},
	}

	if err := p.publisher.Publish(ctx, event); err != nil {
		p.logger.Error("Failed to publish processing failed event",
			zap.String("job_id", job.ID.String()),
			zap.Error(err))
		return err
	}

	p.logger.Warn("Published processing failed event",
		zap.String("job_id", job.ID.String()),
		zap.String("job_type", job.JobType),
		zap.String("error", job.ErrorMessage))

	return nil
}