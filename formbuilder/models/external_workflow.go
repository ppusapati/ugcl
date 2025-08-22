package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ExternalWorkflowInstance tracks BPMN engine integrations
type ExternalWorkflowInstance struct {
	BaseModel
	FormInstanceID       uuid.UUID `gorm:"type:uuid;not null;index"`
	EngineType           string    `gorm:"type:varchar(50);not null"` // 'camunda', 'flowable', 'zeebe'
	ProcessDefinitionKey string    `gorm:"type:varchar(255);not null"`
	ProcessInstanceID    string    `gorm:"type:varchar(255);not null;index"`
	Status               string    `gorm:"type:varchar(50);default:'active'"`
	StartedAt            time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	CompletedAt          *time.Time
	Variables            datatypes.JSON `gorm:"type:jsonb"`
	Result               datatypes.JSON `gorm:"type:jsonb"`
	ErrorMessage         string         `gorm:"type:text"`
	Metadata             datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	FormInstance *FormInstance `gorm:"foreignKey:FormInstanceID"`
}
