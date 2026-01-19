// =============================================================================
// handlers/databridge_handler.go - Main DataBridge ConnectRPC handler
// =============================================================================
package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/api/databridge/databridgeconnect"
	"p9e.in/ugcl/databridge/services"
	"p9e.in/ugcl/packages/p9log"

	"github.com/google/uuid"
)

// DataBridgeHandler implements the DataBridge ConnectRPC service
type DataBridgeHandler struct {
	services *services.ServiceManager
	logger   p9log.Logger
}

// NewDataBridgeHandler creates a new DataBridge handler
func NewDataBridgeHandler(services *services.ServiceManager, logger p9log.Logger) *DataBridgeHandler {
	return &DataBridgeHandler{
		services: services,
		logger:   logger,
	}
}

// Ensure DataBridgeHandler implements the ConnectRPC interface
var _ databridgeconnect.DataBridgeServiceHandler = (*DataBridgeHandler)(nil)

// Schema and Table Management

// GetSchemas retrieves all active schemas
func (h *DataBridgeHandler) GetSchemas(
	ctx context.Context,
	req *connect.Request[pb.GetSchemasRequest],
) (*connect.Response[pb.GetSchemasResponse], error) {
	h.logger.Info("GetSchemas request received")

	schemas, err := h.services.Metadata.GetSchemas(ctx)
	if err != nil {
		h.logger.Error("Failed to get schemas", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get schemas: %w", err))
	}

	h.logger.Info("GetSchemas request completed", "count", len(schemas))

	return connect.NewResponse(&pb.GetSchemasResponse{
		Schemas: schemas,
	}), nil
}

// GetTables retrieves tables based on criteria
func (h *DataBridgeHandler) GetTables(
	ctx context.Context,
	req *connect.Request[pb.GetTablesRequest],
) (*connect.Response[pb.GetTablesResponse], error) {
	h.logger.Info("GetTables request received", "schema_id", req.Msg.SchemaId, "importable_only", req.Msg.ImportableOnly)

	var schemaID *uuid.UUID
	if req.Msg.SchemaId != nil {
		parsedID, err := uuid.Parse(*req.Msg.SchemaId)
		if err != nil {
			h.logger.Error("Invalid schema ID", "schema_id", *req.Msg.SchemaId, "error", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid schema ID: %w", err))
		}
		schemaID = &parsedID
	}

	tables, err := h.services.Metadata.GetTables(ctx, schemaID, req.Msg.ImportableOnly)
	if err != nil {
		h.logger.Error("Failed to get tables", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get tables: %w", err))
	}

	h.logger.Info("GetTables request completed", "count", len(tables))

	return connect.NewResponse(&pb.GetTablesResponse{
		Tables: tables,
	}), nil
}

// GetColumns retrieves columns for a table
func (h *DataBridgeHandler) GetColumns(
	ctx context.Context,
	req *connect.Request[pb.GetColumnsRequest],
) (*connect.Response[pb.GetColumnsResponse], error) {
	h.logger.Info("GetColumns request received", "table_id", req.Msg.TableId, "importable_only", req.Msg.ImportableOnly)

	tableID, err := uuid.Parse(req.Msg.TableId)
	if err != nil {
		h.logger.Error("Invalid table ID", "table_id", req.Msg.TableId, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid table ID: %w", err))
	}

	columns, err := h.services.Metadata.GetColumns(ctx, tableID, req.Msg.ImportableOnly)
	if err != nil {
		h.logger.Error("Failed to get columns", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get columns: %w", err))
	}

	h.logger.Info("GetColumns request completed", "count", len(columns))

	return connect.NewResponse(&pb.GetColumnsResponse{
		Columns: columns,
	}), nil
}

// Import Mapping Management

// GetMappings retrieves import mappings
func (h *DataBridgeHandler) GetMappings(
	ctx context.Context,
	req *connect.Request[pb.GetMappingsRequest],
) (*connect.Response[pb.GetMappingsResponse], error) {
	h.logger.Info("GetMappings request received", "table_id", req.Msg.TableId, "user_id", req.Msg.UserId)

	response, err := h.services.Mapping.GetMappings(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to get mappings", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get mappings: %w", err))
	}

	h.logger.Info("GetMappings request completed", "count", len(response.Mappings))

	return connect.NewResponse(response), nil
}

// GetMappingByID retrieves a specific mapping by ID
func (h *DataBridgeHandler) GetMappingByID(
	ctx context.Context,
	req *connect.Request[pb.GetMappingByIDRequest],
) (*connect.Response[pb.GetMappingByIDResponse], error) {
	h.logger.Info("GetMappingByID request received", "mapping_id", req.Msg.MappingId)

	mappingID, err := uuid.Parse(req.Msg.MappingId)
	if err != nil {
		h.logger.Error("Invalid mapping ID", "mapping_id", req.Msg.MappingId, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid mapping ID: %w", err))
	}

	mapping, err := h.services.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil {
		h.logger.Error("Failed to get mapping", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get mapping: %w", err))
	}

	h.logger.Info("GetMappingByID request completed")

	return connect.NewResponse(&pb.GetMappingByIDResponse{
		Mapping: mapping,
	}), nil
}

// CreateMapping creates a new import mapping
func (h *DataBridgeHandler) CreateMapping(
	ctx context.Context,
	req *connect.Request[pb.CreateMappingRequest],
) (*connect.Response[pb.CreateMappingResponse], error) {
	h.logger.Info("CreateMapping request received", "mapping_name", req.Msg.MappingName, "table_id", req.Msg.TableId)

	mapping, err := h.services.Mapping.CreateMapping(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to create mapping", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create mapping: %w", err))
	}

	h.logger.Info("CreateMapping request completed", "mapping_id", mapping.Id)

	return connect.NewResponse(&pb.CreateMappingResponse{
		Mapping: mapping,
	}), nil
}

// UpdateMapping updates an existing import mapping
func (h *DataBridgeHandler) UpdateMapping(
	ctx context.Context,
	req *connect.Request[pb.UpdateMappingRequest],
) (*connect.Response[pb.UpdateMappingResponse], error) {
	h.logger.Info("UpdateMapping request received", "mapping_id", req.Msg.MappingId)

	mapping, err := h.services.Mapping.UpdateMapping(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to update mapping", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update mapping: %w", err))
	}

	h.logger.Info("UpdateMapping request completed")

	return connect.NewResponse(&pb.UpdateMappingResponse{
		Mapping: mapping,
	}), nil
}

// DeleteMapping deletes an import mapping
func (h *DataBridgeHandler) DeleteMapping(
	ctx context.Context,
	req *connect.Request[pb.DeleteMappingRequest],
) (*connect.Response[pb.DeleteMappingResponse], error) {
	h.logger.Info("DeleteMapping request received", "mapping_id", req.Msg.MappingId)

	mappingID, err := uuid.Parse(req.Msg.MappingId)
	if err != nil {
		h.logger.Error("Invalid mapping ID", "mapping_id", req.Msg.MappingId, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid mapping ID: %w", err))
	}

	err = h.services.Mapping.DeleteMapping(ctx, mappingID, req.Msg.UpdatedBy)
	if err != nil {
		h.logger.Error("Failed to delete mapping", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete mapping: %w", err))
	}

	h.logger.Info("DeleteMapping request completed")

	return connect.NewResponse(&pb.DeleteMappingResponse{}), nil
}

// CSV Processing

// AnalyzeCSV performs comprehensive analysis of CSV data
func (h *DataBridgeHandler) AnalyzeCSV(
	ctx context.Context,
	req *connect.Request[pb.AnalyzeCSVRequest],
) (*connect.Response[pb.AnalyzeCSVResponse], error) {
	h.logger.Info("AnalyzeCSV request received", "filename", req.Msg.FileName, "size", len(req.Msg.CsvData))

	// Validate file size
	if err := h.services.File.ValidateFileSize(int64(len(req.Msg.CsvData))); err != nil {
		h.logger.Error("File size validation failed", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("file size validation failed: %w", err))
	}

	// Validate file type
	if err := h.services.File.ValidateFileType(req.Msg.FileName); err != nil {
		h.logger.Error("File type validation failed", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("file type validation failed: %w", err))
	}

	analysis, err := h.services.CSV.AnalyzeCSV(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to analyze CSV", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to analyze CSV: %w", err))
	}

	h.logger.Info("AnalyzeCSV request completed", "headers", len(analysis.Headers), "rows", analysis.TotalRows)

	return connect.NewResponse(&pb.AnalyzeCSVResponse{
		Analysis: analysis,
	}), nil
}

// ValidateMapping validates CSV data against an import mapping
func (h *DataBridgeHandler) ValidateMapping(
	ctx context.Context,
	req *connect.Request[pb.ValidateMappingRequest],
) (*connect.Response[pb.ValidateMappingResponse], error) {
	h.logger.Info("ValidateMapping request received", "mapping_id", req.Msg.MappingId)

	response, err := h.services.Mapping.ValidateMapping(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to validate mapping", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to validate mapping: %w", err))
	}

	h.logger.Info("ValidateMapping request completed", "valid", response.IsValid, "errors", len(response.Errors))

	return connect.NewResponse(response), nil
}

// PreviewImport generates a preview of the import operation
func (h *DataBridgeHandler) PreviewImport(
	ctx context.Context,
	req *connect.Request[pb.PreviewImportRequest],
) (*connect.Response[pb.PreviewImportResponse], error) {
	h.logger.Info("PreviewImport request received", "mapping_id", req.Msg.MappingId, "preview_rows", req.Msg.PreviewRows)

	response, err := h.services.CSV.PreviewImport(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to preview import", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to preview import: %w", err))
	}

	h.logger.Info("PreviewImport request completed", "preview_rows", len(response.PreviewRows))

	return connect.NewResponse(response), nil
}

// Import Job Management

// CreateImportJob creates a new import job
func (h *DataBridgeHandler) CreateImportJob(
	ctx context.Context,
	req *connect.Request[pb.CreateImportJobRequest],
) (*connect.Response[pb.CreateImportJobResponse], error) {
	h.logger.Info("CreateImportJob request received", "job_name", req.Msg.JobName, "mapping_id", req.Msg.MappingId)

	job, err := h.services.Import.CreateImportJob(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to create import job", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create import job: %w", err))
	}

	h.logger.Info("CreateImportJob request completed", "job_id", job.Id)

	return connect.NewResponse(&pb.CreateImportJobResponse{
		Job: job,
	}), nil
}

// GetImportJob retrieves an import job by ID
func (h *DataBridgeHandler) GetImportJob(
	ctx context.Context,
	req *connect.Request[pb.GetImportJobRequest],
) (*connect.Response[pb.GetImportJobResponse], error) {
	h.logger.Info("GetImportJob request received", "job_id", req.Msg.JobId)

	jobID, err := uuid.Parse(req.Msg.JobId)
	if err != nil {
		h.logger.Error("Invalid job ID", "job_id", req.Msg.JobId, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	job, err := h.services.Import.GetImportJob(ctx, jobID)
	if err != nil {
		h.logger.Error("Failed to get import job", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get import job: %w", err))
	}

	h.logger.Info("GetImportJob request completed")

	return connect.NewResponse(&pb.GetImportJobResponse{
		Job: job,
	}), nil
}

// GetImportJobs retrieves import jobs with pagination
func (h *DataBridgeHandler) GetImportJobs(
	ctx context.Context,
	req *connect.Request[pb.GetImportJobsRequest],
) (*connect.Response[pb.GetImportJobsResponse], error) {
	h.logger.Info("GetImportJobs request received", "user_id", req.Msg.UserId, "status", req.Msg.Status)

	response, err := h.services.Import.GetImportJobs(ctx, req.Msg)
	if err != nil {
		h.logger.Error("Failed to get import jobs", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get import jobs: %w", err))
	}

	h.logger.Info("GetImportJobs request completed", "count", len(response.Jobs))

	return connect.NewResponse(response), nil
}

// CancelImportJob cancels an active import job
func (h *DataBridgeHandler) CancelImportJob(
	ctx context.Context,
	req *connect.Request[pb.CancelImportJobRequest],
) (*connect.Response[pb.CancelImportJobResponse], error) {
	h.logger.Info("CancelImportJob request received", "job_id", req.Msg.JobId)

	jobID, err := uuid.Parse(req.Msg.JobId)
	if err != nil {
		h.logger.Error("Invalid job ID", "job_id", req.Msg.JobId, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	err = h.services.Import.CancelImportJob(ctx, jobID)
	if err != nil {
		h.logger.Error("Failed to cancel import job", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to cancel import job: %w", err))
	}

	h.logger.Info("CancelImportJob request completed")

	return connect.NewResponse(&pb.CancelImportJobResponse{}), nil
}

// ExecuteImport executes an import job with streaming progress updates
func (h *DataBridgeHandler) ExecuteImport(
	ctx context.Context,
	req *connect.Request[pb.ExecuteImportRequest],
	stream *connect.ServerStream[pb.ExecuteImportResponse],
) error {
	h.logger.Info("ExecuteImport request received", "job_id", req.Msg.JobId)

	jobID, err := uuid.Parse(req.Msg.JobId)
	if err != nil {
		h.logger.Error("Invalid job ID", "job_id", req.Msg.JobId, "error", err)
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	// Create progress channel for streaming updates
	progressChan := make(chan *pb.ImportProgress, 100)
	defer close(progressChan)

	// Start import execution in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in import execution", "panic", r)
				// Send error through channel
				progressChan <- &pb.ImportProgress{
					// Handle panic scenario
				}
			}
		}()

		err := h.services.Import.ExecuteImport(ctx, jobID, req.Msg.CsvData, progressChan)
		if err != nil {
			h.logger.Error("Import execution failed", "error", err)
			// Send final error status
			progressChan <- &pb.ImportProgress{
				// Handle error scenario
			}
		}
	}()

	// Stream progress updates
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("ExecuteImport context cancelled")
			return ctx.Err()

		case progress, ok := <-progressChan:
			if !ok {
				h.logger.Info("ExecuteImport completed")
				return nil
			}

			if err := stream.Send(&pb.ExecuteImportResponse{
				Response: &pb.ExecuteImportResponse_Progress{
					Progress: progress,
				},
			}); err != nil {
				h.logger.Error("Failed to send progress update", "error", err)
				return err
			}
		}
	}
}