// =============================================================================
// services/iservices.go - Service interfaces for DataBridge module (Import Only)
// =============================================================================
package services

import (
	"context"
	"io"

	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/models"

	"github.com/google/uuid"
)

// IMappingService defines import mapping service interface
type IMappingService interface {
	GetMappings(ctx context.Context, req *pb.GetMappingsRequest) (*pb.GetMappingsResponse, error)
	GetMappingByID(ctx context.Context, mappingID uuid.UUID) (*pb.ImportMapping, error)
	CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.ImportMapping, error)
	UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.ImportMapping, error)
	DeleteMapping(ctx context.Context, mappingID uuid.UUID, userID string) error
	ValidateMapping(ctx context.Context, req *pb.ValidateMappingRequest) (*pb.ValidateMappingResponse, error)
}

// ICSVService defines CSV processing service interface
type ICSVService interface {
	AnalyzeCSV(ctx context.Context, req *pb.AnalyzeCSVRequest) (*pb.CSVAnalysis, error)
	ValidateCSVWithMapping(ctx context.Context, csvData []byte, mapping *pb.ImportMapping) (*models.ValidationResult, error)
	PreviewImport(ctx context.Context, req *pb.PreviewImportRequest) (*pb.PreviewImportResponse, error)
	ParseCSVHeaders(csvData []byte, delimiter string, hasHeader bool) ([]string, error)
	InferColumnTypes(csvData []byte, headers []string, delimiter string) ([]*models.ColumnTypeInference, error)
}

// IImportService defines import execution service interface
type IImportService interface {
	CreateImportJob(ctx context.Context, req *pb.CreateImportJobRequest) (*pb.ImportJob, error)
	GetImportJob(ctx context.Context, jobID uuid.UUID) (*pb.ImportJob, error)
	GetImportJobs(ctx context.Context, req *pb.GetImportJobsRequest) (*pb.GetImportJobsResponse, error)
	CancelImportJob(ctx context.Context, jobID uuid.UUID) error
	ExecuteImport(ctx context.Context, jobID uuid.UUID, csvData []byte, progressChan chan<- *pb.ImportProgress) error
}

// IValidationService defines data validation service interface
type IValidationService interface {
	ValidateFieldValue(value string, column *pb.Column) error
	ValidateRequiredFields(data map[string]string, requiredColumns []*pb.Column) []string
	TransformFieldValue(value string, transformation string, dataType pb.DataType) (interface{}, error)
	ValidateDataTypes(data map[string]string, columns []*pb.Column) []*models.ValidationError
	ApplyBusinessRules(data map[string]string, rules []*pb.ValidationRule) []*models.ValidationError
}

// IFileService defines file handling service interface
type IFileService interface {
	ValidateFileSize(size int64) error
	ValidateFileType(filename string) error
	CalculateFileHash(data []byte) string
	DetectCSVEncoding(data []byte) (string, error)
	DetectCSVDelimiter(data []byte) (string, error)
	ReadCSVStream(reader io.Reader, delimiter string, hasHeader bool) (<-chan []string, <-chan error)
}

// ServiceManager combines all service interfaces for DataBridge (Import-focused)
type ServiceManager struct {
	Mapping    IMappingService
	CSV        ICSVService
	Import     IImportService
	Validation IValidationService
	File       IFileService
}

// NewServiceManager creates a new service manager instance for import services
func NewServiceManager(
	mapping IMappingService,
	csv ICSVService,
	importSvc IImportService,
	validation IValidationService,
	file IFileService,
) *ServiceManager {
	return &ServiceManager{
		Mapping:    mapping,
		CSV:        csv,
		Import:     importSvc,
		Validation: validation,
		File:       file,
	}
}