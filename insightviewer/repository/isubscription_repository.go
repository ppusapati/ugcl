package repository

import (
	"context"

	"github.com/google/uuid"

	"p9e.in/ugcl/insightviewer/models"
)

// ISubscriptionRepository defines the interface for report subscription data access operations
type ISubscriptionRepository interface {
	// Report Subscription operations
	CreateReportSubscription(ctx context.Context, subscription *models.ReportSubscription) (*models.ReportSubscription, error)
	GetReportSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.ReportSubscription, error)
	UpdateReportSubscription(ctx context.Context, subscription *models.ReportSubscription) (*models.ReportSubscription, error)
	DeleteReportSubscription(ctx context.Context, id uuid.UUID) error
	GetReportSubscriptions(ctx context.Context, reportID uuid.UUID) ([]*models.ReportSubscription, error)
	GetUserSubscriptions(ctx context.Context, userID string) ([]*models.ReportSubscription, error)
	GetActiveSubscriptions(ctx context.Context) ([]*models.ReportSubscription, error)

	// Subscription Delivery operations
	CreateSubscriptionDelivery(ctx context.Context, delivery *models.SubscriptionDelivery) (*models.SubscriptionDelivery, error)
	GetSubscriptionDeliveryByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionDelivery, error)
	UpdateSubscriptionDelivery(ctx context.Context, delivery *models.SubscriptionDelivery) (*models.SubscriptionDelivery, error)
	GetSubscriptionDeliveries(ctx context.Context, subscriptionID uuid.UUID, limit, offset int32) ([]*models.SubscriptionDelivery, int32, error)
	MarkDeliveryAsSent(ctx context.Context, deliveryID uuid.UUID) error
	MarkDeliveryAsFailed(ctx context.Context, deliveryID uuid.UUID, errorMsg string) error
}