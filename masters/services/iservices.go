// =============================================================================
// services/iservices.go - Service interfaces for Masters module
// =============================================================================
package services

import (
	"context"

	pb "p9e.in/ugcl/masters/api/proto"

	"github.com/google/uuid"
)

// IMetadataService defines metadata management service interface
type IMetadataService interface {
	// Schema operations
	GetSchemas(ctx context.Context) ([]*pb.Schema, error)
	GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*pb.Schema, error)
	CreateSchema(ctx context.Context, req *pb.CreateSchemaRequest) (*pb.Schema, error)
	UpdateSchema(ctx context.Context, req *pb.UpdateSchemaRequest) (*pb.Schema, error)
	DeleteSchema(ctx context.Context, schemaID uuid.UUID, userID string) error

	// Table operations
	GetTables(ctx context.Context, schemaID *uuid.UUID, importableOnly bool) ([]*pb.Table, error)
	GetTableByID(ctx context.Context, tableID uuid.UUID) (*pb.Table, error)
	CreateTable(ctx context.Context, req *pb.CreateTableRequest) (*pb.Table, error)
	UpdateTable(ctx context.Context, req *pb.UpdateTableRequest) (*pb.Table, error)
	DeleteTable(ctx context.Context, tableID uuid.UUID, userID string) error

	// Column operations
	GetColumns(ctx context.Context, tableID uuid.UUID, importableOnly bool) ([]*pb.Column, error)
	GetColumnByID(ctx context.Context, columnID uuid.UUID) (*pb.Column, error)
	CreateColumn(ctx context.Context, req *pb.CreateColumnRequest) (*pb.Column, error)
	UpdateColumn(ctx context.Context, req *pb.UpdateColumnRequest) (*pb.Column, error)
	DeleteColumn(ctx context.Context, columnID uuid.UUID, userID string) error

	// Relationship operations
	CreateRelationship(ctx context.Context, req *pb.CreateRelationshipRequest) (*pb.TableRelationship, error)
	UpdateRelationship(ctx context.Context, req *pb.UpdateRelationshipRequest) (*pb.TableRelationship, error)
	DeleteRelationship(ctx context.Context, relationshipID uuid.UUID, userID string) error
	GetRelationshipsByTable(ctx context.Context, tableID uuid.UUID) ([]*pb.TableRelationship, error)
	GetAllRelationships(ctx context.Context) ([]*pb.TableRelationship, error)

	// Business term operations
	CreateBusinessTerm(ctx context.Context, req *pb.CreateBusinessTermRequest) (*pb.BusinessTerm, error)
	UpdateBusinessTerm(ctx context.Context, req *pb.UpdateBusinessTermRequest) (*pb.BusinessTerm, error)
	DeleteBusinessTerm(ctx context.Context, businessTermID uuid.UUID, userID string) error
	GetBusinessTerms(ctx context.Context) ([]*pb.BusinessTerm, error)
	GetBusinessTermsByDomain(ctx context.Context, domain string) ([]*pb.BusinessTerm, error)
	SearchBusinessTerms(ctx context.Context, query string) ([]*pb.BusinessTerm, error)
	LinkColumnToBusinessTerm(ctx context.Context, req *pb.LinkColumnToBusinessTermRequest) (*pb.ColumnBusinessTerm, error)
	UnlinkColumnFromBusinessTerm(ctx context.Context, columnID, businessTermID uuid.UUID) error
	GetBusinessTermsForColumn(ctx context.Context, columnID uuid.UUID) ([]*pb.BusinessTerm, error)
	GetColumnsForBusinessTerm(ctx context.Context, businessTermID uuid.UUID) ([]*pb.Column, error)
}

// ServiceManager combines all service interfaces for masters module
type ServiceManager struct {
	Metadata IMetadataService
}

// NewServiceManager creates a new service manager instance for masters module
func NewServiceManager(metadata IMetadataService) *ServiceManager {
	return &ServiceManager{
		Metadata: metadata,
	}
}