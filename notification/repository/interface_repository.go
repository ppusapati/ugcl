package repository

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
	"p9e.in/ugcl/notification/api/v2/notification"
)

// NotificationRepository defines the interface for notification data operations
type INotificationRepository interface {
	// Core notification operations
	CreateNotification(ctx context.Context, notif *notification.Notification) (*notification.Notification, error)
	GetNotification(ctx context.Context, notificationID string) (*notification.Notification, error)
	ListNotifications(ctx context.Context, req *notification.ListNotificationsRequest) ([]*notification.Notification, int64, error)
	UpdateNotificationStatus(ctx context.Context, notificationID string, status notification.NotificationStatus) error
	MarkNotificationRead(ctx context.Context, notificationID, userID string) error
	MarkAllNotificationsRead(ctx context.Context, userID string, typeFilter []notification.NotificationType) (int32, error)

	// Template operations
	CreateTemplate(ctx context.Context, template *notification.NotificationTemplate) (*notification.NotificationTemplate, error)
	GetTemplate(ctx context.Context, templateID string) (*notification.NotificationTemplate, error)
	ListTemplates(ctx context.Context, req *notification.ListTemplatesRequest) ([]*notification.NotificationTemplate, int64, error)
	UpdateTemplate(ctx context.Context, templateID string, template *notification.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, templateID string) error
	GetTemplateByTypeChannel(ctx context.Context, notifType notification.NotificationType, channel notification.NotificationChannel, language string) (*notification.NotificationTemplate, error)

	// User preferences
	GetUserPreferences(ctx context.Context, userID string) ([]*notification.NotificationPreference, error)
	UpdateUserPreferences(ctx context.Context, userID string, preferences []*notification.NotificationPreference) error

	// Statistics
	GetNotificationStats(ctx context.Context, userID string, fromDate, toDate *timestamppb.Timestamp) (*notification.GetNotificationStatsResponse, error)
	CountUnreadNotifications(ctx context.Context, userID string) (int64, error)

	// Bulk operations
	CreateBulkNotifications(ctx context.Context, notifications []*notification.Notification) ([]string, error)
	GetFailedNotifications(ctx context.Context, maxRetryCount int32, limit int32) ([]*notification.Notification, error)
}
