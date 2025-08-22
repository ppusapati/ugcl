package models

import (
	"time"

	"gorm.io/gorm"
)

type Site struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	SiteName    string         `gorm:"type:varchar(100);not null;unique" json:"site_name"`
	SiteCode    string         `gorm:"type:varchar(20)" json:"site_code"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Zones            []Zone            `gorm:"foreignKey:SiteID;constraint:OnDelete:CASCADE" json:"zones,omitempty"`
	Nodes            []Node            `gorm:"foreignKey:SiteID;constraint:OnDelete:CASCADE" json:"nodes,omitempty"`
	PipelineSegments []PipelineSegment `gorm:"foreignKey:SiteID;constraint:OnDelete:CASCADE" json:"pipeline_segments,omitempty"`
}

func (Site) TableName() string {
	return "sites"
}
