package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Contractor corresponds to your Dart ContractorModel.
type Contractor struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SiteName          string         `gorm:"not null" json:"siteName"`
	ContractorName    string         `gorm:"not null" json:"contractorName"`
	ContractorPhone   string         `gorm:"not null" json:"contractorPhone"`
	ChainageFrom      string         `gorm:"not null" json:"chainageFrom"`
	ChainageTo        string         `gorm:"not null" json:"chainageTo"`
	ActualMeters      string         `gorm:"not null" json:"actualMeters"`
	DieselTaken       string         `gorm:"not null" json:"dieselTaken"`
	MeterPhotos       pq.StringArray `gorm:"type:text[]" json:"meterPhotos"`
	CardNumber        string         `gorm:"not null" json:"cardNumber"`
	AreaPhotos        pq.StringArray `gorm:"type:text[]" json:"areaPhotos"`
	SiteEngineerName  string         `gorm:"not null" json:"siteEngineerName"`
	SiteEngineerPhone string         `gorm:"not null" json:"siteEngineerPhone"`
	Latitude          float64        `gorm:"not null" json:"latitude"`
	Longitude         float64        `gorm:"not null" json:"longitude"`
	SubmittedAt       JSONTime       `gorm:"not null" json:"submittedAt"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
