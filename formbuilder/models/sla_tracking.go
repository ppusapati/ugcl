package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SLATracking monitors SLA compliance
type SLATracking struct {
	BaseModel
	InstanceID       uuid.UUID `gorm:"type:uuid;not null;index"`
	SLARuleID        string    `gorm:"type:varchar(255);not null"`
	State            string    `gorm:"type:varchar(100);not null"`
	StartTime        time.Time `gorm:"not null"`
	DueTime          time.Time `gorm:"not null;index"`
	CompletedTime    *time.Time
	IsBreached       bool `gorm:"default:false;index"`
	BreachTime       *time.Time
	EscalationLevel  int `gorm:"default:0"`
	LastEscalationAt *time.Time
	Metadata         datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	Instance *FormInstance `gorm:"foreignKey:InstanceID"`
}
