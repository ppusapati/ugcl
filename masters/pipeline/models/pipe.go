package models

import (
	"time"

	"gorm.io/gorm"
)

type PipeType string

const (
	PipeTypeHDPE PipeType = "HDPE"
	PipeTypeDI   PipeType = "DI"
	PipeTypeMS   PipeType = "MS"
)

type Pipe struct {
	ID             string   `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	PipeName       string   `gorm:"type:varchar(50);not null;unique" json:"pipe_name"`
	PipeType       PipeType `gorm:"type:pipe_type_enum;not null" json:"pipe_type"`
	PressureRating string   `gorm:"type:varchar(20)" json:"pressure_rating"`
	Description    string   `gorm:"type:text" json:"description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	PipelineSegments []PipelineSegment `gorm:"foreignKey:PipeID;constraint:OnDelete:RESTRICT" json:"pipeline_segments,omitempty"`
}

func (Pipe) TableName() string {
	return "pipes"
}
