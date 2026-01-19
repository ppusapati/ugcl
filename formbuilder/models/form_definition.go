package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Base model with common fields
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FormDefinition stores form schemas
type FormDefinition struct {
	BaseModel
	FormID              string         `gorm:"type:varchar(255);not null;index:idx_form_version,unique"`
	Title               string         `gorm:"type:varchar(500);not null"`
	Description         string         `gorm:"type:text"`
	Version             string         `gorm:"type:varchar(50);not null;index:idx_form_version,unique"`
	SchemaVersion       int            `gorm:"not null;default:1"`
	Definition          datatypes.JSON `gorm:"type:jsonb;not null"`
	TableName           string         `gorm:"type:varchar(255)"`
	CoreFields          StringArray    `gorm:"type:text[]"`
	CreatedBy           string         `gorm:"type:varchar(255);not null"`
	IsActive            bool           `gorm:"default:true;index"`
	IsPublished         bool           `gorm:"default:false;index"`
	Metadata            datatypes.JSON `gorm:"type:jsonb"`
	WorkflowIntegration datatypes.JSON `gorm:"type:jsonb"`
}

// FormVersionMigration tracks schema migrations
type FormVersionMigration struct {
	BaseModel
	FormID          string    `gorm:"type:varchar(255);not null;index"`
	FromVersion     int       `gorm:"not null"`
	ToVersion       int       `gorm:"not null"`
	MigrationType   string    `gorm:"type:varchar(50);not null"` // 'schema', 'data', 'both'
	StartedAt       time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	CompletedAt     *time.Time
	RecordsAffected int            `gorm:"default:0"`
	Status          string         `gorm:"type:varchar(50);default:'in_progress';index"`
	PerformedBy     string         `gorm:"type:varchar(255);not null"`
	MigrationScript string         `gorm:"type:text"`
	RollbackScript  string         `gorm:"type:text"`
	Errors          datatypes.JSON `gorm:"type:jsonb"`
	Metadata        datatypes.JSON `gorm:"type:jsonb"`
}

// StringArray handles PostgreSQL text[] type
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return json.Unmarshal([]byte("[]"), a)
	}
}

// Dynamic form table model - created at runtime
type DynamicFormData struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()"`
	FormID       string         `gorm:"type:varchar(255);not null;index"`
	FormVersion  int            `gorm:"not null"`
	CurrentState string         `gorm:"type:varchar(100);index"`
	CreatedBy    string         `gorm:"type:varchar(255)"`
	AssignedTo   string         `gorm:"type:varchar(255);index"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	ExtraData    datatypes.JSON `gorm:"type:jsonb"` // For version-specific fields
	// Dynamic fields are added based on form definition
}

// Hooks
