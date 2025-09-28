// =============================================================================
// mappers/job_mapper.go - Import job entity mapping implementation
// =============================================================================
package mappers

import (
	"encoding/json"

	pb "p9e.in/ugcl/databridge/api/databridge"
	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobMapper implements IJobMapper
type JobMapper struct{}

// NewJobMapper creates a new job mapper instance
func NewJobMapper() IJobMapper {
	return &JobMapper{}
}

// DBToProto converts database model to protobuf
func (m *JobMapper) DBToProto(job *db.ImportJob) *pb.ImportJob {
	if job == nil {
		return nil
	}

	pbJob := &pb.ImportJob{
		Id:             job.ID.String(),
		JobName:        job.JobName,
		MappingId:      job.MappingID.String(),
		FileName:       job.FileName,
		FileSize:       job.FileSize,
		Status:         m.StatusDBToProto(job.Status),
		ProcessedRows:  job.ProcessedRows,
		SuccessfulRows: job.SuccessfulRows,
		FailedRows:     job.FailedRows,
		CreatedAt:      timestamppb.New(job.CreatedAt.Time),
		CreatedBy:      job.CreatedBy,
	}

	if job.FileHash.Valid {
		pbJob.FileHash = job.FileHash.String
	}

	if job.TotalRows.Valid {
		totalRows := job.TotalRows.Int32
		pbJob.TotalRows = &totalRows
	}

	if job.StartedAt.Valid {
		pbJob.StartedAt = timestamppb.New(job.StartedAt.Time)
	}

	if job.CompletedAt.Valid {
		pbJob.CompletedAt = timestamppb.New(job.CompletedAt.Time)
	}

	// Parse error details from JSONB
	if len(job.ErrorDetails) > 0 {
		var errorDetails []*pb.ImportError
		if err := json.Unmarshal(job.ErrorDetails, &errorDetails); err == nil {
			pbJob.ErrorDetails = errorDetails
		}
	}

	return pbJob
}

// DetailDBToProto converts detailed job query result to protobuf
func (m *JobMapper) DetailDBToProto(job *db.GetJobByIDRow) *pb.ImportJob {
	if job == nil {
		return nil
	}

	pbJob := &pb.ImportJob{
		Id:             job.ID.String(),
		JobName:        job.JobName,
		MappingId:      job.MappingID.String(),
		FileName:       job.FileName,
		FileSize:       job.FileSize,
		Status:         m.StatusDBToProto(job.Status),
		ProcessedRows:  job.ProcessedRows,
		SuccessfulRows: job.SuccessfulRows,
		FailedRows:     job.FailedRows,
		CreatedAt:      timestamppb.New(job.CreatedAt.Time),
		CreatedBy:      job.CreatedBy,
	}

	if job.FileHash.Valid {
		pbJob.FileHash = job.FileHash.String
	}

	if job.TotalRows.Valid {
		totalRows := job.TotalRows.Int32
		pbJob.TotalRows = &totalRows
	}

	if job.StartedAt.Valid {
		pbJob.StartedAt = timestamppb.New(job.StartedAt.Time)
	}

	if job.CompletedAt.Valid {
		pbJob.CompletedAt = timestamppb.New(job.CompletedAt.Time)
	}

	if job.MappingName.Valid {
		pbJob.MappingName = &job.MappingName.String
	}

	if job.TableName.Valid {
		pbJob.TableName = &job.TableName.String
	}

	if job.TableDisplayName.Valid {
		pbJob.TableDisplayName = &job.TableDisplayName.String
	}

	if job.SchemaName.Valid {
		pbJob.SchemaName = &job.SchemaName.String
	}

	if job.SchemaDisplayName.Valid {
		pbJob.SchemaDisplayName = &job.SchemaDisplayName.String
	}

	// Parse error details from JSONB
	if len(job.ErrorDetails) > 0 {
		var errorDetails []*pb.ImportError
		if err := json.Unmarshal(job.ErrorDetails, &errorDetails); err == nil {
			pbJob.ErrorDetails = errorDetails
		}
	}

	return pbJob
}

// UserJobDBToProto converts user jobs query result to protobuf
func (m *JobMapper) UserJobDBToProto(job *db.GetJobsByUserRow) *pb.ImportJob {
	if job == nil {
		return nil
	}

	pbJob := &pb.ImportJob{
		Id:             job.ID.String(),
		JobName:        job.JobName,
		MappingId:      job.MappingID.String(),
		FileName:       job.FileName,
		FileSize:       job.FileSize,
		Status:         m.StatusDBToProto(job.Status),
		ProcessedRows:  job.ProcessedRows,
		SuccessfulRows: job.SuccessfulRows,
		FailedRows:     job.FailedRows,
		CreatedAt:      timestamppb.New(job.CreatedAt.Time),
		CreatedBy:      job.CreatedBy,
	}

	if job.FileHash.Valid {
		pbJob.FileHash = job.FileHash.String
	}

	if job.TotalRows.Valid {
		totalRows := job.TotalRows.Int32
		pbJob.TotalRows = &totalRows
	}

	if job.StartedAt.Valid {
		pbJob.StartedAt = timestamppb.New(job.StartedAt.Time)
	}

	if job.CompletedAt.Valid {
		pbJob.CompletedAt = timestamppb.New(job.CompletedAt.Time)
	}

	if job.MappingName.Valid {
		pbJob.MappingName = &job.MappingName.String
	}

	if job.TableName.Valid {
		pbJob.TableName = &job.TableName.String
	}

	if job.TableDisplayName.Valid {
		pbJob.TableDisplayName = &job.TableDisplayName.String
	}

	if job.SchemaName.Valid {
		pbJob.SchemaName = &job.SchemaName.String
	}

	if job.SchemaDisplayName.Valid {
		pbJob.SchemaDisplayName = &job.SchemaDisplayName.String
	}

	// Parse error details from JSONB
	if len(job.ErrorDetails) > 0 {
		var errorDetails []*pb.ImportError
		if err := json.Unmarshal(job.ErrorDetails, &errorDetails); err == nil {
			pbJob.ErrorDetails = errorDetails
		}
	}

	return pbJob
}

// ProtoToDB converts protobuf to database model (mainly for reference)
func (m *JobMapper) ProtoToDB(job *pb.ImportJob) *db.ImportJob {
	if job == nil {
		return nil
	}

	// This is mainly for reference - typically we use repository parameter structs
	return &db.ImportJob{
		JobName:        job.JobName,
		FileName:       job.FileName,
		FileSize:       job.FileSize,
		Status:         m.StatusProtoToDB(job.Status),
		ProcessedRows:  job.ProcessedRows,
		SuccessfulRows: job.SuccessfulRows,
		FailedRows:     job.FailedRows,
		CreatedBy:      job.CreatedBy,
	}
}

// CreateRequestToRepoParams converts create request to repository parameters
func (m *JobMapper) CreateRequestToRepoParams(req *pb.CreateImportJobRequest) *repository.CreateJobParams {
	if req == nil {
		return nil
	}

	mappingID, err := uuid.Parse(req.MappingId)
	if err != nil {
		return nil
	}

	params := &repository.CreateJobParams{
		JobName:   req.JobName,
		MappingID: mappingID,
		FileName:  req.FileName,
		FileSize:  req.FileSize,
		CreatedBy: req.CreatedBy,
	}

	if req.FileHash != "" {
		params.FileHash = &req.FileHash
	}

	if req.TotalRows > 0 {
		params.TotalRows = &req.TotalRows
	}

	return params
}

// StatusDBToProto converts database status string to protobuf enum
func (m *JobMapper) StatusDBToProto(status string) pb.ImportJobStatus {
	switch status {
	case "pending":
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_PENDING
	case "processing":
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_PROCESSING
	case "completed":
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_COMPLETED
	case "failed":
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_FAILED
	case "cancelled":
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_CANCELLED
	default:
		return pb.ImportJobStatus_IMPORT_JOB_STATUS_UNSPECIFIED
	}
}

// StatusProtoToDB converts protobuf status enum to database string
func (m *JobMapper) StatusProtoToDB(status pb.ImportJobStatus) string {
	switch status {
	case pb.ImportJobStatus_IMPORT_JOB_STATUS_PENDING:
		return "pending"
	case pb.ImportJobStatus_IMPORT_JOB_STATUS_PROCESSING:
		return "processing"
	case pb.ImportJobStatus_IMPORT_JOB_STATUS_COMPLETED:
		return "completed"
	case pb.ImportJobStatus_IMPORT_JOB_STATUS_FAILED:
		return "failed"
	case pb.ImportJobStatus_IMPORT_JOB_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "pending" // Default fallback
	}
}