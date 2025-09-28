package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WatermarkConfig represents watermark configuration
type WatermarkConfig struct {
	BaseModel
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name         string         `gorm:"type:varchar(255);not null;unique"`
	Description  string         `gorm:"type:text"`
	IsDefault    bool           `gorm:"default:false"`
	IsActive     bool           `gorm:"default:true"`

	// Watermark content
	Type         string         `gorm:"type:varchar(50);not null"` // text, image, combined
	TextContent  string         `gorm:"type:text"`
	ImagePath    string         `gorm:"type:text"`
	LogoPath     string         `gorm:"type:text"`

	// Dynamic content placeholders
	UseDynamicText    bool      `gorm:"default:true"`
	ShowUsername      bool      `gorm:"default:true"`
	ShowTimestamp     bool      `gorm:"default:true"`
	ShowCompanyLogo   bool      `gorm:"default:false"`
	ShowDocumentInfo  bool      `gorm:"default:false"`
	ShowIPAddress     bool      `gorm:"default:false"`

	// Positioning
	Position     string         `gorm:"type:varchar(50);default:'bottom_right'"` // top_left, top_right, bottom_left, bottom_right, center
	OffsetX      int            `gorm:"default:10"`
	OffsetY      int            `gorm:"default:10"`
	Rotation     float64        `gorm:"default:0"`

	// Appearance
	FontFamily   string         `gorm:"type:varchar(100);default:'Arial'"`
	FontSize     int            `gorm:"default:12"`
	FontColor    string         `gorm:"type:varchar(7);default:'#000000'"` // hex color
	FontBold     bool           `gorm:"default:false"`
	FontItalic   bool           `gorm:"default:false"`

	// Transparency and effects
	Opacity      float64        `gorm:"default:0.5"`
	BlendMode    string         `gorm:"type:varchar(50);default:'normal'"`
	Shadow       bool           `gorm:"default:false"`
	ShadowColor  string         `gorm:"type:varchar(7);default:'#808080'"`
	ShadowOffsetX int           `gorm:"default:2"`
	ShadowOffsetY int           `gorm:"default:2"`

	// Size constraints
	MaxWidth     int            `gorm:"default:200"`
	MaxHeight    int            `gorm:"default:50"`
	Scale        float64        `gorm:"default:1.0"`

	// Page settings
	ApplyToAllPages bool         `gorm:"default:true"`
	PageNumbers     datatypes.JSON `gorm:"type:jsonb"` // specific pages if not all

	// Template for dynamic text
	TextTemplate string         `gorm:"type:text;default:'{{.Username}} - {{.Timestamp}}'"`

	// Security settings
	PreventRemoval bool          `gorm:"default:true"`
	EncryptWatermark bool        `gorm:"default:false"`

	// Usage tracking
	UsageCount   int64          `gorm:"default:0"`
	LastUsedAt   *time.Time

	// Owner and permissions
	CreatedBy    string         `gorm:"type:varchar(255);not null"`
	UpdatedBy    string         `gorm:"type:varchar(255)"`
	Permissions  datatypes.JSON `gorm:"type:jsonb"`

	// Additional metadata
	Tags         datatypes.JSON `gorm:"type:jsonb"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
}

// WatermarkTemplate represents predefined watermark templates
type WatermarkTemplate struct {
	BaseModel
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name         string         `gorm:"type:varchar(255);not null;unique"`
	Description  string         `gorm:"type:text"`
	Category     string         `gorm:"type:varchar(100);index"`
	IsBuiltIn    bool           `gorm:"default:false"`
	IsPublic     bool           `gorm:"default:false"`

	// Template configuration (JSON of WatermarkConfig)
	Config       datatypes.JSON `gorm:"type:jsonb"`

	// Preview image
	PreviewPath  string         `gorm:"type:text"`

	// Usage and ratings
	UsageCount   int64          `gorm:"default:0"`
	Rating       float64        `gorm:"default:0"`
	RatingCount  int64          `gorm:"default:0"`

	// Owner
	CreatedBy    string         `gorm:"type:varchar(255);not null"`
	Tags         datatypes.JSON `gorm:"type:jsonb"`
}

// WatermarkJob represents watermarking jobs
type WatermarkJob struct {
	BaseModel
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DocumentID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	ConfigID        uuid.UUID      `gorm:"type:uuid;index"`
	Status          string         `gorm:"type:varchar(50);default:'pending'"`
	Priority        int            `gorm:"default:0"`

	// Job details
	RequestedBy     string         `gorm:"type:varchar(255);not null"`
	JobType         string         `gorm:"type:varchar(50);not null"` // preview, download, print
	OutputPath      string         `gorm:"type:text"`

	// Dynamic data for watermark
	DynamicData     datatypes.JSON `gorm:"type:jsonb"`

	// Processing info
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Progress        int            `gorm:"default:0"`
	ErrorMessage    string         `gorm:"type:text"`
	ProcessingTime  int64          `gorm:"default:0"` // milliseconds

	// Relationships
	Document        Document       `gorm:"foreignKey:DocumentID"`
	Config          WatermarkConfig `gorm:"foreignKey:ConfigID"`
}

// WatermarkUsage tracks watermark usage for analytics
type WatermarkUsage struct {
	BaseModel
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConfigID     uuid.UUID `gorm:"type:uuid;not null;index"`
	DocumentID   uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID       string    `gorm:"type:varchar(255);not null;index"`
	Action       string    `gorm:"type:varchar(50);not null"` // preview, download, print
	IPAddress    string    `gorm:"type:varchar(45)"`
	UserAgent    string    `gorm:"type:text"`
	Success      bool      `gorm:"default:true"`
	ErrorMessage string    `gorm:"type:text"`

	// Applied watermark details
	WatermarkData datatypes.JSON `gorm:"type:jsonb"`

	// Relationships
	Config       WatermarkConfig `gorm:"foreignKey:ConfigID"`
	Document     Document        `gorm:"foreignKey:DocumentID"`
}

// TableName specifies the table name for WatermarkConfig
func (WatermarkConfig) TableName() string {
	return "watermark_configs"
}

// TableName specifies the table name for WatermarkTemplate
func (WatermarkTemplate) TableName() string {
	return "watermark_templates"
}

// TableName specifies the table name for WatermarkJob
func (WatermarkJob) TableName() string {
	return "watermark_jobs"
}

// TableName specifies the table name for WatermarkUsage
func (WatermarkUsage) TableName() string {
	return "watermark_usage"
}