package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WorkflowHistory tracks state transitions
type WorkflowHistory struct {
	BaseModel
	InstanceID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	FormID      string         `gorm:"type:varchar(255);not null"`
	FromState   string         `gorm:"type:varchar(100)"`
	ToState     string         `gorm:"type:varchar(100);not null"`
	Event       string         `gorm:"type:varchar(100);not null"`
	PerformedBy string         `gorm:"type:varchar(255);not null"`
	PerformedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Comments    string         `gorm:"type:text"`
	Metadata    datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	Instance *FormInstance `gorm:"foreignKey:InstanceID"`
}
