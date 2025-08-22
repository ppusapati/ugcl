package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// APIAccessLog tracks API usage
type APIAccessLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()"`
	UserID         string         `gorm:"type:varchar(255);not null;index"`
	APIEndpoint    string         `gorm:"type:varchar(500);not null;index"`
	HTTPMethod     string         `gorm:"type:varchar(10);not null"`
	RequestBody    datatypes.JSON `gorm:"type:jsonb"`
	ResponseStatus int
	ResponseTimeMs int
	IPAddress      string     `gorm:"type:varchar(45)"`
	UserAgent      string     `gorm:"type:text"`
	FormID         string     `gorm:"type:varchar(255)"`
	InstanceID     *uuid.UUID `gorm:"type:uuid"`
	Timestamp      time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
}
