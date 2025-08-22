package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FormNotification stores system notifications
type FormNotification struct {
	BaseModel
	InstanceID       *uuid.UUID `gorm:"type:uuid"`
	Recipient        string     `gorm:"type:varchar(255);not null;index"`
	RecipientType    string     `gorm:"type:varchar(50);not null"` // 'user', 'role', 'email'
	NotificationType string     `gorm:"type:varchar(100);not null"`
	Subject          string     `gorm:"type:varchar(500)"`
	Body             string     `gorm:"type:text;not null"`
	Status           string     `gorm:"type:varchar(50);default:'pending';index"` // 'pending', 'sent', 'failed', 'read'
	SentAt           *time.Time
	ReadAt           *time.Time
	ErrorMessage     string         `gorm:"type:text"`
	RetryCount       int            `gorm:"default:0"`
	Metadata         datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	Instance *FormInstance `gorm:"foreignKey:InstanceID"`
}
