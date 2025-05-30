package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// PaintingReport represents a “painting” form submission.
type Painting struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NameOfYard         *string        `gorm:"column:name_of_yard" json:"nameOfYard,omitempty"`
	ContractorName     *string        `gorm:"column:contractor_name" json:"contractorName,omitempty"`
	WorkDoneActivity   string         `gorm:"column:work_done_activity;not null" json:"workDoneActivity"`
	NumberOfCoats      int            `gorm:"column:number_of_coats;not null" json:"numberOfCoats"`
	DiaOfPipe          string         `gorm:"column:dia_of_pipe;not null" json:"diaOfPipe"`
	PipeNo             string         `gorm:"column:pipe_no;not null" json:"pipeNo"`
	LengthOfPipe       string         `gorm:"column:length_of_pipe;not null" json:"lengthOfPipe"`
	SquareMeters       string         `gorm:"column:square_meters;not null" json:"squareMeters"`
	PhotoOfPaintedPipe pq.StringArray `gorm:"column:photo_of_painted_pipe;type:text[];not null" json:"photoOfPaintedPipe"`
	Remarks            *string        `gorm:"column:remarks" json:"remarks,omitempty"`
	SiteEngineerName   string         `gorm:"column:site_engineer_name" json:"siteEngineerName,omitempty"`
	PhoneNumber        string         `gorm:"column:phone_number" json:"phoneNumber,omitempty"`
	Latitude           float64        `gorm:"column:latitude;not null" json:"latitude"`
	Longitude          float64        `gorm:"column:longitude;not null" json:"longitude"`
	SubmittedAt        JSONTime       `gorm:"column:submitted_at;not null" json:"submittedAt"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
