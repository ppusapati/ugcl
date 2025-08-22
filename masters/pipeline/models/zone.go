package models

import (
	"time"

	"gorm.io/gorm"
)

type Zone struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	SiteID      string         `gorm:"not null" json:"site_id"`
	ZoneName    string         `gorm:"type:varchar(100);not null" json:"zone_name"`
	ZoneCode    string         `gorm:"type:varchar(20)" json:"zone_code"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Site             Site              `gorm:"foreignKey:SiteID;references:ID;constraint:OnDelete:CASCADE" json:"site,omitempty"`
	Nodes            []Node            `gorm:"foreignKey:ZoneID;constraint:OnDelete:SET NULL" json:"nodes,omitempty"`
	PipelineSegments []PipelineSegment `gorm:"foreignKey:ZoneID;constraint:OnDelete:CASCADE" json:"pipeline_segments,omitempty"`
}

func (Zone) TableName() string {
	return "zones"
}
