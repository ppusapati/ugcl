package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Document represents a document in the system
type Document struct {
	BaseModel
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	FileName        string         `gorm:"type:varchar(500);not null;index"`
	OriginalName    string         `gorm:"type:varchar(500);not null"`
	Title           string         `gorm:"type:varchar(1000)"`
	Description     string         `gorm:"type:text"`
	MimeType        string         `gorm:"type:varchar(255);not null;index"`
	FileExtension   string         `gorm:"type:varchar(10);index"`
	SizeBytes       int64          `gorm:"not null"`
	CompressedSize  int64          `gorm:"default:0"`
	Checksum        string         `gorm:"type:varchar(64);unique;not null"`
	StoragePath     string         `gorm:"type:text;not null"`
	CompressedPath  string         `gorm:"type:text"`
	ThumbnailPath   string         `gorm:"type:text"`
	PreviewPath     string         `gorm:"type:text"`

	// Document specific fields
	PageCount       int            `gorm:"default:0"`
	WordCount       int            `gorm:"default:0"`
	ExtractedText   string         `gorm:"type:text"`
	Language        string         `gorm:"type:varchar(10)"`
	Author          string         `gorm:"type:varchar(255)"`
	Subject         string         `gorm:"type:varchar(500)"`
	Keywords        string         `gorm:"type:text"`

	// Processing status
	ProcessingStatus string        `gorm:"type:varchar(50);default:'pending'"`
	ProcessingError  string        `gorm:"type:text"`
	ProcessedAt      *time.Time

	// Compression info
	CompressionType  string        `gorm:"type:varchar(50);default:'tar_zstd'"`
	CompressionRatio float64       `gorm:"default:0"`

	// OCR info
	OCRStatus        string        `gorm:"type:varchar(50);default:'pending'"`
	OCRConfidence    float64       `gorm:"default:0"`
	OCRProcessedAt   *time.Time

	// Access control
	UploadedBy       string        `gorm:"type:varchar(255);not null;index"`
	Permissions      datatypes.JSON `gorm:"type:jsonb"`
	Tags             datatypes.JSON `gorm:"type:jsonb"`
	Categories       datatypes.JSON `gorm:"type:jsonb"`

	// Security
	VirusScanStatus  string        `gorm:"type:varchar(50);default:'pending'"`
	VirusScanResult  string        `gorm:"type:varchar(50)"`
	VirusScanAt      *time.Time

	// Watermark info
	WatermarkConfig  datatypes.JSON `gorm:"type:jsonb"`
	HasWatermark     bool           `gorm:"default:false"`

	// Metadata
	Metadata         datatypes.JSON `gorm:"type:jsonb"`

	// Soft delete
	IsDeleted        bool           `gorm:"default:false;index"`
	DeletedAt        *time.Time
	DeletedBy        string         `gorm:"type:varchar(255)"`

	// Relationships
	Versions         []DocumentVersion `gorm:"foreignKey:DocumentID"`
	AccessLogs       []DocumentAccessLog `gorm:"foreignKey:DocumentID"`
}

// BaseModel provides common fields for all models
type BaseModel struct {
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

// DocumentVersion represents different versions of a document
type DocumentVersion struct {
	BaseModel
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DocumentID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Version      int       `gorm:"not null"`
	FileName     string    `gorm:"type:varchar(500);not null"`
	StoragePath  string    `gorm:"type:text;not null"`
	SizeBytes    int64     `gorm:"not null"`
	Checksum     string    `gorm:"type:varchar(64);not null"`
	ChangeLog    string    `gorm:"type:text"`
	CreatedBy    string    `gorm:"type:varchar(255);not null"`
	IsActive     bool      `gorm:"default:false"`

	// Relationship
	Document     Document  `gorm:"foreignKey:DocumentID"`
}

// DocumentAccessLog tracks document access for audit purposes
type DocumentAccessLog struct {
	BaseModel
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DocumentID   uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID       string    `gorm:"type:varchar(255);not null;index"`
	Action       string    `gorm:"type:varchar(50);not null"` // view, download, print, etc.
	IPAddress    string    `gorm:"type:varchar(45)"`
	UserAgent    string    `gorm:"type:text"`
	Duration     int64     `gorm:"default:0"` // in milliseconds
	Success      bool      `gorm:"default:true"`
	ErrorMessage string    `gorm:"type:text"`

	// Watermark info for this access
	WatermarkApplied bool           `gorm:"default:false"`
	WatermarkData    datatypes.JSON `gorm:"type:jsonb"`

	// Relationship
	Document     Document  `gorm:"foreignKey:DocumentID"`
}

// ProcessingJob represents background processing jobs
type ProcessingJob struct {
	BaseModel
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DocumentID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	JobType      string         `gorm:"type:varchar(50);not null"` // compress, thumbnail, ocr, etc.
	Status       string         `gorm:"type:varchar(50);default:'pending'"`
	Priority     int            `gorm:"default:0"`
	Progress     int            `gorm:"default:0"` // percentage 0-100
	StartedAt    *time.Time
	CompletedAt  *time.Time
	ErrorMessage string         `gorm:"type:text"`
	JobData      datatypes.JSON `gorm:"type:jsonb"`
	Result       datatypes.JSON `gorm:"type:jsonb"`

	// Relationship
	Document     Document       `gorm:"foreignKey:DocumentID"`
}

// TableName specifies the table name for Document
func (Document) TableName() string {
	return "documents"
}

// TableName specifies the table name for DocumentVersion
func (DocumentVersion) TableName() string {
	return "document_versions"
}

// TableName specifies the table name for DocumentAccessLog
func (DocumentAccessLog) TableName() string {
	return "document_access_logs"
}

// TableName specifies the table name for ProcessingJob
func (ProcessingJob) TableName() string {
	return "document_processing_jobs"
}