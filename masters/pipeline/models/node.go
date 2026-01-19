package models

import (
	"time"

	"gorm.io/gorm"
)

type NodeType string

const (
	NodeTypeJunction NodeType = "junction"
	NodeTypeTank     NodeType = "tank"
	NodeTypeSource   NodeType = "source"
	NodeTypeEndpoint NodeType = "endpoint"
)

type Node struct {
	ID       string   `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	NodeName string   `gorm:"type:varchar(100);not null" json:"node_name"`
	SiteID   string   `gorm:"not null" json:"site_id"`
	ZoneID   *string  `json:"zone_id"`
	NodeType NodeType `gorm:"type:node_type_enum;default:'junction'" json:"node_type"`

	CoordinatesX *float64 `gorm:"type:decimal(10,6)" json:"coordinates_x"`
	CoordinatesY *float64 `gorm:"type:decimal(10,6)" json:"coordinates_y"`
	Elevation    *float64 `gorm:"type:decimal(8,2)" json:"elevation"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// Relationships
	Site Site  `gorm:"foreignKey:SiteID;references:ID;constraint:OnDelete:CASCADE" json:"site,omitempty"`
	Zone *Zone `gorm:"foreignKey:ZoneID;references:ID;constraint:OnDelete:SET NULL" json:"zone,omitempty"`

	// Pipeline segments where this node is start or stop
	StartSegments []PipelineSegment `gorm:"foreignKey:StartNodeID;constraint:OnDelete:RESTRICT" json:"start_segments,omitempty"`
	StopSegments  []PipelineSegment `gorm:"foreignKey:StopNodeID;constraint:OnDelete:RESTRICT" json:"stop_segments,omitempty"`
}

func (Node) TableName() string {
	return "nodes"
}
