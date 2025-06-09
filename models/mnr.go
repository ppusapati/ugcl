package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MnrReport represents a “MNR” form submission.
type Mnr struct {
	ID                   uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NameOfSite           string         `gorm:"not null" json:"nameOfSite"`
	ZoneName             string         `gorm:"not null" json:"zoneName"`
	WorkDescription      string         `gorm:"not null" json:"workDescription"`
	SkilledLabourCount   string         `gorm:"not null" json:"skilledLabourCount"`
	UnskilledLabourCount string         `gorm:"not null" json:"unskilledLabourCount"`
	WomenCount           string         `gorm:"not null" json:"womenCount"`
	ContractorName       string         `gorm:"not null" json:"contractorName"`
	AttendanceTakenBy    string         `gorm:"not null" json:"attendanceTakenBy"`
	AttendancePhone      string         `gorm:"not null" json:"attendancePhone"`
	WorkPhotos           datatypes.JSON `gorm:"type:jsonb;not null" json:"workPhotos"` // e.g. ["img1.jpg", "img2.png"]
	Remarks              *string        `json:"remarks,omitempty"`
	Latitude             float64        `gorm:"not null" json:"latitude"`
	Longitude            float64        `gorm:"not null" json:"longitude"`
	SubmittedAt          JSONTime       `gorm:"not null" json:"submittedAt"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
