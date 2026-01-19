package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"p9e.in/ugcl/notification/api/v2/notification"
	"p9e.in/ugcl/notification/services"
	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

// NotificationEventHandler handles domain events and creates notifications
type NotificationEventHandler struct {
	svc       services.INotificationService
	publisher *domain.DomainEventPublisher
	logger    p9log.Helper
}

// NewNotificationEventHandler creates a new domain event handler
func NewNotificationEventHandler(svc services.INotificationService, publisher *domain.DomainEventPublisher, logger p9log.Logger) *NotificationEventHandler {
	return &NotificationEventHandler{
		svc:       svc,
		publisher: publisher,
		logger:    *p9log.NewHelper(p9log.With(logger, "component", "NotificationEventHandler")),
	}
}

// HandleDomainEvent processes domain events and creates appropriate notifications
func (h *NotificationEventHandler) HandleDomainEvent(ctx context.Context, event *domain.DomainEvent) error {
	h.logger.Infof("Processing domain event %s of type %s", event.ID, event.Type)

	switch event.Type {
	case domain.EventTypeWorkflowTransition:
		return h.handleWorkflowTransition(ctx, event)
	case domain.EventTypeSLABreach:
		return h.handleSLABreach(ctx, event)
	case domain.EventTypeSLAWarning:
		return h.handleSLAWarning(ctx, event)
	case domain.EventTypeSLAEscalation:
		return h.handleSLAEscalation(ctx, event)
	case domain.EventTypeFormSubmission:
		return h.handleFormSubmission(ctx, event)
	case domain.EventTypeFormApproval:
		return h.handleFormApproval(ctx, event)
	case domain.EventTypeFormRejection:
		return h.handleFormRejection(ctx, event)
	default:
		h.logger.Infof("No notification handler for event type %s", event.Type)
		return nil
	}
}

// handleWorkflowTransition handles workflow transition events
func (h *NotificationEventHandler) handleWorkflowTransition(ctx context.Context, event *domain.DomainEvent) error {
	var workflowData domain.WorkflowTransitionEvent
	if err := h.mapEventData(event.Data, &workflowData); err != nil {
		return fmt.Errorf("failed to map workflow transition data: %w", err)
	}

	// Create notification for assigned user
	if workflowData.AssignedTo != "" {
		n := h.createWorkflowTransitionNotification(event, &workflowData, workflowData.AssignedTo, "user")
		if err := h.createNotification(ctx, n); err != nil {
			return fmt.Errorf("failed to create notification for assigned user: %w", err)
		}
	}

	// Create notification for assigned role
	if workflowData.AssignedRole != "" {
		n := h.createWorkflowTransitionNotification(event, &workflowData, workflowData.AssignedRole, "role")
		if err := h.createNotification(ctx, n); err != nil {
			return fmt.Errorf("failed to create notification for assigned role: %w", err)
		}
	}

	return nil
}

// handleSLABreach handles SLA breach events
func (h *NotificationEventHandler) handleSLABreach(ctx context.Context, event *domain.DomainEvent) error {
	var slaData domain.SLAEvent
	if err := h.mapEventData(event.Data, &slaData); err != nil {
		return fmt.Errorf("failed to map SLA breach data: %w", err)
	}

	n := h.createSLANotification(event, &slaData, notification.NotificationType_SLA_BREACH, "SLA Breach Alert")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create SLA breach notification: %w", err)
	}

	return nil
}

// handleSLAWarning handles SLA warning events
func (h *NotificationEventHandler) handleSLAWarning(ctx context.Context, event *domain.DomainEvent) error {
	var slaData domain.SLAEvent
	if err := h.mapEventData(event.Data, &slaData); err != nil {
		return fmt.Errorf("failed to map SLA warning data: %w", err)
	}

	n := h.createSLANotification(event, &slaData, notification.NotificationType_SLA_WARNING, "SLA Warning")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create SLA warning notification: %w", err)
	}

	return nil
}

// handleSLAEscalation handles SLA escalation events
func (h *NotificationEventHandler) handleSLAEscalation(ctx context.Context, event *domain.DomainEvent) error {
	var slaData domain.SLAEvent
	if err := h.mapEventData(event.Data, &slaData); err != nil {
		return fmt.Errorf("failed to map SLA escalation data: %w", err)
	}

	n := h.createSLANotification(event, &slaData, notification.NotificationType_SLA_ESCALATION, "SLA Escalation")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create SLA escalation notification: %w", err)
	}

	return nil
}

// handleFormSubmission handles form submission events
func (h *NotificationEventHandler) handleFormSubmission(ctx context.Context, event *domain.DomainEvent) error {
	var formData domain.FormEvent
	if err := h.mapEventData(event.Data, &formData); err != nil {
		return fmt.Errorf("failed to map form submission data: %w", err)
	}

	n := h.createFormNotification(event, &formData, notification.NotificationType_FORM_SUBMISSION, "New Form Submission")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create form submission notification: %w", err)
	}

	return nil
}

// handleFormApproval handles form approval events
func (h *NotificationEventHandler) handleFormApproval(ctx context.Context, event *domain.DomainEvent) error {
	var formData domain.FormEvent
	if err := h.mapEventData(event.Data, &formData); err != nil {
		return fmt.Errorf("failed to map form approval data: %w", err)
	}

	n := h.createFormNotification(event, &formData, notification.NotificationType_FORM_SUBMISSION, "Form Approved")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create form approval notification: %w", err)
	}

	return nil
}

// handleFormRejection handles form rejection events
func (h *NotificationEventHandler) handleFormRejection(ctx context.Context, event *domain.DomainEvent) error {
	var formData domain.FormEvent
	if err := h.mapEventData(event.Data, &formData); err != nil {
		return fmt.Errorf("failed to map form rejection data: %w", err)
	}

	n := h.createFormNotification(event, &formData, notification.NotificationType_FORM_SUBMISSION, "Form Rejected")
	if err := h.createNotification(ctx, n); err != nil {
		return fmt.Errorf("failed to create form rejection notification: %w", err)
	}

	return nil
}

// createWorkflowTransitionNotification creates a notification for workflow transitions
func (h *NotificationEventHandler) createWorkflowTransitionNotification(event *domain.DomainEvent, workflowData *domain.WorkflowTransitionEvent, recipientID, recipientType string) *notification.Notification {
	subject := fmt.Sprintf("Workflow State Changed: %s → %s", workflowData.FromState, workflowData.ToState)
	message := fmt.Sprintf("Form instance %s has transitioned from %s to %s state. Please review and take necessary action.",
		workflowData.FormInstanceID, workflowData.FromState, workflowData.ToState)

	if workflowData.Reason != "" {
		message += fmt.Sprintf(" Reason: %s", workflowData.Reason)
	}

	priority := h.mapDomainPriorityToNotificationPriority(event.Priority)

	return &notification.Notification{
		Id:            uuid.New().String(),
		Type:          notification.NotificationType_WORKFLOW_STATE_CHANGE,
		Channel:       notification.NotificationChannel_EMAIL,
		Priority:      priority,
		RecipientId:   recipientID,
		RecipientType: recipientType,
		Subject:       subject,
		Message:       message,
		SourceId:      event.AggregateID,
		SourceType:    event.AggregateType,
		Status:        notification.NotificationStatus_PENDING,
		ScheduledAt:   timestamppb.New(time.Now()),
		CreatedAt:     timestamppb.New(time.Now()),
		UpdatedAt:     timestamppb.New(time.Now()),
		Metadata: map[string]string{
			"event_id":       event.ID,
			"correlation_id": event.CorrelationID,
			"causation_id":   event.CausationID,
			"form_id":        workflowData.FormID,
			"from_state":     workflowData.FromState,
			"to_state":       workflowData.ToState,
			"transition_by":  workflowData.TransitionBy,
		},
	}
}

// createSLANotification creates a notification for SLA events
func (h *NotificationEventHandler) createSLANotification(event *domain.DomainEvent, slaData *domain.SLAEvent, notificationType notification.NotificationType, subjectPrefix string) *notification.Notification {
	var message string

	switch event.Type {
	case domain.EventTypeSLABreach:
		message = fmt.Sprintf("SLA breach detected for form instance %s. Due date was %s.",
			slaData.FormInstanceID, slaData.DueDate.Format("2006-01-02 15:04:05"))
	case domain.EventTypeSLAWarning:
		message = fmt.Sprintf("SLA warning for form instance %s. Due date is %s.",
			slaData.FormInstanceID, slaData.DueDate.Format("2006-01-02 15:04:05"))
	case domain.EventTypeSLAEscalation:
		message = fmt.Sprintf("SLA escalation (level %d) for form instance %s. Immediate action required.",
			slaData.EscalationLevel, slaData.FormInstanceID)
	}

	priority := h.mapDomainPriorityToNotificationPriority(event.Priority)

	return &notification.Notification{
		Id:            uuid.New().String(),
		Type:          notificationType,
		Channel:       notification.NotificationChannel_EMAIL,
		Priority:      priority,
		RecipientId:   slaData.EscalationTo,
		RecipientType: "user",
		Subject:       subjectPrefix,
		Message:       message,
		SourceId:      event.AggregateID,
		SourceType:    event.AggregateType,
		Status:        notification.NotificationStatus_PENDING,
		ScheduledAt:   timestamppb.New(time.Now()),
		CreatedAt:     timestamppb.New(time.Now()),
		UpdatedAt:     timestamppb.New(time.Now()),
		Metadata: map[string]string{
			"event_id":         event.ID,
			"correlation_id":   event.CorrelationID,
			"causation_id":     event.CausationID,
			"form_id":          slaData.FormID,
			"sla_rule_id":      slaData.SLARuleID,
			"current_state":    slaData.CurrentState,
			"escalation_level": fmt.Sprintf("%d", slaData.EscalationLevel),
			"severity":         slaData.Severity,
		},
	}
}

// createFormNotification creates a notification for form events
func (h *NotificationEventHandler) createFormNotification(event *domain.DomainEvent, formData *domain.FormEvent, notificationType notification.NotificationType, subjectPrefix string) *notification.Notification {
	message := fmt.Sprintf("Form %s (instance: %s) - %s", formData.FormID, formData.FormInstanceID, formData.Action)

	if formData.Reason != "" {
		message += fmt.Sprintf(". Reason: %s", formData.Reason)
	}

	priority := h.mapDomainPriorityToNotificationPriority(event.Priority)

	return &notification.Notification{
		Id:            uuid.New().String(),
		Type:          notificationType,
		Channel:       notification.NotificationChannel_EMAIL,
		Priority:      priority,
		RecipientId:   formData.UserID, // This should be determined based on business rules
		RecipientType: "user",
		Subject:       subjectPrefix,
		Message:       message,
		SourceId:      event.AggregateID,
		SourceType:    event.AggregateType,
		Status:        notification.NotificationStatus_PENDING,
		ScheduledAt:   timestamppb.New(time.Now()),
		CreatedAt:     timestamppb.New(time.Now()),
		UpdatedAt:     timestamppb.New(time.Now()),
		Metadata: map[string]string{
			"event_id":       event.ID,
			"correlation_id": event.CorrelationID,
			"causation_id":   event.CausationID,
			"form_id":        formData.FormID,
			"action":         formData.Action,
			"user_id":        formData.UserID,
		},
	}
}

// createNotification creates a notification via service layer and publishes notification events
func (h *NotificationEventHandler) createNotification(ctx context.Context, notif *notification.Notification) error {
	resp, err := h.svc.SendNotification(ctx, &notification.SendNotificationRequest{Notification: notif})
	if err != nil {
		return fmt.Errorf("failed to send notification via service: %w", err)
	}

	// Publish notification created/sent event (best-effort)
	if h.publisher != nil {
		nEvent := &domain.NotificationEvent{
			NotificationID: resp.NotificationId,
			RecipientID:    notif.RecipientId,
			Channel:        notif.Channel.String(),
			Status:         notif.Status.String(),
		}
		if err := h.publisher.PublishNotificationSent(ctx, nEvent); err != nil {
			h.logger.Warnf("Failed to publish notification sent event: %v", err)
			// Don't fail the flow on publish error
		}
	}

	h.logger.Infof("Created notification %s for recipient %s", resp.NotificationId, notif.RecipientId)
	return nil
}

// mapEventData maps event data to a specific struct
func (h *NotificationEventHandler) mapEventData(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// mapDomainPriorityToNotificationPriority maps domain event priority to notification priority
func (h *NotificationEventHandler) mapDomainPriorityToNotificationPriority(priority domain.Priority) notification.NotificationPriority {
	switch priority {
	case domain.PriorityLow:
		return notification.NotificationPriority_LOW
	case domain.PriorityMedium:
		return notification.NotificationPriority_NORMAL
	case domain.PriorityHigh:
		return notification.NotificationPriority_HIGH
	case domain.PriorityCritical:
		return notification.NotificationPriority_CRITICAL
	default:
		return notification.NotificationPriority_NORMAL
	}
}

// NotificationEventSubscriber sets up subscriptions to domain events
type NotificationEventSubscriber struct {
	subscriber *domain.EventSubscriber
	handler    *NotificationEventHandler
	logger     p9log.Helper
}

// NewNotificationEventSubscriber creates a new notification event subscriber
func NewNotificationEventSubscriber(subscriber *domain.EventSubscriber, handler *NotificationEventHandler, logger p9log.Logger) *NotificationEventSubscriber {
	return &NotificationEventSubscriber{
		subscriber: subscriber,
		handler:    handler,
		logger:     *p9log.NewHelper(p9log.With(logger, "component", "NotificationEventSubscriber")),
	}
}

// Start starts subscribing to relevant domain events
func (s *NotificationEventSubscriber) Start(ctx context.Context) error {
	// Subscribe to workflow events
	if _, err := s.subscriber.SubscribeToWorkflowEvents(s.handler.HandleDomainEvent); err != nil {
		return fmt.Errorf("failed to subscribe to workflow events: %w", err)
	}

	// Subscribe to SLA events
	if _, err := s.subscriber.SubscribeToSLAEvents(s.handler.HandleDomainEvent); err != nil {
		return fmt.Errorf("failed to subscribe to SLA events: %w", err)
	}

	// Subscribe to form events (for form submissions, approvals, rejections)
	if _, err := s.subscriber.SubscribeToAllEvents(func(ctx context.Context, event *domain.DomainEvent) error {
		if s.isFormEvent(event.Type) {
			return s.handler.HandleDomainEvent(ctx, event)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to subscribe to form events: %w", err)
	}

	s.logger.Info("Notification event subscriber started successfully")
	return nil
}

// isFormEvent checks if the event is a form-related event that should trigger notifications
func (s *NotificationEventSubscriber) isFormEvent(eventType domain.EventType) bool {
	return eventType == domain.EventTypeFormSubmission ||
		eventType == domain.EventTypeFormApproval ||
		eventType == domain.EventTypeFormRejection
}
