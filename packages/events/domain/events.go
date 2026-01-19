package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of domain event
type EventType string

const (
	// Workflow Events
	EventTypeWorkflowTransition EventType = "workflow.transition"
	EventTypeWorkflowCreated    EventType = "workflow.created"
	EventTypeWorkflowUpdated    EventType = "workflow.updated"
	EventTypeWorkflowDeleted    EventType = "workflow.deleted"

	// SLA Events
	EventTypeSLABreach     EventType = "sla.breach"
	EventTypeSLAWarning    EventType = "sla.warning"
	EventTypeSLAEscalation EventType = "sla.escalation"
	EventTypeSLACompliance EventType = "sla.compliance"
	EventTypeSLACreated    EventType = "sla.created"
	EventTypeSLAUpdated    EventType = "sla.updated"

	// Form Events
	EventTypeFormSubmission EventType = "form.submission"
	EventTypeFormApproval   EventType = "form.approval"
	EventTypeFormRejection  EventType = "form.rejection"
	EventTypeFormCreated    EventType = "form.created"
	EventTypeFormUpdated    EventType = "form.updated"
	EventTypeFormDeleted    EventType = "form.deleted"

	// Notification Events
	EventTypeNotificationSent      EventType = "notification.sent"
	EventTypeNotificationDelivered EventType = "notification.delivered"
	EventTypeNotificationFailed    EventType = "notification.failed"
	EventTypeNotificationRead      EventType = "notification.read"

	// Monitoring Events
	EventTypeMetricRecorded    EventType = "monitoring.metric.recorded"
	EventTypeAlertTriggered    EventType = "monitoring.alert.triggered"
	EventTypeHealthCheckFailed EventType = "monitoring.healthcheck.failed"
	EventTypePerformanceIssue  EventType = "monitoring.performance.issue"

	// Identity Events
	EventTypeUserCreated     EventType = "identity.user.created"
	EventTypeUserUpdated     EventType = "identity.user.updated"
	EventTypeUserDeactivated EventType = "identity.user.deactivated"
	EventTypeRoleAssigned    EventType = "identity.role.assigned"
	EventTypeRoleRevoked     EventType = "identity.role.revoked"

	// Tenant Events
	EventTypeTenantCreated    EventType = "identity.tenant.created"
	EventTypeTenantUpdated    EventType = "identity.tenant.updated"
	EventTypeTenantDeactivated EventType = "identity.tenant.deactivated"
	EventTypeTenantUserAdded   EventType = "identity.tenant.user.added"
	EventTypeTenantUserRemoved EventType = "identity.tenant.user.removed"

	// Masters Module Events
	EventTypeSchemaCreated EventType = "masters.schema.created"
	EventTypeSchemaUpdated EventType = "masters.schema.updated"
	EventTypeSchemaDeleted EventType = "masters.schema.deleted"
	EventTypeTableCreated  EventType = "masters.table.created"
	EventTypeTableUpdated  EventType = "masters.table.updated"
	EventTypeTableDeleted  EventType = "masters.table.deleted"
	EventTypeColumnCreated EventType = "masters.column.created"
	EventTypeColumnUpdated EventType = "masters.column.updated"
	EventTypeColumnDeleted EventType = "masters.column.deleted"

	// DataBridge Module Events
	EventTypeMappingCreated   EventType = "databridge.mapping.created"
	EventTypeMappingUpdated   EventType = "databridge.mapping.updated"
	EventTypeDataImported     EventType = "databridge.data.imported"
	EventTypeImportFailed     EventType = "databridge.import.failed"
	EventTypeImportJobStarted EventType = "databridge.import.started"

	// System Events
	EventTypeSystemStartup        EventType = "system.startup"
	EventTypeSystemShutdown       EventType = "system.shutdown"
	EventTypeConfigurationChanged EventType = "system.config.changed"
)

// Priority represents the priority level of an event
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// DomainEvent represents a domain event that occurred in the system
type DomainEvent struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]string      `json:"metadata"`
	Priority      Priority               `json:"priority"`
	Source        string                 `json:"source"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType EventType, aggregateID, aggregateType string, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:            uuid.New().String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       1,
		Data:          data,
		Metadata:      make(map[string]string),
		Priority:      PriorityMedium,
		Timestamp:     time.Now(),
	}
}

// WithVersion sets the version of the event
func (e *DomainEvent) WithVersion(version int64) *DomainEvent {
	e.Version = version
	return e
}

// WithPriority sets the priority of the event
func (e *DomainEvent) WithPriority(priority Priority) *DomainEvent {
	e.Priority = priority
	return e
}

// WithSource sets the source service of the event
func (e *DomainEvent) WithSource(source string) *DomainEvent {
	e.Source = source
	return e
}

// WithCorrelationID sets the correlation ID for tracing
func (e *DomainEvent) WithCorrelationID(correlationID string) *DomainEvent {
	e.CorrelationID = correlationID
	return e
}

// WithCausationID sets the causation ID for event causality
func (e *DomainEvent) WithCausationID(causationID string) *DomainEvent {
	e.CausationID = causationID
	return e
}

// WithMetadata adds metadata to the event
func (e *DomainEvent) WithMetadata(key, value string) *DomainEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// GetTopic returns the Kafka topic for this event type
func (e *DomainEvent) GetTopic() string {
	switch e.Type {
	case EventTypeWorkflowTransition, EventTypeWorkflowCreated, EventTypeWorkflowUpdated, EventTypeWorkflowDeleted:
		return "ugcl.workflow.events"
	case EventTypeSLABreach, EventTypeSLAWarning, EventTypeSLAEscalation, EventTypeSLACompliance, EventTypeSLACreated, EventTypeSLAUpdated:
		return "ugcl.sla.events"
	case EventTypeFormSubmission, EventTypeFormApproval, EventTypeFormRejection, EventTypeFormCreated, EventTypeFormUpdated, EventTypeFormDeleted:
		return "ugcl.form.events"
	case EventTypeNotificationSent, EventTypeNotificationDelivered, EventTypeNotificationFailed, EventTypeNotificationRead:
		return "ugcl.notification.events"
	case EventTypeMetricRecorded, EventTypeAlertTriggered, EventTypeHealthCheckFailed, EventTypePerformanceIssue:
		return "ugcl.monitoring.events"
	case EventTypeUserCreated, EventTypeUserUpdated, EventTypeUserDeactivated, EventTypeRoleAssigned, EventTypeRoleRevoked:
		return "ugcl.identity.events"
	case EventTypeTenantCreated, EventTypeTenantUpdated, EventTypeTenantDeactivated, EventTypeTenantUserAdded, EventTypeTenantUserRemoved:
		return "ugcl.tenant.events"
	case EventTypeSchemaCreated, EventTypeSchemaUpdated, EventTypeSchemaDeleted, EventTypeTableCreated, EventTypeTableUpdated, EventTypeTableDeleted, EventTypeColumnCreated, EventTypeColumnUpdated, EventTypeColumnDeleted:
		return "ugcl.masters.events"
	case EventTypeMappingCreated, EventTypeMappingUpdated, EventTypeDataImported, EventTypeImportFailed, EventTypeImportJobStarted:
		return "ugcl.databridge.events"
	case EventTypeSystemStartup, EventTypeSystemShutdown, EventTypeConfigurationChanged:
		return "ugcl.system.events"
	default:
		return "ugcl.domain.events"
	}
}

// WorkflowTransitionEvent represents a workflow state transition
type WorkflowTransitionEvent struct {
	FormInstanceID string     `json:"form_instance_id"`
	FormID         string     `json:"form_id"`
	FromState      string     `json:"from_state"`
	ToState        string     `json:"to_state"`
	TransitionBy   string     `json:"transition_by"`
	AssignedTo     string     `json:"assigned_to"`
	AssignedRole   string     `json:"assigned_role"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	Reason         string     `json:"reason,omitempty"`
}

// SLAEvent represents an SLA-related event
type SLAEvent struct {
	FormInstanceID  string     `json:"form_instance_id"`
	FormID          string     `json:"form_id"`
	SLARuleID       string     `json:"sla_rule_id"`
	CurrentState    string     `json:"current_state"`
	DueDate         time.Time  `json:"due_date"`
	BreachTime      *time.Time `json:"breach_time,omitempty"`
	EscalationLevel int32      `json:"escalation_level"`
	EscalationTo    string     `json:"escalation_to"`
	Severity        string     `json:"severity"`
}

// FormEvent represents a form-related event
type FormEvent struct {
	FormID         string                 `json:"form_id"`
	FormInstanceID string                 `json:"form_instance_id"`
	UserID         string                 `json:"user_id"`
	Action         string                 `json:"action"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Reason         string                 `json:"reason,omitempty"`
}

// NotificationEvent represents a notification-related event
type NotificationEvent struct {
	NotificationID string    `json:"notification_id"`
	RecipientID    string    `json:"recipient_id"`
	Channel        string    `json:"channel"`
	Status         string    `json:"status"`
	DeliveredAt    time.Time `json:"delivered_at,omitempty"`
	ReadAt         time.Time `json:"read_at,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// MonitoringEvent represents a monitoring-related event
type MonitoringEvent struct {
	MetricName   string                 `json:"metric_name"`
	MetricValue  float64                `json:"metric_value"`
	ServiceName  string                 `json:"service_name"`
	InstanceID   string                 `json:"instance_id"`
	Threshold    float64                `json:"threshold,omitempty"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Measurements map[string]interface{} `json:"measurements,omitempty"`
}

// IdentityEvent represents an identity-related event
type IdentityEvent struct {
	UserID      string            `json:"user_id"`
	RoleID      string            `json:"role_id,omitempty"`
	Action      string            `json:"action"`
	Permissions []string          `json:"permissions,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	ChangedBy   string            `json:"changed_by"`
}

// TenantEvent represents a tenant-related event
type TenantEvent struct {
	TenantID    string            `json:"tenant_id"`
	TenantName  string            `json:"tenant_name,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Action      string            `json:"action"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	ChangedBy   string            `json:"changed_by"`
}

// SystemEvent represents a system-related event
type SystemEvent struct {
	ServiceName   string            `json:"service_name"`
	Version       string            `json:"version"`
	Action        string            `json:"action"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Status        string            `json:"status"`
	Message       string            `json:"message,omitempty"`
}

// EventBuilder provides a fluent interface for building domain events
type EventBuilder struct {
	event *DomainEvent
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(eventType EventType, aggregateID, aggregateType string) *EventBuilder {
	return &EventBuilder{
		event: NewDomainEvent(eventType, aggregateID, aggregateType, make(map[string]interface{})),
	}
}

// WithData adds data to the event
func (b *EventBuilder) WithData(data map[string]interface{}) *EventBuilder {
	b.event.Data = data
	return b
}

// WithPriority sets the priority of the event
func (b *EventBuilder) WithPriority(priority Priority) *EventBuilder {
	b.event.Priority = priority
	return b
}

// WithSource sets the source service of the event
func (b *EventBuilder) WithSource(source string) *EventBuilder {
	b.event.Source = source
	return b
}

// WithCorrelationID sets the correlation ID for tracing
func (b *EventBuilder) WithCorrelationID(correlationID string) *EventBuilder {
	b.event.CorrelationID = correlationID
	return b
}

// WithCausationID sets the causation ID for event causality
func (b *EventBuilder) WithCausationID(causationID string) *EventBuilder {
	b.event.CausationID = causationID
	return b
}

// WithMetadata adds metadata to the event
func (b *EventBuilder) WithMetadata(key, value string) *EventBuilder {
	if b.event.Metadata == nil {
		b.event.Metadata = make(map[string]string)
	}
	b.event.Metadata[key] = value
	return b
}

// WithWorkflowTransition adds workflow transition data
func (b *EventBuilder) WithWorkflowTransition(wte *WorkflowTransitionEvent) *EventBuilder {
	data := map[string]interface{}{
		"form_instance_id": wte.FormInstanceID,
		"form_id":          wte.FormID,
		"from_state":       wte.FromState,
		"to_state":         wte.ToState,
		"transition_by":    wte.TransitionBy,
		"assigned_to":      wte.AssignedTo,
		"assigned_role":    wte.AssignedRole,
		"reason":           wte.Reason,
	}
	if wte.DueDate != nil {
		data["due_date"] = wte.DueDate
	}
	b.event.Data = data
	return b
}

// WithSLA adds SLA event data
func (b *EventBuilder) WithSLA(sla *SLAEvent) *EventBuilder {
	data := map[string]interface{}{
		"form_instance_id": sla.FormInstanceID,
		"form_id":          sla.FormID,
		"sla_rule_id":      sla.SLARuleID,
		"current_state":    sla.CurrentState,
		"due_date":         sla.DueDate,
		"escalation_level": sla.EscalationLevel,
		"escalation_to":    sla.EscalationTo,
		"severity":         sla.Severity,
	}
	if sla.BreachTime != nil {
		data["breach_time"] = sla.BreachTime
	}
	b.event.Data = data
	return b
}

// WithForm adds form event data
func (b *EventBuilder) WithForm(fe *FormEvent) *EventBuilder {
	data := map[string]interface{}{
		"form_id":          fe.FormID,
		"form_instance_id": fe.FormInstanceID,
		"user_id":          fe.UserID,
		"action":           fe.Action,
		"reason":           fe.Reason,
	}
	if fe.Data != nil {
		data["form_data"] = fe.Data
	}
	b.event.Data = data
	return b
}

// WithNotification adds notification event data
func (b *EventBuilder) WithNotification(ne *NotificationEvent) *EventBuilder {
	data := map[string]interface{}{
		"notification_id": ne.NotificationID,
		"recipient_id":    ne.RecipientID,
		"channel":         ne.Channel,
		"status":          ne.Status,
		"error_message":   ne.ErrorMessage,
	}
	if !ne.DeliveredAt.IsZero() {
		data["delivered_at"] = ne.DeliveredAt
	}
	if !ne.ReadAt.IsZero() {
		data["read_at"] = ne.ReadAt
	}
	b.event.Data = data
	return b
}

// WithMonitoring adds monitoring event data
func (b *EventBuilder) WithMonitoring(me *MonitoringEvent) *EventBuilder {
	data := map[string]interface{}{
		"metric_name":  me.MetricName,
		"metric_value": me.MetricValue,
		"service_name": me.ServiceName,
		"instance_id":  me.InstanceID,
		"threshold":    me.Threshold,
		"severity":     me.Severity,
		"description":  me.Description,
	}
	if me.Tags != nil {
		data["tags"] = me.Tags
	}
	if me.Measurements != nil {
		data["measurements"] = me.Measurements
	}
	b.event.Data = data
	return b
}

// WithIdentity adds identity event data
func (b *EventBuilder) WithIdentity(ie *IdentityEvent) *EventBuilder {
	data := map[string]interface{}{
		"user_id":    ie.UserID,
		"role_id":    ie.RoleID,
		"action":     ie.Action,
		"changed_by": ie.ChangedBy,
	}
	if ie.Permissions != nil {
		data["permissions"] = ie.Permissions
	}
	if ie.Attributes != nil {
		data["attributes"] = ie.Attributes
	}
	b.event.Data = data
	return b
}

// WithTenant adds tenant event data
func (b *EventBuilder) WithTenant(te *TenantEvent) *EventBuilder {
	data := map[string]interface{}{
		"tenant_id":   te.TenantID,
		"tenant_name": te.TenantName,
		"user_id":     te.UserID,
		"action":      te.Action,
		"changed_by":  te.ChangedBy,
	}
	if te.Attributes != nil {
		data["attributes"] = te.Attributes
	}
	b.event.Data = data
	return b
}

// WithSystem adds system event data
func (b *EventBuilder) WithSystem(se *SystemEvent) *EventBuilder {
	data := map[string]interface{}{
		"service_name": se.ServiceName,
		"version":      se.Version,
		"action":       se.Action,
		"status":       se.Status,
		"message":      se.Message,
	}
	if se.Configuration != nil {
		data["configuration"] = se.Configuration
	}
	b.event.Data = data
	return b
}

// Build returns the built domain event
func (b *EventBuilder) Build() *DomainEvent {
	return b.event
}
