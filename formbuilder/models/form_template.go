package models

import (
	"gorm.io/datatypes"
)

// FormTemplate stores reusable templates
type FormTemplate struct {
	BaseModel
	TemplateID         string         `gorm:"type:varchar(255);unique;not null"`
	Name               string         `gorm:"type:varchar(500);not null"`
	Description        string         `gorm:"type:text"`
	Category           string         `gorm:"type:varchar(255);index"`
	TemplateDefinition datatypes.JSON `gorm:"type:jsonb;not null"`
	IsActive           bool           `gorm:"default:true;index"`
	CreatedBy          string         `gorm:"type:varchar(255);not null"`
	Metadata           datatypes.JSON `gorm:"type:jsonb"`
}
