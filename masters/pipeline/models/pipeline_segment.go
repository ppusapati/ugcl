package models

import (
	"time"

	"gorm.io/gorm"
)

type SegmentStatus string

const (
	SegmentStatusActive      SegmentStatus = "active"
	SegmentStatusInactive    SegmentStatus = "inactive"
	SegmentStatusMaintenance SegmentStatus = "maintenance"
)

type PipelineSegment struct {
	ID               string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	SiteID           string        `gorm:"not null" json:"site_id"`
	ZoneID           string        `gorm:"not null" json:"zone_id"`
	Label            string        `gorm:"type:varchar(100);not null" json:"label" unique` // Label is unique for each zone
	StartNodeID      string        `gorm:"not null" json:"start_node_id"`
	StopNodeID       string        `gorm:"not null" json:"stop_node_id"`
	PipeID           string        `gorm:"not null" json:"pipe_id"`
	DiameterMM       int           `gorm:"not null" json:"diameter_mm"`
	LengthM          float64       `gorm:"type:decimal(10,2);not null" json:"length_m"`
	InstallationDate *time.Time    `json:"installation_date"`
	Status           SegmentStatus `gorm:"type:segment_status_enum;default:'active'" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// Relationships
	Site      Site `gorm:"foreignKey:SiteID;references:ID;constraint:OnDelete:CASCADE" json:"site"`
	Zone      Zone `gorm:"foreignKey:ZoneID;references:ID;constraint:OnDelete:CASCADE" json:"zone"`
	StartNode Node `gorm:"foreignKey:StartNodeID;references:ID;constraint:OnDelete:RESTRICT" json:"start_node"`
	StopNode  Node `gorm:"foreignKey:StopNodeID;references:ID;constraint:OnDelete:RESTRICT" json:"stop_node"`
	Pipe      Pipe `gorm:"foreignKey:PipeID;references:ID;constraint:OnDelete:RESTRICT" json:"pipe"`
}

func (PipelineSegment) TableName() string {
	return "pipeline_segments"
}

type PipelineSegmentView struct {
	SiteName  string    `json:"site_name"`
	ZoneName  string    `json:"zone_name"`
	Label     string    `json:"label"`
	StartNode string    `json:"start_node"`
	StopNode  string    `json:"stop_node"`
	Diameter  int       `json:"diameter"`
	Length    float64   `json:"length"`
	PipeName  string    `json:"pipe_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
