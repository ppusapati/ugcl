package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/ugcl/notification/api/v2/notification"
	"p9e.in/ugcl/notification/repository"
	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

// notificationService implements NotificationService
type NotificationService struct {
	repo      repository.INotificationRepository
	publisher *domain.DomainEventPublisher
	logger    p9log.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo repository.INotificationRepository,
	publisher *domain.DomainEventPublisher,
	logger p9log.Logger,
) INotificationService {
	return &NotificationService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// SendNotification sends a single notification
func (s *NotificationService) SendNotification(ctx context.Context, req *notification.SendNotificationRequest) (*notification.SendNotificationResponse, error) {
	if req.Notification == nil {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: "notification is required",
		}, nil
	}

	// Validate notification
	if err := s.validateNotification(req.Notification); err != nil {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: fmt.Sprintf("validation failed: %v", err),
		}, nil
	}

	// Generate ID if not provided
	if req.Notification.Id == "" {
		req.Notification.Id = uuid.New().String()
	}

	// Set default values
	s.setNotificationDefaults(req.Notification)

	// Process template if template_id is provided
	if req.Notification.TemplateId != "" {
		if err := s.processTemplate(ctx, req.Notification); err != nil {
			s.logger.Log(p9log.LevelWarn, "Failed to process template %s: %v", req.Notification.TemplateId, err)
		}
	}

	// Create notification in database
	created, err := s.repo.CreateNotification(ctx, req.Notification)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to create notification: %v", err)
		return &notification.SendNotificationResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create notification: %v", err),
		}, nil
	}

	// Publish notification event
	if s.publisher != nil {
		notificationEvent := &domain.NotificationEvent{
			NotificationID: created.Id,
			RecipientID:    created.RecipientId,
			Channel:        created.Channel.String(),
			Status:         created.Status.String(),
		}

		event := domain.NewEventBuilder(domain.EventTypeNotificationSent, created.Id, "notification").
			WithNotification(notificationEvent).
			WithPriority(s.mapNotificationPriorityToDomainPriority(created.Priority)).
			WithSource("notification-service").
			Build()

		if err := s.publisher.PublishEvent(ctx, event); err != nil {
			s.logger.Log(p9log.LevelWarn, "Failed to publish notification event: %v", err)
		}
	}

	s.logger.Log(p9log.LevelInfo, "Notification %s sent successfully to %s", created.Id, created.RecipientId)

	return &notification.SendNotificationResponse{
		NotificationId: created.Id,
		Success:        true,
		Message:        "notification sent successfully",
	}, nil
}

// SendBulkNotification sends multiple notifications
func (s *NotificationService) SendBulkNotification(ctx context.Context, req *notification.SendBulkNotificationRequest) (*notification.SendBulkNotificationResponse, error) {
	if req.BulkRequest == nil {
		return &notification.SendBulkNotificationResponse{
			TotalSent:     0,
			TotalFailed:   1,
			ErrorMessages: []string{"bulk request is required"},
		}, nil
	}

	bulkReq := req.BulkRequest
	notifications := make([]*notification.Notification, 0, len(bulkReq.RecipientIds))

	// Create individual notifications for each recipient
	for _, recipientID := range bulkReq.RecipientIds {
		notif := &notification.Notification{
			Id:            uuid.New().String(),
			Type:          bulkReq.Type,
			Channel:       bulkReq.Channel,
			Priority:      bulkReq.Priority,
			RecipientId:   recipientID,
			RecipientType: bulkReq.RecipientType,
			Subject:       bulkReq.Subject,
			Message:       bulkReq.Message,
			TemplateId:    bulkReq.TemplateId,
			TemplateData:  bulkReq.TemplateData,
			ScheduledAt:   bulkReq.ScheduledAt,
			Metadata:      bulkReq.Metadata,
		}

		s.setNotificationDefaults(notif)

		// Process template if provided
		if notif.TemplateId != "" {
			if err := s.processTemplate(ctx, notif); err != nil {
				s.logger.Log(p9log.LevelWarn, "Failed to process template for recipient %s: %v", recipientID, err)
			}
		}

		notifications = append(notifications, notif)
	}

	// Create notifications in bulk
	notificationIDs, err := s.repo.CreateBulkNotifications(ctx, notifications)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to create bulk notifications: %v", err)
		return &notification.SendBulkNotificationResponse{
			TotalSent:     0,
			TotalFailed:   int32(len(bulkReq.RecipientIds)),
			ErrorMessages: []string{fmt.Sprintf("failed to create bulk notifications: %v", err)},
		}, nil
	}

	// Publish events for successful notifications
	if s.publisher != nil {
		for i, notifID := range notificationIDs {
			if i < len(notifications) {
				notif := notifications[i]
				notificationEvent := &domain.NotificationEvent{
					NotificationID: notifID,
					RecipientID:    notif.RecipientId,
					Channel:        notif.Channel.String(),
					Status:         notif.Status.String(),
				}

				event := domain.NewEventBuilder(domain.EventTypeNotificationSent, notifID, "notification").
					WithNotification(notificationEvent).
					WithPriority(s.mapNotificationPriorityToDomainPriority(notif.Priority)).
					WithSource("notification-service").
					Build()

				if err := s.publisher.PublishEvent(ctx, event); err != nil {
					s.logger.Log(p9log.LevelWarn, "Failed to publish notification event for %s: %v", notifID, err)
				}
			}
		}
	}

	s.logger.Log(p9log.LevelInfo, "Bulk notification sent to %d recipients", len(notificationIDs))

	return &notification.SendBulkNotificationResponse{
		NotificationIds: notificationIDs,
		TotalSent:       int32(len(notificationIDs)),
		TotalFailed:     0,
		ErrorMessages:   []string{},
	}, nil
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, req *notification.GetNotificationRequest) (*notification.GetNotificationResponse, error) {
	if req.NotificationId == "" {
		return nil, fmt.Errorf("notification ID is required")
	}

	notif, err := s.repo.GetNotification(ctx, req.NotificationId)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification.GetNotificationResponse{
		Notification: notif,
	}, nil
}

// ListNotifications retrieves notifications with filtering and pagination
func (s *NotificationService) ListNotifications(ctx context.Context, req *notification.ListNotificationsRequest) (*notification.ListNotificationsResponse, error) {
	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Max page size limit
	}

	notifications, totalCount, err := s.repo.ListNotifications(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	// Get unread count for the recipient
	unreadCount := int32(0)
	if req.RecipientId != "" {
		count, err := s.repo.CountUnreadNotifications(ctx, req.RecipientId)
		if err == nil {
			unreadCount = int32(count)
		}
	}

	return &notification.ListNotificationsResponse{
		Notifications: notifications,
		TotalCount:    totalCount,
		Page:          req.Page,
		PageSize:      req.PageSize,
		UnreadCount:   unreadCount,
	}, nil
}

// MarkNotificationRead marks a notification as read
func (s *NotificationService) MarkNotificationRead(ctx context.Context, req *notification.MarkNotificationReadRequest) (*notification.MarkNotificationReadResponse, error) {
	if req.NotificationId == "" {
		return &notification.MarkNotificationReadResponse{
			Success: false,
			Message: "notification ID is required",
		}, nil
	}

	err := s.repo.MarkNotificationRead(ctx, req.NotificationId, req.UserId)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to mark notification as read: %v", err)
		return &notification.MarkNotificationReadResponse{
			Success: false,
			Message: fmt.Sprintf("failed to mark notification as read: %v", err),
		}, nil
	}

	// Publish notification read event
	if s.publisher != nil {
		notificationEvent := &domain.NotificationEvent{
			NotificationID: req.NotificationId,
			RecipientID:    req.UserId,
			Status:         "read",
			ReadAt:         time.Now(),
		}

		event := domain.NewEventBuilder(domain.EventTypeNotificationRead, req.NotificationId, "notification").
			WithNotification(notificationEvent).
			WithSource("notification-service").
			Build()

		if err := s.publisher.PublishEvent(ctx, event); err != nil {
			s.logger.Log(p9log.LevelWarn, "Failed to publish notification read event: %v", err)
		}
	}

	return &notification.MarkNotificationReadResponse{
		Success: true,
		Message: "notification marked as read",
	}, nil
}

// MarkAllNotificationsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllNotificationsRead(ctx context.Context, req *notification.MarkAllNotificationsReadRequest) (*notification.MarkAllNotificationsReadResponse, error) {
	if req.UserId == "" {
		return &notification.MarkAllNotificationsReadResponse{
			Success: false,
			Message: "user ID is required",
		}, nil
	}

	markedCount, err := s.repo.MarkAllNotificationsRead(ctx, req.UserId, req.TypeFilter)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to mark all notifications as read: %v", err)
		return &notification.MarkAllNotificationsReadResponse{
			Success: false,
			Message: fmt.Sprintf("failed to mark all notifications as read: %v", err),
		}, nil
	}

	return &notification.MarkAllNotificationsReadResponse{
		MarkedCount: markedCount,
		Success:     true,
		Message:     fmt.Sprintf("marked %d notifications as read", markedCount),
	}, nil
}

// CreateTemplate creates a new notification template
func (s *NotificationService) CreateTemplate(ctx context.Context, req *notification.CreateTemplateRequest) (*notification.CreateTemplateResponse, error) {
	if req.Template == nil {
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: "template is required",
		}, nil
	}

	// Validate template
	if err := s.validateTemplate(req.Template); err != nil {
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: fmt.Sprintf("validation failed: %v", err),
		}, nil
	}

	// Generate ID if not provided
	if req.Template.Id == "" {
		req.Template.Id = uuid.New().String()
	}

	// Set defaults
	if req.Template.Language == "" {
		req.Template.Language = "en"
	}
	req.Template.Active = true

	created, err := s.repo.CreateTemplate(ctx, req.Template)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to create template: %v", err)
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create template: %v", err),
		}, nil
	}

	s.logger.Log(p9log.LevelInfo, "Template %s created successfully", created.Id)

	return &notification.CreateTemplateResponse{
		TemplateId: created.Id,
		Success:    true,
		Message:    "template created successfully",
	}, nil
}

// GetTemplate retrieves a template by ID
func (s *NotificationService) GetTemplate(ctx context.Context, req *notification.GetTemplateRequest) (*notification.GetTemplateResponse, error) {
	if req.TemplateId == "" {
		return nil, fmt.Errorf("template ID is required")
	}

	template, err := s.repo.GetTemplate(ctx, req.TemplateId)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &notification.GetTemplateResponse{
		Template: template,
	}, nil
}

// ListTemplates retrieves templates with filtering and pagination
func (s *NotificationService) ListTemplates(ctx context.Context, req *notification.ListTemplatesRequest) (*notification.ListTemplatesResponse, error) {
	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	templates, totalCount, err := s.repo.ListTemplates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return &notification.ListTemplatesResponse{
		Templates:  templates,
		TotalCount: totalCount,
	}, nil
}

// UpdateTemplate updates an existing template
func (s *NotificationService) UpdateTemplate(ctx context.Context, req *notification.UpdateTemplateRequest) (*notification.CreateTemplateResponse, error) {
	if req.TemplateId == "" {
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: "template ID is required",
		}, nil
	}

	if req.Template == nil {
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: "template is required",
		}, nil
	}

	// Validate template
	if err := s.validateTemplate(req.Template); err != nil {
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: fmt.Sprintf("validation failed: %v", err),
		}, nil
	}

	err := s.repo.UpdateTemplate(ctx, req.TemplateId, req.Template)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to update template: %v", err)
		return &notification.CreateTemplateResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update template: %v", err),
		}, nil
	}

	s.logger.Log(p9log.LevelInfo, "Template %s updated successfully", req.TemplateId)

	return &notification.CreateTemplateResponse{
		TemplateId: req.TemplateId,
		Success:    true,
		Message:    "template updated successfully",
	}, nil
}

// DeleteTemplate deletes a template
func (s *NotificationService) DeleteTemplate(ctx context.Context, req *notification.DeleteTemplateRequest) error {
	if req.TemplateId == "" {
		return fmt.Errorf("template ID is required")
	}

	err := s.repo.DeleteTemplate(ctx, req.TemplateId)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	s.logger.Log(p9log.LevelInfo, "Template %s deleted successfully", req.TemplateId)
	return nil
}

// GetUserPreferences retrieves user notification preferences
func (s *NotificationService) GetUserPreferences(ctx context.Context, req *notification.GetUserPreferencesRequest) (*notification.GetUserPreferencesResponse, error) {
	if req.UserId == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	preferences, err := s.repo.GetUserPreferences(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return &notification.GetUserPreferencesResponse{
		Preferences: preferences,
	}, nil
}

// UpdateUserPreferences updates user notification preferences
func (s *NotificationService) UpdateUserPreferences(ctx context.Context, req *notification.UpdateUserPreferencesRequest) (*notification.UpdateUserPreferencesResponse, error) {
	if req.UserId == "" {
		return &notification.UpdateUserPreferencesResponse{
			Success: false,
			Message: "user ID is required",
		}, nil
	}

	err := s.repo.UpdateUserPreferences(ctx, req.UserId, req.Preferences)
	if err != nil {
		s.logger.Log(p9log.LevelError, "Failed to update user preferences: %v", err)
		return &notification.UpdateUserPreferencesResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update preferences: %v", err),
		}, nil
	}

	s.logger.Log(p9log.LevelInfo, "User preferences updated for user %s", req.UserId)

	return &notification.UpdateUserPreferencesResponse{
		Success: true,
		Message: "preferences updated successfully",
	}, nil
}

// SendWorkflowNotification sends workflow-specific notifications
func (s *NotificationService) SendWorkflowNotification(ctx context.Context, req *notification.SendWorkflowNotificationRequest) (*notification.SendNotificationResponse, error) {
	if req.WorkflowNotification == nil {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: "workflow notification is required",
		}, nil
	}

	// Create notifications for each recipient
	var lastNotificationID string
	successCount := 0

	for _, recipientID := range req.RecipientIds {
		notif := &notification.Notification{
			Id:            uuid.New().String(),
			Type:          notification.NotificationType_WORKFLOW_STATE_CHANGE,
			Channel:       req.Channel,
			Priority:      req.Priority,
			RecipientId:   recipientID,
			RecipientType: req.RecipientType,
			Subject:       s.generateWorkflowSubject(req.WorkflowNotification),
			Message:       s.generateWorkflowMessage(req.WorkflowNotification),
			TemplateId:    req.TemplateId,
			SourceId:      req.WorkflowNotification.InstanceId,
			SourceType:    "workflow",
		}

		s.setNotificationDefaults(notif)

		// Process template if provided
		if notif.TemplateId != "" {
			if err := s.processTemplate(ctx, notif); err != nil {
				s.logger.Log(p9log.LevelWarn, "Failed to process template for workflow notification: %v", err)
			}
		}

		created, err := s.repo.CreateNotification(ctx, notif)
		if err != nil {
			s.logger.Log(p9log.LevelError, "Failed to create workflow notification for %s: %v", recipientID, err)
			continue
		}

		lastNotificationID = created.Id
		successCount++
	}

	if successCount == 0 {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: "failed to send workflow notifications to any recipient",
		}, nil
	}

	return &notification.SendNotificationResponse{
		NotificationId: lastNotificationID,
		Success:        true,
		Message:        fmt.Sprintf("workflow notification sent to %d recipients", successCount),
	}, nil
}

// SendSLANotification sends SLA-specific notifications
func (s *NotificationService) SendSLANotification(ctx context.Context, req *notification.SendSLANotificationRequest) (*notification.SendNotificationResponse, error) {
	if req.SlaNotification == nil {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: "SLA notification is required",
		}, nil
	}

	// Determine notification type based on SLA event type
	notifType := s.mapSLAEventTypeToNotificationType(req.SlaNotification.EventType)

	// Create notifications for each recipient
	var lastNotificationID string
	successCount := 0

	for _, recipientID := range req.RecipientIds {
		notif := &notification.Notification{
			Id:            uuid.New().String(),
			Type:          notifType,
			Channel:       req.Channel,
			Priority:      req.Priority,
			RecipientId:   recipientID,
			RecipientType: req.RecipientType,
			Subject:       s.generateSLASubject(req.SlaNotification),
			Message:       s.generateSLAMessage(req.SlaNotification),
			TemplateId:    req.TemplateId,
			SourceId:      req.SlaNotification.SlaInstanceId,
			SourceType:    "sla",
		}

		s.setNotificationDefaults(notif)

		// Process template if provided
		if notif.TemplateId != "" {
			if err := s.processTemplate(ctx, notif); err != nil {
				s.logger.Log(p9log.LevelWarn, "Failed to process template for SLA notification: %v", err)
			}
		}

		created, err := s.repo.CreateNotification(ctx, notif)
		if err != nil {
			s.logger.Log(p9log.LevelError, "Failed to create SLA notification for %s: %v", recipientID, err)
			continue
		}

		lastNotificationID = created.Id
		successCount++
	}

	if successCount == 0 {
		return &notification.SendNotificationResponse{
			Success: false,
			Message: "failed to send SLA notifications to any recipient",
		}, nil
	}

	return &notification.SendNotificationResponse{
		NotificationId: lastNotificationID,
		Success:        true,
		Message:        fmt.Sprintf("SLA notification sent to %d recipients", successCount),
	}, nil
}

// GetNotificationStats retrieves notification statistics
func (s *NotificationService) GetNotificationStats(ctx context.Context, req *notification.GetNotificationStatsRequest) (*notification.GetNotificationStatsResponse, error) {
	if req.UserId == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	stats, err := s.repo.GetNotificationStats(ctx, req.UserId, req.FromDate, req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	return stats, nil
}

// Helper methods

// validateNotification validates notification data
func (s *NotificationService) validateNotification(notif *notification.Notification) error {
	if notif.RecipientId == "" {
		return fmt.Errorf("recipient ID is required")
	}
	if notif.Subject == "" && notif.Message == "" && notif.TemplateId == "" {
		return fmt.Errorf("subject, message, or template ID is required")
	}
	return nil
}

// validateTemplate validates template data
func (s *NotificationService) validateTemplate(template *notification.NotificationTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.SubjectTemplate == "" && template.BodyTemplate == "" {
		return fmt.Errorf("subject template or body template is required")
	}
	return nil
}

// setNotificationDefaults sets default values for notification
func (s *NotificationService) setNotificationDefaults(notif *notification.Notification) {
	if notif.Status == notification.NotificationStatus_PENDING {
		notif.Status = notification.NotificationStatus_PENDING
	}
	if notif.Priority == notification.NotificationPriority_LOW {
		notif.Priority = notification.NotificationPriority_NORMAL
	}
	if notif.Channel == notification.NotificationChannel_EMAIL {
		notif.Channel = notification.NotificationChannel_EMAIL
	}
	if notif.ScheduledAt == nil {
		notif.ScheduledAt = timestamppb.Now()
	}
	if notif.CreatedAt == nil {
		notif.CreatedAt = timestamppb.Now()
	}
	if notif.UpdatedAt == nil {
		notif.UpdatedAt = timestamppb.Now()
	}
}

// processTemplate processes notification template
func (s *NotificationService) processTemplate(ctx context.Context, notif *notification.Notification) error {
	template, err := s.repo.GetTemplate(ctx, notif.TemplateId)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Simple template processing - replace variables
	if template.SubjectTemplate != "" {
		notif.Subject = s.processTemplateString(template.SubjectTemplate, notif.TemplateData)
	}
	if template.BodyTemplate != "" {
		notif.Message = s.processTemplateString(template.BodyTemplate, notif.TemplateData)
	}

	return nil
}

// processTemplateString processes template string with variables
func (s *NotificationService) processTemplateString(template string, data map[string]*anypb.Any) string {
	// Simple variable replacement - in production, use a proper template engine
	result := template
	for key, value := range data {
		// Convert Any to string for template replacement
		var strValue string
		if value != nil {
			// Try to unmarshal based on the type URL
			switch value.GetTypeUrl() {
			case "type.googleapis.com/google.protobuf.StringValue":
				var stringWrapper wrapperspb.StringValue
				if err := value.UnmarshalTo(&stringWrapper); err == nil {
					strValue = stringWrapper.GetValue()
				} else {
					strValue = value.String()
				}
			default:
				// For other types, try to extract as string or use string representation
				strValue = value.String()
			}
		}

		// Replace template variables in format {{key}} or ${key}
		result = strings.ReplaceAll(result, "{{"+key+"}}", strValue)
		result = strings.ReplaceAll(result, "${"+key+"}", strValue)
	}
	return result
}

// generateWorkflowSubject generates subject for workflow notifications
func (s *NotificationService) generateWorkflowSubject(workflow *notification.WorkflowStateChangeNotification) string {
	return fmt.Sprintf("Workflow State Changed: %s → %s", workflow.PreviousState, workflow.CurrentState)
}

// generateWorkflowMessage generates message for workflow notifications
func (s *NotificationService) generateWorkflowMessage(workflow *notification.WorkflowStateChangeNotification) string {
	return fmt.Sprintf("Form %s has transitioned from %s to %s by %s",
		workflow.FormId, workflow.PreviousState, workflow.CurrentState, workflow.ChangedBy)
}

// generateSLASubject generates subject for SLA notifications
func (s *NotificationService) generateSLASubject(sla *notification.SLAEventNotification) string {
	return fmt.Sprintf("SLA %s: %s", sla.EventType.String(), sla.SlaRuleName)
}

// generateSLAMessage generates message for SLA notifications
func (s *NotificationService) generateSLAMessage(sla *notification.SLAEventNotification) string {
	return fmt.Sprintf("SLA event %s occurred for rule %s in state %s",
		sla.EventType.String(), sla.SlaRuleName, sla.State)
}

// mapSLAEventTypeToNotificationType maps SLA event type to notification type
func (s *NotificationService) mapSLAEventTypeToNotificationType(eventType notification.SLAEventType) notification.NotificationType {
	switch eventType {
	case notification.SLAEventType_SLA_WARNING_THRESHOLD:
		return notification.NotificationType_SLA_WARNING
	case notification.SLAEventType_SLA_BREACHED:
		return notification.NotificationType_SLA_BREACH
	case notification.SLAEventType_SLA_ESCALATED:
		return notification.NotificationType_SLA_ESCALATION
	case notification.SLAEventType_SLA_RESOLVED:
		return notification.NotificationType_SLA_COMPLETION
	default:
		return notification.NotificationType_SLA_WARNING
	}
}

// mapNotificationPriorityToDomainPriority maps notification priority to domain priority
func (s *NotificationService) mapNotificationPriorityToDomainPriority(priority notification.NotificationPriority) domain.Priority {
	switch priority {
	case notification.NotificationPriority_LOW:
		return domain.PriorityLow
	case notification.NotificationPriority_NORMAL:
		return domain.PriorityMedium
	case notification.NotificationPriority_HIGH:
		return domain.PriorityHigh
	case notification.NotificationPriority_URGENT, notification.NotificationPriority_CRITICAL:
		return domain.PriorityCritical
	default:
		return domain.PriorityMedium
	}
}
