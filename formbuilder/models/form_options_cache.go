package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FormOptionsCache caches dropdown options
type FormOptionsCache struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()"`
	SourceEndpoint   string         `gorm:"type:varchar(500);not null;uniqueIndex:idx_cache_key"`
	CacheKey         string         `gorm:"type:varchar(500);not null;uniqueIndex:idx_cache_key"`
	OptionsData      datatypes.JSON `gorm:"type:jsonb;not null"`
	RequestContext   datatypes.JSON `gorm:"type:jsonb"`
	ResponseMetadata datatypes.JSON `gorm:"type:jsonb"`
	ExpiresAt        time.Time      `gorm:"not null;index"`
	CreatedAt        time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastAccessedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	AccessCount      int            `gorm:"default:1"`
}
