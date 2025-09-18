package models

import (
	"time"

	"p9e.in/ugcl/identity/user/models"
	pmodels "p9e.in/ugcl/packages/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WrappingReport represents one “wrapping” form submission.
type Wrapping struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v7();primaryKey" json:"id"`
	YardName string    `gorm:"column:yard_name;not null"              json:"yardName"`

	Activity     string         `gorm:"column:activity;not null"               json:"activity"`
	PipeNo       string         `gorm:"column:pipe_no;not null"                json:"pipeNo"`
	LengthOfPipe string         `gorm:"column:length_of_pipe;not null"         json:"lengthOfPipe"`
	SquareMeters string         `gorm:"column:square_meters;not null"          json:"squareMeters"`
	Photos       datatypes.JSON `gorm:"column:photos;type:jsonb;not null"      json:"photos"` // JSON array of filenames/URLs
	Remarks      *string        `gorm:"column:remarks"                         json:"remarks,omitempty"`

	// Foreign Key to users table
	EmployeeId   string      `gorm:"not null" json:"employee_id"`
	User         models.User `gorm:"foreignKey:EmployeeId;references:UUID" json:"-"`
	ContractorId string      `gorm:"column:contractor_id;not null"        json:"contractorId"`
	Contractor   models.User `gorm:"foreignKey:ContractorId;references:UUID" json:"-"`

	//default from
	Latitude    float64          `gorm:"column:latitude;not null"               json:"latitude"`
	Longitude   float64          `gorm:"column:longitude;not null"              json:"longitude"`
	SubmittedAt pmodels.JSONTime `gorm:"column:submitted_at;not null"           json:"submittedAt"`

	CreatedAt time.Time      `gorm:"autoCreateTime"                         json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                         json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                  json:"-"`
}
