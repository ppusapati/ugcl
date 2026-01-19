package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FormAttachment stores file metadata
type FormAttachment struct {
	BaseModel
	FormID      string         `gorm:"type:varchar(255);not null;index:idx_form_field"`
	InstanceID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	FieldID     string         `gorm:"type:varchar(255);not null;index:idx_form_field"`
	Filename    string         `gorm:"type:varchar(500);not null"`
	MimeType    string         `gorm:"type:varchar(255);not null"`
	SizeBytes   int64          `gorm:"not null"`
	StoragePath string         `gorm:"type:text;not null"`
	Checksum    string         `gorm:"type:varchar(64)"`
	UploadedBy  string         `gorm:"type:varchar(255);not null"`
	Metadata    datatypes.JSON `gorm:"type:jsonb"`
	IsDeleted   bool           `gorm:"default:false"`
	DeletedAt   *time.Time
	DeletedBy   string `gorm:"type:varchar(255)"`

	// Relations
	Instance *FormInstance `gorm:"foreignKey:InstanceID"`
}
