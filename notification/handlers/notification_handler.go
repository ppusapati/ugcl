package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"p9e.in/ugcl/notification/api/v2/notification"
	"p9e.in/ugcl/notification/api/v2/notification/notificationconnect"
	"p9e.in/ugcl/notification/services"
	"p9e.in/ugcl/packages/p9log"
)

// NotificationHandler implements the Connect RPC handlers for the notification service
type NotificationHandler struct {
	notificationService services.NotificationService
	logger              p9log.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(
	notificationService services.NotificationService,
	logger p9log.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// Ensure NotificationHandler implements the NotificationServiceHandler interface
var _ notificationconnect.NotificationServiceHandler = (*NotificationHandler)(nil)

// SendNotification handles sending a single notification
func (h *NotificationHandler) SendNotification(
	ctx context.Context,
	req *connect.Request[notification.SendNotificationRequest],
) (*connect.Response[notification.SendNotificationResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received SendNotification request", map[string]interface{}{
		"recipient_id": req.Msg.Notification.RecipientId,
		"type":         req.Msg.Notification.Type.String(),
		"channel":      req.Msg.Notification.Channel.String(),
	})

	response, err := h.notificationService.SendNotification(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to send notification", err, map[string]interface{}{
			"recipient_id": req.Msg.Notification.RecipientId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to send notification: %w", err))
	}

	return connect.NewResponse(response), nil
}

// SendBulkNotification handles sending multiple notifications
func (h *NotificationHandler) SendBulkNotification(
	ctx context.Context,
	req *connect.Request[notification.SendBulkNotificationRequest],
) (*connect.Response[notification.SendBulkNotificationResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received SendBulkNotification request", map[string]interface{}{
		"recipient_count": len(req.Msg.BulkRequest.RecipientIds),
	})

	response, err := h.notificationService.SendBulkNotification(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to send bulk notifications", err, map[string]interface{}{
			"recipient_count": len(req.Msg.BulkRequest.RecipientIds),
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to send bulk notifications: %w", err))
	}

	return connect.NewResponse(response), nil
}

// GetNotification handles retrieving a specific notification
func (h *NotificationHandler) GetNotification(
	ctx context.Context,
	req *connect.Request[notification.GetNotificationRequest],
) (*connect.Response[notification.GetNotificationResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received GetNotification request", map[string]interface{}{
		"notification_id": req.Msg.NotificationId,
	})

	response, err := h.notificationService.GetNotification(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to get notification", err, map[string]interface{}{
			"notification_id": req.Msg.NotificationId,
		})
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("notification not found: %w", err))
	}

	return connect.NewResponse(response), nil
}

// ListNotifications handles listing notifications with filtering and pagination
func (h *NotificationHandler) ListNotifications(
	ctx context.Context,
	req *connect.Request[notification.ListNotificationsRequest],
) (*connect.Response[notification.ListNotificationsResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received ListNotifications request", map[string]interface{}{
		"recipient_id": req.Msg.RecipientId,
		"page":         req.Msg.Page,
		"page_size":    req.Msg.PageSize,
	})

	response, err := h.notificationService.ListNotifications(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to list notifications", err, map[string]interface{}{
			"recipient_id": req.Msg.RecipientId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list notifications: %w", err))
	}

	return connect.NewResponse(response), nil
}

// MarkNotificationRead handles marking a notification as read
func (h *NotificationHandler) MarkNotificationRead(
	ctx context.Context,
	req *connect.Request[notification.MarkNotificationReadRequest],
) (*connect.Response[notification.MarkNotificationReadResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received MarkNotificationRead request", map[string]interface{}{
		"notification_id": req.Msg.NotificationId,
		"user_id":         req.Msg.UserId,
	})

	response, err := h.notificationService.MarkNotificationRead(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to mark notification as read", err, map[string]interface{}{
			"notification_id": req.Msg.NotificationId,
			"user_id":         req.Msg.UserId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to mark notification as read: %w", err))
	}

	return connect.NewResponse(response), nil
}

// MarkAllNotificationsRead handles marking all notifications as read for a user
func (h *NotificationHandler) MarkAllNotificationsRead(
	ctx context.Context,
	req *connect.Request[notification.MarkAllNotificationsReadRequest],
) (*connect.Response[notification.MarkAllNotificationsReadResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received MarkAllNotificationsRead request", map[string]interface{}{
		"user_id": req.Msg.UserId,
	})

	response, err := h.notificationService.MarkAllNotificationsRead(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to mark all notifications as read", err, map[string]interface{}{
			"user_id": req.Msg.UserId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to mark all notifications as read: %w", err))
	}

	return connect.NewResponse(response), nil
}

// CreateTemplate handles creating a new notification template
func (h *NotificationHandler) CreateTemplate(
	ctx context.Context,
	req *connect.Request[notification.CreateTemplateRequest],
) (*connect.Response[notification.CreateTemplateResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received CreateTemplate request", map[string]interface{}{
		"template_name": req.Msg.Template.Name,
		"template_type": req.Msg.Template.Type.String(),
	})

	response, err := h.notificationService.CreateTemplate(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to create template", err, map[string]interface{}{
			"template_name": req.Msg.Template.Name,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create template: %w", err))
	}

	return connect.NewResponse(response), nil
}

// GetTemplate handles retrieving a specific template
func (h *NotificationHandler) GetTemplate(
	ctx context.Context,
	req *connect.Request[notification.GetTemplateRequest],
) (*connect.Response[notification.GetTemplateResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received GetTemplate request", map[string]interface{}{
		"template_id": req.Msg.TemplateId,
	})

	response, err := h.notificationService.GetTemplate(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to get template", err, map[string]interface{}{
			"template_id": req.Msg.TemplateId,
		})
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("template not found: %w", err))
	}

	return connect.NewResponse(response), nil
}

// ListTemplates handles listing notification templates
func (h *NotificationHandler) ListTemplates(
	ctx context.Context,
	req *connect.Request[notification.ListTemplatesRequest],
) (*connect.Response[notification.ListTemplatesResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received ListTemplates request", map[string]interface{}{
		"page":      req.Msg.Page,
		"page_size": req.Msg.PageSize,
	})

	response, err := h.notificationService.ListTemplates(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to list templates", err, map[string]interface{}{
			"page":      req.Msg.Page,
			"page_size": req.Msg.PageSize,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list templates: %w", err))
	}

	return connect.NewResponse(response), nil
}

// UpdateTemplate handles updating an existing template
func (h *NotificationHandler) UpdateTemplate(
	ctx context.Context,
	req *connect.Request[notification.UpdateTemplateRequest],
) (*connect.Response[notification.CreateTemplateResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received UpdateTemplate request", map[string]interface{}{
		"template_id":   req.Msg.TemplateId,
		"template_name": req.Msg.Template.Name,
	})

	response, err := h.notificationService.UpdateTemplate(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to update template", err, map[string]interface{}{
			"template_id": req.Msg.TemplateId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update template: %w", err))
	}

	return connect.NewResponse(response), nil
}

// DeleteTemplate handles deleting a template
func (h *NotificationHandler) DeleteTemplate(
	ctx context.Context,
	req *connect.Request[notification.DeleteTemplateRequest],
) (*connect.Response[emptypb.Empty], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received DeleteTemplate request", map[string]interface{}{
		"template_id": req.Msg.TemplateId,
	})

	err := h.notificationService.DeleteTemplate(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to delete template", err, map[string]interface{}{
			"template_id": req.Msg.TemplateId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete template: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetUserPreferences handles retrieving user notification preferences
func (h *NotificationHandler) GetUserPreferences(
	ctx context.Context,
	req *connect.Request[notification.GetUserPreferencesRequest],
) (*connect.Response[notification.GetUserPreferencesResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received GetUserPreferences request", map[string]interface{}{
		"user_id": req.Msg.UserId,
	})

	response, err := h.notificationService.GetUserPreferences(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to get user preferences", err, map[string]interface{}{
			"user_id": req.Msg.UserId,
		})
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user preferences not found: %w", err))
	}

	return connect.NewResponse(response), nil
}

// UpdateUserPreferences handles updating user notification preferences
func (h *NotificationHandler) UpdateUserPreferences(
	ctx context.Context,
	req *connect.Request[notification.UpdateUserPreferencesRequest],
) (*connect.Response[notification.UpdateUserPreferencesResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received UpdateUserPreferences request", map[string]interface{}{
		"user_id": req.Msg.UserId,
	})

	response, err := h.notificationService.UpdateUserPreferences(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to update user preferences", err, map[string]interface{}{
			"user_id": req.Msg.UserId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user preferences: %w", err))
	}

	return connect.NewResponse(response), nil
}

// SendWorkflowNotification handles sending workflow-specific notifications
func (h *NotificationHandler) SendWorkflowNotification(
	ctx context.Context,
	req *connect.Request[notification.SendWorkflowNotificationRequest],
) (*connect.Response[notification.SendNotificationResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received SendWorkflowNotification request", map[string]interface{}{
		"instance_id":      req.Msg.WorkflowNotification.InstanceId,
		"transition_event": req.Msg.WorkflowNotification.TransitionEvent,
		"recipient_count":  len(req.Msg.RecipientIds),
	})

	response, err := h.notificationService.SendWorkflowNotification(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to send workflow notification", err, map[string]interface{}{
			"instance_id":      req.Msg.WorkflowNotification.InstanceId,
			"transition_event": req.Msg.WorkflowNotification.TransitionEvent,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to send workflow notification: %w", err))
	}

	return connect.NewResponse(response), nil
}

// SendSLANotification handles sending SLA-specific notifications
func (h *NotificationHandler) SendSLANotification(
	ctx context.Context,
	req *connect.Request[notification.SendSLANotificationRequest],
) (*connect.Response[notification.SendNotificationResponse], error) {
	h.logger.Log(p9log.LevelInfo, ctx, "Received SendSLANotification request", map[string]interface{}{
		"sla_rule_id":     req.Msg.SlaNotification.SlaRuleId,
		"instance_id":     req.Msg.SlaNotification.InstanceId,
		"breach_type":     req.Msg.SlaNotification.EventType.String(),
		"recipient_count": len(req.Msg.RecipientIds),
	})

	response, err := h.notificationService.SendSLANotification(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to send SLA notification", err, map[string]interface{}{
			"sla_rule_id": req.Msg.SlaNotification.SlaRuleId,
			"instance_id": req.Msg.SlaNotification.InstanceId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to send SLA notification: %w", err))
	}

	return connect.NewResponse(response), nil
}

// GetNotificationStats handles retrieving notification statistics
func (h *NotificationHandler) GetNotificationStats(
	ctx context.Context,
	req *connect.Request[notification.GetNotificationStatsRequest],
) (*connect.Response[notification.GetNotificationStatsResponse], error) {
	h.logger.Log(p9log.LevelDebug, ctx, "Received GetNotificationStats request", map[string]interface{}{
		"user_id":   req.Msg.UserId,
		"from_date": req.Msg.FromDate,
		"to_date":   req.Msg.ToDate,
	})

	response, err := h.notificationService.GetNotificationStats(ctx, req.Msg)
	if err != nil {
		h.logger.Log(p9log.LevelError, ctx, "Failed to get notification stats", err, map[string]interface{}{
			"user_id": req.Msg.UserId,
		})
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get notification stats: %w", err))
	}

	return connect.NewResponse(response), nil
}
