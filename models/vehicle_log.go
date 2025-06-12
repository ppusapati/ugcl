package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type VehicleLog struct {
	ID                   uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email                string         `gorm:"type:varchar(255);not null" json:"email" validate:"required,email"`
	SiteLocation         string         `gorm:"type:varchar(255);not null" json:"site_location"`
	WorkingZone          string         `gorm:"type:varchar(255)" json:"working_zone,omitempty"`
	Date                 time.Time      `gorm:"type:timestamp;not null" json:"date" validate:"required"`
	VehicleType          string         `gorm:"type:varchar(255);not null" json:"vehicle_type"`
	RegistrationNumber   string         `gorm:"type:varchar(255)" json:"registration_number,omitempty"`
	OwnerName            string         `gorm:"type:varchar(255)" json:"owner_name,omitempty"`
	DriverName           string         `gorm:"type:varchar(255)" json:"driver_name,omitempty"`
	StartingReadingFiles pq.StringArray `gorm:"type:text[]" json:"starting_reading_files,omitempty"` // Postgres array
	ClosingReadingFiles  pq.StringArray `gorm:"type:text[]" json:"closing_reading_files,omitempty"`
	ReadingTotalKMHrs    string         `gorm:"type:varchar(100)" json:"reading_total_km_hrs,omitempty"`
	TotalWorkingHours    string         `gorm:"type:varchar(100)" json:"total_working_hours,omitempty"`
	DieselIssuedLitres   string         `gorm:"type:varchar(100)" json:"diesel_issued_litres,omitempty"`
	WorkDescription      string         `gorm:"type:text" json:"work_description,omitempty"`
	WorkImages           pq.StringArray `gorm:"type:text[]" json:"work_images,omitempty"`
	Remarks              string         `gorm:"type:text" json:"remarks,omitempty"`
	SiteEngineerName     string         `gorm:"not null" json:"siteEngineerName"`
	SiteEngineerPhone    string         `gorm:"not null" json:"siteEngineerPhone"`
	Latitude             float64        `gorm:"not null" json:"latitude"`
	Longitude            float64        `gorm:"not null" json:"longitude"`
	SubmittedAt          JSONTime       `gorm:"not null" json:"submittedAt"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// Optionally, add TableName() for custom table name
func (VehicleLog) TableName() string {
	return "vehicle_logs"
}
