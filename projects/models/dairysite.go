package models

import (
	"time"

	"gorm.io/gorm"
	"p9e.in/ugcl/identity/models"
	pmodels "p9e.in/ugcl/packages/models"
)

type DairySite struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v7()" json:"id"`
	NameOfSite string `json:"nameOfSite"`
	TodaysWork string `json:"todaysWork"`
	// Foreign Key to users table
	EmployeeId string      `gorm:"not null" json:"employee_id"`
	User       models.User `gorm:"foreignKey:EmployeeId;references:Uuid" json:"-"`

	// Deprecated fields – will be removed soon
	SiteEngineerName  string           `gorm:"-" json:"-"`
	SiteEngineerPhone string           `gorm:"-" json:"-"`
	Latitude          float64          `json:"latitude"`
	Longitude         float64          `json:"longitude"`
	SubmittedAt       pmodels.JSONTime `json:"submittedAt"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
