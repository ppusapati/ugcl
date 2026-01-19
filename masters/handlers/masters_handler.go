package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "p9e.in/ugcl/masters/api/proto"
	"p9e.in/ugcl/masters/api/proto/mastersv1connect"
	"p9e.in/ugcl/masters/services"
)

// MastersHandler implements the Masters Connect RPC service
type MastersHandler struct {
	serviceManager *services.ServiceManager
}

// NewMastersHandler creates a new masters handler instance
func NewMastersHandler(serviceManager *services.ServiceManager) *MastersHandler {
	return &MastersHandler{
		serviceManager: serviceManager,
	}
}

// GetPath returns the Connect RPC path for this handler
func (h *MastersHandler) GetPath() (string, any) {
	return mastersv1connect.NewMastersServiceHandler(h)
}

// =============================================================================
// Schema Operations
// =============================================================================

func (h *MastersHandler) GetSchemas(ctx context.Context, req *connect.Request[pb.GetSchemasRequest]) (*connect.Response[pb.GetSchemasResponse], error) {
	schemas, err := h.serviceManager.Metadata.GetSchemas(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get schemas: %w", err))
	}

	resp := &pb.GetSchemasResponse{
		Schemas: schemas,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetSchemaByID(ctx context.Context, req *connect.Request[pb.GetSchemaByIDRequest]) (*connect.Response[pb.Schema], error) {
	schemaID, err := uuid.Parse(req.Msg.SchemaId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid schema ID: %w", err))
	}

	schema, err := h.serviceManager.Metadata.GetSchemaByID(ctx, schemaID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("schema not found: %w", err))
	}

	return connect.NewResponse(schema), nil
}

func (h *MastersHandler) CreateSchema(ctx context.Context, req *connect.Request[pb.CreateSchemaRequest]) (*connect.Response[pb.Schema], error) {
	schema, err := h.serviceManager.Metadata.CreateSchema(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create schema: %w", err))
	}

	return connect.NewResponse(schema), nil
}

func (h *MastersHandler) UpdateSchema(ctx context.Context, req *connect.Request[pb.UpdateSchemaRequest]) (*connect.Response[pb.Schema], error) {
	schema, err := h.serviceManager.Metadata.UpdateSchema(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update schema: %w", err))
	}

	return connect.NewResponse(schema), nil
}

func (h *MastersHandler) DeleteSchema(ctx context.Context, req *connect.Request[pb.DeleteSchemaRequest]) (*connect.Response[emptypb.Empty], error) {
	schemaID, err := uuid.Parse(req.Msg.SchemaId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid schema ID: %w", err))
	}

	err = h.serviceManager.Metadata.DeleteSchema(ctx, schemaID, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete schema: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// =============================================================================
// Table Operations
// =============================================================================

func (h *MastersHandler) GetTables(ctx context.Context, req *connect.Request[pb.GetTablesRequest]) (*connect.Response[pb.GetTablesResponse], error) {
	var schemaID *uuid.UUID
	if req.Msg.SchemaId != "" {
		id, err := uuid.Parse(req.Msg.SchemaId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid schema ID: %w", err))
		}
		schemaID = &id
	}

	tables, err := h.serviceManager.Metadata.GetTables(ctx, schemaID, req.Msg.ImportableOnly)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get tables: %w", err))
	}

	resp := &pb.GetTablesResponse{
		Tables: tables,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetTableByID(ctx context.Context, req *connect.Request[pb.GetTableByIDRequest]) (*connect.Response[pb.Table], error) {
	tableID, err := uuid.Parse(req.Msg.TableId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid table ID: %w", err))
	}

	table, err := h.serviceManager.Metadata.GetTableByID(ctx, tableID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("table not found: %w", err))
	}

	return connect.NewResponse(table), nil
}

func (h *MastersHandler) CreateTable(ctx context.Context, req *connect.Request[pb.CreateTableRequest]) (*connect.Response[pb.Table], error) {
	table, err := h.serviceManager.Metadata.CreateTable(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create table: %w", err))
	}

	return connect.NewResponse(table), nil
}

func (h *MastersHandler) UpdateTable(ctx context.Context, req *connect.Request[pb.UpdateTableRequest]) (*connect.Response[pb.Table], error) {
	table, err := h.serviceManager.Metadata.UpdateTable(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update table: %w", err))
	}

	return connect.NewResponse(table), nil
}

func (h *MastersHandler) DeleteTable(ctx context.Context, req *connect.Request[pb.DeleteTableRequest]) (*connect.Response[emptypb.Empty], error) {
	tableID, err := uuid.Parse(req.Msg.TableId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid table ID: %w", err))
	}

	err = h.serviceManager.Metadata.DeleteTable(ctx, tableID, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete table: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// =============================================================================
// Column Operations
// =============================================================================

func (h *MastersHandler) GetColumns(ctx context.Context, req *connect.Request[pb.GetColumnsRequest]) (*connect.Response[pb.GetColumnsResponse], error) {
	tableID, err := uuid.Parse(req.Msg.TableId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid table ID: %w", err))
	}

	columns, err := h.serviceManager.Metadata.GetColumns(ctx, tableID, req.Msg.ImportableOnly)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get columns: %w", err))
	}

	resp := &pb.GetColumnsResponse{
		Columns: columns,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetColumnByID(ctx context.Context, req *connect.Request[pb.GetColumnByIDRequest]) (*connect.Response[pb.Column], error) {
	columnID, err := uuid.Parse(req.Msg.ColumnId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid column ID: %w", err))
	}

	column, err := h.serviceManager.Metadata.GetColumnByID(ctx, columnID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("column not found: %w", err))
	}

	return connect.NewResponse(column), nil
}

func (h *MastersHandler) CreateColumn(ctx context.Context, req *connect.Request[pb.CreateColumnRequest]) (*connect.Response[pb.Column], error) {
	column, err := h.serviceManager.Metadata.CreateColumn(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create column: %w", err))
	}

	return connect.NewResponse(column), nil
}

func (h *MastersHandler) UpdateColumn(ctx context.Context, req *connect.Request[pb.UpdateColumnRequest]) (*connect.Response[pb.Column], error) {
	column, err := h.serviceManager.Metadata.UpdateColumn(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update column: %w", err))
	}

	return connect.NewResponse(column), nil
}

func (h *MastersHandler) DeleteColumn(ctx context.Context, req *connect.Request[pb.DeleteColumnRequest]) (*connect.Response[emptypb.Empty], error) {
	columnID, err := uuid.Parse(req.Msg.ColumnId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid column ID: %w", err))
	}

	err = h.serviceManager.Metadata.DeleteColumn(ctx, columnID, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete column: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// =============================================================================
// Relationship Operations
// =============================================================================

func (h *MastersHandler) CreateRelationship(ctx context.Context, req *connect.Request[pb.CreateRelationshipRequest]) (*connect.Response[pb.TableRelationship], error) {
	relationship, err := h.serviceManager.Metadata.CreateRelationship(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create relationship: %w", err))
	}

	return connect.NewResponse(relationship), nil
}

func (h *MastersHandler) UpdateRelationship(ctx context.Context, req *connect.Request[pb.UpdateRelationshipRequest]) (*connect.Response[pb.TableRelationship], error) {
	relationship, err := h.serviceManager.Metadata.UpdateRelationship(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update relationship: %w", err))
	}

	return connect.NewResponse(relationship), nil
}

func (h *MastersHandler) DeleteRelationship(ctx context.Context, req *connect.Request[pb.DeleteRelationshipRequest]) (*connect.Response[emptypb.Empty], error) {
	relationshipID, err := uuid.Parse(req.Msg.RelationshipId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid relationship ID: %w", err))
	}

	err = h.serviceManager.Metadata.DeleteRelationship(ctx, relationshipID, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete relationship: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *MastersHandler) GetRelationshipsByTable(ctx context.Context, req *connect.Request[pb.GetRelationshipsByTableRequest]) (*connect.Response[pb.GetRelationshipsByTableResponse], error) {
	tableID, err := uuid.Parse(req.Msg.TableId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid table ID: %w", err))
	}

	relationships, err := h.serviceManager.Metadata.GetRelationshipsByTable(ctx, tableID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get relationships by table: %w", err))
	}

	resp := &pb.GetRelationshipsByTableResponse{
		Relationships: relationships,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetAllRelationships(ctx context.Context, req *connect.Request[pb.GetAllRelationshipsRequest]) (*connect.Response[pb.GetAllRelationshipsResponse], error) {
	relationships, err := h.serviceManager.Metadata.GetAllRelationships(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get all relationships: %w", err))
	}

	resp := &pb.GetAllRelationshipsResponse{
		Relationships: relationships,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Business Term Operations
// =============================================================================

func (h *MastersHandler) CreateBusinessTerm(ctx context.Context, req *connect.Request[pb.CreateBusinessTermRequest]) (*connect.Response[pb.BusinessTerm], error) {
	businessTerm, err := h.serviceManager.Metadata.CreateBusinessTerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create business term: %w", err))
	}

	return connect.NewResponse(businessTerm), nil
}

func (h *MastersHandler) UpdateBusinessTerm(ctx context.Context, req *connect.Request[pb.UpdateBusinessTermRequest]) (*connect.Response[pb.BusinessTerm], error) {
	businessTerm, err := h.serviceManager.Metadata.UpdateBusinessTerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update business term: %w", err))
	}

	return connect.NewResponse(businessTerm), nil
}

func (h *MastersHandler) DeleteBusinessTerm(ctx context.Context, req *connect.Request[pb.DeleteBusinessTermRequest]) (*connect.Response[emptypb.Empty], error) {
	businessTermID, err := uuid.Parse(req.Msg.BusinessTermId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid business term ID: %w", err))
	}

	err = h.serviceManager.Metadata.DeleteBusinessTerm(ctx, businessTermID, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete business term: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *MastersHandler) GetBusinessTerms(ctx context.Context, req *connect.Request[pb.GetBusinessTermsRequest]) (*connect.Response[pb.GetBusinessTermsResponse], error) {
	businessTerms, err := h.serviceManager.Metadata.GetBusinessTerms(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get business terms: %w", err))
	}

	resp := &pb.GetBusinessTermsResponse{
		BusinessTerms: businessTerms,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetBusinessTermsByDomain(ctx context.Context, req *connect.Request[pb.GetBusinessTermsByDomainRequest]) (*connect.Response[pb.GetBusinessTermsByDomainResponse], error) {
	businessTerms, err := h.serviceManager.Metadata.GetBusinessTermsByDomain(ctx, req.Msg.Domain)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get business terms by domain: %w", err))
	}

	resp := &pb.GetBusinessTermsByDomainResponse{
		BusinessTerms: businessTerms,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) SearchBusinessTerms(ctx context.Context, req *connect.Request[pb.SearchBusinessTermsRequest]) (*connect.Response[pb.SearchBusinessTermsResponse], error) {
	businessTerms, err := h.serviceManager.Metadata.SearchBusinessTerms(ctx, req.Msg.Query)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to search business terms: %w", err))
	}

	resp := &pb.SearchBusinessTermsResponse{
		BusinessTerms: businessTerms,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) LinkColumnToBusinessTerm(ctx context.Context, req *connect.Request[pb.LinkColumnToBusinessTermRequest]) (*connect.Response[pb.ColumnBusinessTerm], error) {
	link, err := h.serviceManager.Metadata.LinkColumnToBusinessTerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to link column to business term: %w", err))
	}

	return connect.NewResponse(link), nil
}

func (h *MastersHandler) UnlinkColumnFromBusinessTerm(ctx context.Context, req *connect.Request[pb.UnlinkColumnFromBusinessTermRequest]) (*connect.Response[emptypb.Empty], error) {
	columnID, err := uuid.Parse(req.Msg.ColumnId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid column ID: %w", err))
	}

	businessTermID, err := uuid.Parse(req.Msg.BusinessTermId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid business term ID: %w", err))
	}

	err = h.serviceManager.Metadata.UnlinkColumnFromBusinessTerm(ctx, columnID, businessTermID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unlink column from business term: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *MastersHandler) GetBusinessTermsForColumn(ctx context.Context, req *connect.Request[pb.GetBusinessTermsForColumnRequest]) (*connect.Response[pb.GetBusinessTermsForColumnResponse], error) {
	columnID, err := uuid.Parse(req.Msg.ColumnId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid column ID: %w", err))
	}

	businessTerms, err := h.serviceManager.Metadata.GetBusinessTermsForColumn(ctx, columnID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get business terms for column: %w", err))
	}

	resp := &pb.GetBusinessTermsForColumnResponse{
		BusinessTerms: businessTerms,
	}

	return connect.NewResponse(resp), nil
}

func (h *MastersHandler) GetColumnsForBusinessTerm(ctx context.Context, req *connect.Request[pb.GetColumnsForBusinessTermRequest]) (*connect.Response[pb.GetColumnsForBusinessTermResponse], error) {
	businessTermID, err := uuid.Parse(req.Msg.BusinessTermId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid business term ID: %w", err))
	}

	columns, err := h.serviceManager.Metadata.GetColumnsForBusinessTerm(ctx, businessTermID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get columns for business term: %w", err))
	}

	resp := &pb.GetColumnsForBusinessTermResponse{
		Columns: columns,
	}

	return connect.NewResponse(resp), nil
}