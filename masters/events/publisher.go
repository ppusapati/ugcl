// =============================================================================
// masters/events/publisher.go - Masters module event publishers
// =============================================================================
package events

import (
	"context"
	"encoding/json"

	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"

	"github.com/google/uuid"
)

// MastersEventPublisher handles publishing masters-related events
type MastersEventPublisher struct {
	publisher *domain.DomainEventPublisher
	logger    p9log.Helper
}

// NewMastersEventPublisher creates a new masters event publisher
func NewMastersEventPublisher(publisher *domain.DomainEventPublisher, logger p9log.Logger) *MastersEventPublisher {
	return &MastersEventPublisher{
		publisher: publisher,
		logger:    *p9log.NewHelper(p9log.With(logger, "component", "MastersEventPublisher")),
	}
}

// SchemaEventData represents schema event data
type SchemaEventData struct {
	ID          string `json:"id"`
	SchemaName  string `json:"schema_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedBy   string `json:"created_by,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
}

// TableEventData represents table event data
type TableEventData struct {
	ID             string `json:"id"`
	SchemaID       string `json:"schema_id"`
	TableName      string `json:"table_name"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description,omitempty"`
	IsActive       bool   `json:"is_active"`
	SupportsImport bool   `json:"supports_import"`
	CreatedBy      string `json:"created_by,omitempty"`
	UpdatedBy      string `json:"updated_by,omitempty"`
}

// ColumnEventData represents column event data
type ColumnEventData struct {
	ID              string   `json:"id"`
	TableID         string   `json:"table_id"`
	ColumnName      string   `json:"column_name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description,omitempty"`
	DataType        string   `json:"data_type"`
	MaxLength       *int32   `json:"max_length,omitempty"`
	IsRequired      bool     `json:"is_required"`
	IsPrimaryKey    bool     `json:"is_primary_key"`
	IsUnique        bool     `json:"is_unique"`
	DefaultValue    string   `json:"default_value,omitempty"`
	ValidationRules string   `json:"validation_rules,omitempty"`
	ExampleValues   []string `json:"example_values,omitempty"`
	IsActive        bool     `json:"is_active"`
	SupportsImport  bool     `json:"supports_import"`
	ColumnOrder     int32    `json:"column_order"`
	CreatedBy       string   `json:"created_by,omitempty"`
	UpdatedBy       string   `json:"updated_by,omitempty"`
}

// PublishSchemaCreated publishes a schema created event
func (p *MastersEventPublisher) PublishSchemaCreated(ctx context.Context, schema *SchemaEventData) error {
	data, err := p.structToMap(schema)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeSchemaCreated, schema.ID, "schema").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing schema created event for schema %s", schema.SchemaName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishSchemaUpdated publishes a schema updated event
func (p *MastersEventPublisher) PublishSchemaUpdated(ctx context.Context, schema *SchemaEventData) error {
	data, err := p.structToMap(schema)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeSchemaUpdated, schema.ID, "schema").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing schema updated event for schema %s", schema.SchemaName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishSchemaDeleted publishes a schema deleted event
func (p *MastersEventPublisher) PublishSchemaDeleted(ctx context.Context, schemaID string, schemaName string, deletedBy string) error {
	data := map[string]interface{}{
		"id":          schemaID,
		"schema_name": schemaName,
		"deleted_by":  deletedBy,
		"is_active":   false,
	}

	event := domain.NewEventBuilder(domain.EventTypeSchemaDeleted, schemaID, "schema").
		WithData(data).
		WithPriority(domain.PriorityHigh).
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing schema deleted event for schema %s", schemaName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishTableCreated publishes a table created event
func (p *MastersEventPublisher) PublishTableCreated(ctx context.Context, table *TableEventData) error {
	data, err := p.structToMap(table)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeTableCreated, table.ID, "table").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing table created event for table %s", table.TableName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishTableUpdated publishes a table updated event
func (p *MastersEventPublisher) PublishTableUpdated(ctx context.Context, table *TableEventData) error {
	data, err := p.structToMap(table)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeTableUpdated, table.ID, "table").
		WithData(data).
		WithPriority(domain.PriorityHigh). // High priority for databridge cache invalidation
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing table updated event for table %s", table.TableName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishTableDeleted publishes a table deleted event
func (p *MastersEventPublisher) PublishTableDeleted(ctx context.Context, tableID string, tableName string, deletedBy string) error {
	data := map[string]interface{}{
		"id":         tableID,
		"table_name": tableName,
		"deleted_by": deletedBy,
		"is_active":  false,
	}

	event := domain.NewEventBuilder(domain.EventTypeTableDeleted, tableID, "table").
		WithData(data).
		WithPriority(domain.PriorityCritical). // Critical priority for databridge cache cleanup
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing table deleted event for table %s", tableName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishColumnCreated publishes a column created event
func (p *MastersEventPublisher) PublishColumnCreated(ctx context.Context, column *ColumnEventData) error {
	data, err := p.structToMap(column)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeColumnCreated, column.ID, "column").
		WithData(data).
		WithPriority(domain.PriorityMedium).
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing column created event for column %s", column.ColumnName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishColumnUpdated publishes a column updated event
func (p *MastersEventPublisher) PublishColumnUpdated(ctx context.Context, column *ColumnEventData) error {
	data, err := p.structToMap(column)
	if err != nil {
		return err
	}

	event := domain.NewEventBuilder(domain.EventTypeColumnUpdated, column.ID, "column").
		WithData(data).
		WithPriority(domain.PriorityHigh). // High priority for databridge validation cache
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing column updated event for column %s", column.ColumnName)
	return p.publisher.PublishEvent(ctx, event)
}

// PublishColumnDeleted publishes a column deleted event
func (p *MastersEventPublisher) PublishColumnDeleted(ctx context.Context, columnID string, columnName string, deletedBy string) error {
	data := map[string]interface{}{
		"id":          columnID,
		"column_name": columnName,
		"deleted_by":  deletedBy,
		"is_active":   false,
	}

	event := domain.NewEventBuilder(domain.EventTypeColumnDeleted, columnID, "column").
		WithData(data).
		WithPriority(domain.PriorityHigh). // High priority for databridge validation updates
		WithSource("masters").
		Build()

	p.logger.Infof("Publishing column deleted event for column %s", columnName)
	return p.publisher.PublishEvent(ctx, event)
}

// Helper function to convert struct to map for event data
func (p *MastersEventPublisher) structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}