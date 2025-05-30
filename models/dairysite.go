package models

import (
	"time"

	"gorm.io/gorm"
)

type DairySite struct {
	ID                string   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	NameOfSite        string   `json:"nameOfSite"`
	TodaysWork        string   `json:"todaysWork"`
	SiteEngineerName  string   `json:"siteEngineerName"`
	SiteEngineerPhone string   `json:"siteEngineerPhone"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	SubmittedAt       JSONTime `json:"submittedAt"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
