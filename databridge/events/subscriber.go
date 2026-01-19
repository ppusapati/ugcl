// =============================================================================
// databridge/events/subscriber.go - DataBridge event subscriber for masters events
// =============================================================================
package events

import (
	"context"
	"encoding/json"

	"databridge/cache"

	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/events/bus"
	"p9e.in/ugcl/packages/p9log"
)

// MastersEventSubscriber handles masters events for cache invalidation
type MastersEventSubscriber struct {
	cache     *cache.MastersCache
	eventBus  *bus.EventBus
	logger    p9log.Helper
	disposers []bus.IDisposable
}

// NewMastersEventSubscriber creates a new masters event subscriber
func NewMastersEventSubscriber(cache *cache.MastersCache, eventBus *bus.EventBus, logger p9log.Logger) *MastersEventSubscriber {
	return &MastersEventSubscriber{
		cache:    cache,
		eventBus: eventBus,
		logger:   *p9log.NewHelper(p9log.With(logger, "component", "MastersEventSubscriber")),
	}
}

// Start subscribes to masters events and begins cache management
func (s *MastersEventSubscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting Masters event subscriber for cache management")

	// Subscribe to all masters events
	subscribe := bus.Subscribe[*domain.DomainEvent](s.eventBus)

	disposer, err := subscribe(func(ctx context.Context, event *domain.DomainEvent) error {
		return s.handleMastersEvent(ctx, event)
	})

	if err != nil {
		return err
	}

	s.disposers = append(s.disposers, disposer)

	s.logger.Info("Successfully subscribed to masters events")
	return nil
}

// Stop unsubscribes from events
func (s *MastersEventSubscriber) Stop() error {
	s.logger.Info("Stopping Masters event subscriber")

	for _, disposer := range s.disposers {
		if err := disposer.Dispose(); err != nil {
			s.logger.Errorf("Failed to dispose event subscription: %v", err)
		}
	}

	s.disposers = nil
	return nil
}

// handleMastersEvent processes masters events and updates cache accordingly
func (s *MastersEventSubscriber) handleMastersEvent(ctx context.Context, event *domain.DomainEvent) error {
	// Only process masters events
	if !s.isMastersEvent(event.Type) {
		return nil
	}

	s.logger.Infof("Processing masters event: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	case domain.EventTypeSchemaCreated:
		return s.handleSchemaCreated(ctx, event)
	case domain.EventTypeSchemaUpdated:
		return s.handleSchemaUpdated(ctx, event)
	case domain.EventTypeSchemaDeleted:
		return s.handleSchemaDeleted(ctx, event)
	case domain.EventTypeTableCreated:
		return s.handleTableCreated(ctx, event)
	case domain.EventTypeTableUpdated:
		return s.handleTableUpdated(ctx, event)
	case domain.EventTypeTableDeleted:
		return s.handleTableDeleted(ctx, event)
	case domain.EventTypeColumnCreated:
		return s.handleColumnCreated(ctx, event)
	case domain.EventTypeColumnUpdated:
		return s.handleColumnUpdated(ctx, event)
	case domain.EventTypeColumnDeleted:
		return s.handleColumnDeleted(ctx, event)
	default:
		s.logger.Warnf("Unknown masters event type: %s", event.Type)
		return nil
	}
}

// Schema event handlers

func (s *MastersEventSubscriber) handleSchemaCreated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Schema created: %s", event.AggregateID)
	// New schema - cache will be populated on first access
	return nil
}

func (s *MastersEventSubscriber) handleSchemaUpdated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Schema updated: %s", event.AggregateID)

	// Update cache with new data
	s.cache.UpdateSchema(event.Data)

	return nil
}

func (s *MastersEventSubscriber) handleSchemaDeleted(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Schema deleted: %s", event.AggregateID)

	// Invalidate schema and all related cached data
	s.cache.InvalidateSchema(event.AggregateID)

	return nil
}

// Table event handlers

func (s *MastersEventSubscriber) handleTableCreated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Table created: %s", event.AggregateID)
	// New table - invalidate schema's table list cache so it gets refreshed
	if schemaID, exists := event.Data["schema_id"].(string); exists {
		s.cache.InvalidateSchema(schemaID)
	}
	return nil
}

func (s *MastersEventSubscriber) handleTableUpdated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Table updated: %s", event.AggregateID)

	// Update cache with new data
	s.cache.UpdateTable(event.Data)

	// Also invalidate schema's table list cache
	if schemaID, exists := event.Data["schema_id"].(string); exists {
		s.cache.InvalidateSchema(schemaID)
	}

	return nil
}

func (s *MastersEventSubscriber) handleTableDeleted(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Table deleted: %s", event.AggregateID)

	// Invalidate table and all related cached data
	s.cache.InvalidateTable(event.AggregateID)

	// Also invalidate schema's table list cache
	if schemaID, exists := event.Data["schema_id"].(string); exists {
		s.cache.InvalidateSchema(schemaID)
	}

	return nil
}

// Column event handlers

func (s *MastersEventSubscriber) handleColumnCreated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Column created: %s", event.AggregateID)
	// New column - invalidate table's column list cache so it gets refreshed
	if tableID, exists := event.Data["table_id"].(string); exists {
		s.cache.InvalidateTable(tableID)
	}
	return nil
}

func (s *MastersEventSubscriber) handleColumnUpdated(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Column updated: %s", event.AggregateID)

	// Update cache with new data
	s.cache.UpdateColumn(event.Data)

	// Also invalidate table's column list cache
	if tableID, exists := event.Data["table_id"].(string); exists {
		s.cache.InvalidateTable(tableID)
	}

	return nil
}

func (s *MastersEventSubscriber) handleColumnDeleted(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Column deleted: %s", event.AggregateID)

	// Invalidate column and related cached data
	s.cache.InvalidateColumn(event.AggregateID)

	// Also invalidate table's column list cache
	if tableID, exists := event.Data["table_id"].(string); exists {
		s.cache.InvalidateTable(tableID)
	}

	return nil
}

// Helper methods

func (s *MastersEventSubscriber) isMastersEvent(eventType domain.EventType) bool {
	switch eventType {
	case domain.EventTypeSchemaCreated, domain.EventTypeSchemaUpdated, domain.EventTypeSchemaDeleted:
		return true
	case domain.EventTypeTableCreated, domain.EventTypeTableUpdated, domain.EventTypeTableDeleted:
		return true
	case domain.EventTypeColumnCreated, domain.EventTypeColumnUpdated, domain.EventTypeColumnDeleted:
		return true
	default:
		return false
	}
}

// DataBridgeEventPublisher handles publishing databridge events
type DataBridgeEventPublisher struct {
	publisher *domain.DomainEventPublisher
	logger    p9log.Helper
}

// NewDataBridgeEventPublisher creates a new databridge event publisher
func NewDataBridgeEventPublisher(publisher *domain.DomainEventPublisher, logger p9log.Logger) *DataBridgeEventPublisher {
	return &DataBridgeEventPublisher{
		publisher: publisher,
		logger:    *p9log.NewHelper(p9log.With(logger, "component", "DataBridgeEventPublisher")),
	}
}

// ImportEventData represents import event data
type ImportEventData struct {
	JobID          string            `json:"job_id"`
	MappingID      string            `json:"mapping_id"`
	TableID        string            `json:"table_id"`
	FileName       string            `json:"file_name"`
	TotalRows      int32             `json:"total_rows"`
	SuccessfulRows int32             `json:"successful_rows"`
	FailedRows     int32             `json:"failed_rows"`
	Status         string            `json:"status"`
	Duration       string            `json:"duration,omitempty"`
	ErrorDetails   string            `json:"error_details,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishDataImported publishes a data imported event
func (p *DataBridgeEventPublisher) PublishDataImported(ctx context.Context, importData *ImportEventData) error {
	data, err := p.structToMap(importData)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeDataImported, importData.JobID, "import_job").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("databridge").
		Build()

	p.logger.Infof("Publishing data imported event for job %s", importData.JobID)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishImportFailed publishes an import failed event
func (p *DataBridgeEventPublisher) PublishImportFailed(ctx context.Context, importData *ImportEventData) error {
	data, err := p.structToMap(importData)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeImportFailed, importData.JobID, "import_job").
		WithData(data).
		WithPriority(domain.PriorityHigh).
		WithSource("databridge").
		Build()

	p.logger.Errorf("Publishing import failed event for job %s: %s", importData.JobID, importData.ErrorDetails)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishMappingCreated publishes a mapping created event
func (p *DataBridgeEventPublisher) PublishMappingCreated(ctx context.Context, mappingID string, mappingName string, tableID string, createdBy string) error {
	data := map[string]interface{}{
		"mapping_id":   mappingID,
		"mapping_name": mappingName,
		"table_id":     tableID,
		"created_by":   createdBy,
	}

	event := domain.NewEventBuilder(domain.EventTypeMappingCreated, mappingID, "import_mapping").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("databridge").
		Build()

	p.logger.Infof("Publishing mapping created event for mapping %s", mappingName)
	return p.publisher.PublishEvent(ctx, event)
}

// Helper function to convert struct to map
func (p *DataBridgeEventPublisher) structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}