package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task corresponds to your Dart TasksModel.
type Task struct {
	ID                     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Label                  string    `gorm:"not null" json:"label"`
	Location               string    `gorm:"not null" json:"location"`
	Measurement            string    `gorm:"not null" json:"measurement"`
	TaskType               string    `gorm:"not null" json:"taskType"`
	ExpectedCompletionDays string    `gorm:"not null" json:"expectedCompletionDays"`
	StartDate              time.Time `gorm:"not null" json:"startDate"`
	EndDate                time.Time `gorm:"not null" json:"endDate"`
	Description            *string   `json:"description,omitempty"`
	PipeMaterial           *string   `json:"pipeMaterial,omitempty"`
	PipeDia                *string   `json:"pipeDia,omitempty"`
	Remarks                *string   `json:"remarks,omitempty"`
	WorkAssignedBy         *string   `json:"workAssignedBy,omitempty"`
	Latitude               float64   `gorm:"not null" json:"latitude"`
	Longitude              float64   `gorm:"not null" json:"longitude"`
	SubmittedAt            time.Time `gorm:"not null" json:"submittedAt"`
	SiteEngineerName       string    `gorm:"not null" json:"siteEngineerName"`
	SiteEngineerPhone      string    `gorm:"not null" json:"siteEngineerPhone"`
	// Example of array type field, if you have photos or attachments:
	// Photos                 pq.StringArray `gorm:"type:text[]" json:"photos,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
