package services

import (
	"context"

	"p9e.in/ugcl/notification/api/v2/notification"
)

// NotificationService defines the interface for notification business logic
type INotificationService interface {
	// Core notification operations
	SendNotification(ctx context.Context, req *notification.SendNotificationRequest) (*notification.SendNotificationResponse, error)
	SendBulkNotification(ctx context.Context, req *notification.SendBulkNotificationRequest) (*notification.SendBulkNotificationResponse, error)
	GetNotification(ctx context.Context, req *notification.GetNotificationRequest) (*notification.GetNotificationResponse, error)
	ListNotifications(ctx context.Context, req *notification.ListNotificationsRequest) (*notification.ListNotificationsResponse, error)

	// Notification management
	MarkNotificationRead(ctx context.Context, req *notification.MarkNotificationReadRequest) (*notification.MarkNotificationReadResponse, error)
	MarkAllNotificationsRead(ctx context.Context, req *notification.MarkAllNotificationsReadRequest) (*notification.MarkAllNotificationsReadResponse, error)

	// Template management
	CreateTemplate(ctx context.Context, req *notification.CreateTemplateRequest) (*notification.CreateTemplateResponse, error)
	GetTemplate(ctx context.Context, req *notification.GetTemplateRequest) (*notification.GetTemplateResponse, error)
	ListTemplates(ctx context.Context, req *notification.ListTemplatesRequest) (*notification.ListTemplatesResponse, error)
	UpdateTemplate(ctx context.Context, req *notification.UpdateTemplateRequest) (*notification.CreateTemplateResponse, error)
	DeleteTemplate(ctx context.Context, req *notification.DeleteTemplateRequest) error

	// User preferences
	GetUserPreferences(ctx context.Context, req *notification.GetUserPreferencesRequest) (*notification.GetUserPreferencesResponse, error)
	UpdateUserPreferences(ctx context.Context, req *notification.UpdateUserPreferencesRequest) (*notification.UpdateUserPreferencesResponse, error)

	// Specialized notifications
	SendWorkflowNotification(ctx context.Context, req *notification.SendWorkflowNotificationRequest) (*notification.SendNotificationResponse, error)
	SendSLANotification(ctx context.Context, req *notification.SendSLANotificationRequest) (*notification.SendNotificationResponse, error)

	// Statistics and reporting
	GetNotificationStats(ctx context.Context, req *notification.GetNotificationStatsRequest) (*notification.GetNotificationStatsResponse, error)
}
