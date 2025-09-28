package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportSubscription represents a user's subscription to a report
type ReportSubscription struct {
	ID               uuid.UUID
	ReportID         uuid.UUID
	UserID           string
	SubscriptionType string   // email_digest, real_time, scheduled
	Frequency        string   // daily, weekly, monthly, on_change
	DeliveryChannels []string // email, slack, webhook
	IsActive         bool
	LastDeliveryAt   *time.Time
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedBy        string
	UpdatedAt        time.Time
	Settings         map[string]interface{} // User-specific delivery preferences
}

// SubscriptionDelivery represents a delivery of a subscription
type SubscriptionDelivery struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	RunID          *uuid.UUID // The report run that triggered this delivery
	Channel        string     // email, slack, webhook
	Status         string     // pending, sent, failed, bounced
	Recipients     []string
	Subject        string
	Content        string
	SentAt         *time.Time
	ErrorMsg       *string
	CreatedAt      time.Time
}