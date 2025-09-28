package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"p9e.in/ugcl/searchservice/models"
	"p9e.in/ugcl/searchservice/repository"
	"p9e.in/ugcl/searchservice/services/interfaces"
)

// IndexingService handles document indexing operations
type IndexingService struct {
	esRepo     *repository.ElasticsearchRepository
	cacheRepo  *repository.CacheRepository
	logger     *zap.Logger

	// Background processing
	indexQueue    chan IndexJob
	bulkQueue     chan BulkIndexJob
	workers       int
	batchSize     int
	batchTimeout  time.Duration

	// State management
	mu            sync.RWMutex
	activeJobs    map[string]*models.IndexingJob
	indexConfigs  map[string]*models.IndexConfig
}

// IndexJob represents a single document indexing job
type IndexJob struct {
	IndexName string
	DocID     string
	Document  map[string]interface{}
	Operation string // "index", "update", "delete"
	Priority  int
}

// BulkIndexJob represents a bulk indexing job
type BulkIndexJob struct {
	Operations []interfaces.BulkOperation
	Callback   func(error)
}

// NewIndexingService creates a new indexing service
func NewIndexingService(
	esRepo *repository.ElasticsearchRepository,
	cacheRepo *repository.CacheRepository,
	logger *zap.Logger,
) *IndexingService {
	service := &IndexingService{
		esRepo:        esRepo,
		cacheRepo:     cacheRepo,
		logger:        logger,
		indexQueue:    make(chan IndexJob, 10000),
		bulkQueue:     make(chan BulkIndexJob, 1000),
		workers:       5,
		batchSize:     100,
		batchTimeout:  5 * time.Second,
		activeJobs:    make(map[string]*models.IndexingJob),
		indexConfigs:  make(map[string]*models.IndexConfig),
	}

	// Start background workers
	service.startWorkers()
	service.startBulkProcessor()

	return service
}

// IndexDocument indexes a single document
func (s *IndexingService) IndexDocument(ctx context.Context, indexName string, docID string, document map[string]interface{}) error {
	// Validate inputs
	if indexName == "" || docID == "" {
		return fmt.Errorf("index name and document ID are required")
	}

	// Ensure index exists
	if err := s.ensureIndexExists(ctx, indexName); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	// Add processing metadata
	document["indexed_at"] = time.Now()
	document["index_name"] = indexName

	// Index document directly for immediate indexing
	if err := s.esRepo.IndexDocument(ctx, indexName, docID, document); err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("doc:%s:%s", indexName, docID)
	if err := s.cacheRepo.Set(ctx, cacheKey, document, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to cache document", zap.Error(err))
	}

	s.logger.Debug("Document indexed successfully",
		zap.String("index", indexName),
		zap.String("doc_id", docID))

	return nil
}

// IndexDocuments indexes multiple documents
func (s *IndexingService) IndexDocuments(ctx context.Context, indexName string, documents []map[string]interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	// Ensure index exists
	if err := s.ensureIndexExists(ctx, indexName); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	// Prepare bulk operations
	operations := make([]interfaces.BulkOperation, 0, len(documents))

	for _, doc := range documents {
		// Document must have an ID
		docID, ok := doc["id"].(string)
		if !ok {
			s.logger.Warn("Document missing ID, skipping", zap.Any("document", doc))
			continue
		}

		// Add processing metadata
		doc["indexed_at"] = time.Now()
		doc["index_name"] = indexName

		operations = append(operations, interfaces.BulkOperation{
			Action:   "index",
			Index:    indexName,
			ID:       docID,
			Document: doc,
		})
	}

	// Perform bulk indexing
	return s.BulkIndex(ctx, operations)
}

// UpdateDocument updates a document
func (s *IndexingService) UpdateDocument(ctx context.Context, indexName string, docID string, updates map[string]interface{}) error {
	// Validate inputs
	if indexName == "" || docID == "" {
		return fmt.Errorf("index name and document ID are required")
	}

	// Add processing metadata
	updates["updated_at"] = time.Now()

	// Update document
	if err := s.esRepo.UpdateDocument(ctx, indexName, docID, updates); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("doc:%s:%s", indexName, docID)
	if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate document cache", zap.Error(err))
	}

	s.logger.Debug("Document updated successfully",
		zap.String("index", indexName),
		zap.String("doc_id", docID))

	return nil
}

// DeleteDocument deletes a document
func (s *IndexingService) DeleteDocument(ctx context.Context, indexName string, docID string) error {
	// Validate inputs
	if indexName == "" || docID == "" {
		return fmt.Errorf("index name and document ID are required")
	}

	// Delete document
	if err := s.esRepo.DeleteDocument(ctx, indexName, docID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("doc:%s:%s", indexName, docID)
	if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to remove document from cache", zap.Error(err))
	}

	s.logger.Debug("Document deleted successfully",
		zap.String("index", indexName),
		zap.String("doc_id", docID))

	return nil
}

// BulkIndex performs bulk indexing operations
func (s *IndexingService) BulkIndex(ctx context.Context, operations []interfaces.BulkOperation) error {
	if len(operations) == 0 {
		return nil
	}

	// Convert to repository operations
	repoOps := make([]repository.BulkOperation, 0, len(operations))
	for _, op := range operations {
		repoOp := repository.BulkOperation{
			Action:   op.Action,
			Index:    op.Index,
			ID:       op.ID,
			Document: op.Document,
			Updates:  op.Updates,
		}

		// Add processing metadata
		if repoOp.Document != nil {
			repoOp.Document["indexed_at"] = time.Now()
			repoOp.Document["index_name"] = op.Index
		}
		if repoOp.Updates != nil {
			repoOp.Updates["updated_at"] = time.Now()
		}

		repoOps = append(repoOps, repoOp)
	}

	// Perform bulk operation
	if err := s.esRepo.BulkIndex(ctx, repoOps); err != nil {
		return fmt.Errorf("failed to perform bulk index: %w", err)
	}

	// Update cache for successful operations
	s.updateCacheForBulkOperations(ctx, operations)

	s.logger.Info("Bulk indexing completed",
		zap.Int("operations", len(operations)))

	return nil
}

// RefreshIndex refreshes an index
func (s *IndexingService) RefreshIndex(ctx context.Context, indexName string) error {
	if err := s.esRepo.RefreshIndex(ctx, indexName); err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}

	s.logger.Debug("Index refreshed successfully", zap.String("index", indexName))
	return nil
}

// QueueDocument queues a document for background indexing
func (s *IndexingService) QueueDocument(indexName, docID string, document map[string]interface{}, operation string, priority int) {
	job := IndexJob{
		IndexName: indexName,
		DocID:     docID,
		Document:  document,
		Operation: operation,
		Priority:  priority,
	}

	select {
	case s.indexQueue <- job:
		s.logger.Debug("Document queued for indexing",
			zap.String("index", indexName),
			zap.String("doc_id", docID),
			zap.String("operation", operation))
	default:
		s.logger.Warn("Index queue is full, dropping document",
			zap.String("index", indexName),
			zap.String("doc_id", docID))
	}
}

// QueueBulkOperations queues bulk operations for background processing
func (s *IndexingService) QueueBulkOperations(operations []interfaces.BulkOperation, callback func(error)) {
	job := BulkIndexJob{
		Operations: operations,
		Callback:   callback,
	}

	select {
	case s.bulkQueue <- job:
		s.logger.Debug("Bulk operations queued", zap.Int("count", len(operations)))
	default:
		s.logger.Warn("Bulk queue is full, dropping operations", zap.Int("count", len(operations)))
		if callback != nil {
			callback(fmt.Errorf("bulk queue is full"))
		}
	}
}

// Private methods

func (s *IndexingService) ensureIndexExists(ctx context.Context, indexName string) error {
	s.mu.RLock()
	_, exists := s.indexConfigs[indexName]
	s.mu.RUnlock()

	if exists {
		return nil
	}

	// Check if index exists in Elasticsearch
	indexExists, err := s.esRepo.IndexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !indexExists {
		// Create index with default configuration
		config := s.getDefaultIndexConfig(indexName)
		if err := s.esRepo.CreateIndex(ctx, config); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		s.mu.Lock()
		s.indexConfigs[indexName] = config
		s.mu.Unlock()

		s.logger.Info("Index created with default configuration", zap.String("index", indexName))
	} else {
		// Store default config for tracking
		s.mu.Lock()
		s.indexConfigs[indexName] = s.getDefaultIndexConfig(indexName)
		s.mu.Unlock()
	}

	return nil
}

func (s *IndexingService) getDefaultIndexConfig(indexName string) *models.IndexConfig {
	return &models.IndexConfig{
		Name: indexName,
		Settings: models.IndexSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 1,
			RefreshInterval:  "1s",
			MaxResultWindow:  10000,
		},
		Mappings: models.IndexMappings{
			Properties: map[string]models.FieldMapping{
				"id": {
					Type:  "keyword",
					Index: true,
				},
				"title": {
					Type:     "text",
					Analyzer: "standard",
					Index:    true,
				},
				"content": {
					Type:     "text",
					Analyzer: "standard",
					Index:    true,
				},
				"@timestamp": {
					Type:  "date",
					Index: true,
				},
				"indexed_at": {
					Type:  "date",
					Index: true,
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
		IsActive:  true,
	}
}

func (s *IndexingService) updateCacheForBulkOperations(ctx context.Context, operations []interfaces.BulkOperation) {
	for _, op := range operations {
		switch op.Action {
		case "index":
			if op.Document != nil {
				cacheKey := fmt.Sprintf("doc:%s:%s", op.Index, op.ID)
				if err := s.cacheRepo.Set(ctx, cacheKey, op.Document, 1*time.Hour); err != nil {
					s.logger.Warn("Failed to cache document", zap.Error(err))
				}
			}
		case "delete":
			cacheKey := fmt.Sprintf("doc:%s:%s", op.Index, op.ID)
			if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
				s.logger.Warn("Failed to remove document from cache", zap.Error(err))
			}
		case "update":
			cacheKey := fmt.Sprintf("doc:%s:%s", op.Index, op.ID)
			if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
				s.logger.Warn("Failed to invalidate document cache", zap.Error(err))
			}
		}
	}
}

// Background workers

func (s *IndexingService) startWorkers() {
	for i := 0; i < s.workers; i++ {
		go s.indexWorker(i)
	}
	s.logger.Info("Started indexing workers", zap.Int("count", s.workers))
}

func (s *IndexingService) indexWorker(workerID int) {
	logger := s.logger.With(zap.Int("worker_id", workerID))
	logger.Info("Indexing worker started")

	for job := range s.indexQueue {
		ctx := context.Background()

		switch job.Operation {
		case "index":
			if err := s.IndexDocument(ctx, job.IndexName, job.DocID, job.Document); err != nil {
				logger.Error("Failed to index document",
					zap.String("index", job.IndexName),
					zap.String("doc_id", job.DocID),
					zap.Error(err))
			}
		case "update":
			if err := s.UpdateDocument(ctx, job.IndexName, job.DocID, job.Document); err != nil {
				logger.Error("Failed to update document",
					zap.String("index", job.IndexName),
					zap.String("doc_id", job.DocID),
					zap.Error(err))
			}
		case "delete":
			if err := s.DeleteDocument(ctx, job.IndexName, job.DocID); err != nil {
				logger.Error("Failed to delete document",
					zap.String("index", job.IndexName),
					zap.String("doc_id", job.DocID),
					zap.Error(err))
			}
		}
	}

	logger.Info("Indexing worker stopped")
}

func (s *IndexingService) startBulkProcessor() {
	go s.bulkProcessor()
	s.logger.Info("Started bulk processor")
}

func (s *IndexingService) bulkProcessor() {
	ticker := time.NewTicker(s.batchTimeout)
	defer ticker.Stop()

	var batch []BulkIndexJob

	for {
		select {
		case job := <-s.bulkQueue:
			batch = append(batch, job)

			// Process batch if it's full
			if len(batch) >= s.batchSize {
				s.processBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			// Process batch on timeout if it has items
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (s *IndexingService) processBatch(batch []BulkIndexJob) {
	ctx := context.Background()

	for _, job := range batch {
		if err := s.BulkIndex(ctx, job.Operations); err != nil {
			s.logger.Error("Failed to process bulk operations",
				zap.Int("operations", len(job.Operations)),
				zap.Error(err))

			if job.Callback != nil {
				job.Callback(err)
			}
		} else {
			if job.Callback != nil {
				job.Callback(nil)
			}
		}
	}

	s.logger.Debug("Processed bulk batch", zap.Int("jobs", len(batch)))
}