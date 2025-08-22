package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FormComment stores comments on form instances
type FormComment struct {
	BaseModel
	InstanceID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID          string         `gorm:"type:varchar(255);not null"`
	CommentText     string         `gorm:"type:text;not null"`
	IsInternal      bool           `gorm:"default:false"`
	IsDeleted       bool           `gorm:"default:false"`
	ParentCommentID *uuid.UUID     `gorm:"type:uuid"`
	Metadata        datatypes.JSON `gorm:"type:jsonb"`

	// Relations
	Instance      *FormInstance `gorm:"foreignKey:InstanceID"`
	ParentComment *FormComment  `gorm:"foreignKey:ParentCommentID"`
	Replies       []FormComment `gorm:"foreignKey:ParentCommentID"`
}
