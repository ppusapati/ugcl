package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FormInstance tracks all form submissions
type FormInstance struct {
	BaseModel
	FormID        string    `gorm:"type:varchar(255);not null;index"`
	FormVersion   string    `gorm:"type:varchar(50);not null"`
	InstanceTable string    `gorm:"type:varchar(255);not null"`
	InstanceID    uuid.UUID `gorm:"type:uuid;not null"`
	CurrentState  string    `gorm:"type:varchar(100);not null;index"`
	CreatedBy     string    `gorm:"type:varchar(255);not null"`
	AssignedTo    string    `gorm:"type:varchar(255);index"`
	CompletedAt   *time.Time
	Metadata      datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	FormDefinition  *FormDefinition   `gorm:"foreignKey:FormID,FormVersion;references:FormID,Version"`
	WorkflowHistory []WorkflowHistory `gorm:"foreignKey:InstanceID"`
	Attachments     []FormAttachment  `gorm:"foreignKey:InstanceID"`
	Comments        []FormComment     `gorm:"foreignKey:InstanceID"`
}
