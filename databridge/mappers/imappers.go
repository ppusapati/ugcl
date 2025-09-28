// =============================================================================
// mappers/imappers.go - Mapper interfaces for DataBridge module
// =============================================================================
package mappers

import (
	pb "p9e.in/ugcl/databridge/api/databridge"
	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
)

// ISchemaMapper defines schema mapping interface
type ISchemaMapper interface {
	DBToProto(schema *db.SchemasMetadatum) *pb.Schema
	ProtoToDB(schema *pb.Schema) *db.SchemasMetadatum
	CreateRequestToRepoParams(req *pb.CreateSchemaRequest) *repository.CreateSchemaParams
	UpdateRequestToRepoParams(req *pb.UpdateSchemaRequest, schemaID uuid.UUID) *repository.UpdateSchemaParams
}

// ITableMapper defines table mapping interface
type ITableMapper interface {
	DBToProto(table *db.GetTablesBySchemaRow) *pb.Table
	SingleDBToProto(table *db.TablesMetadatum) *pb.Table
	ImportableDBToProto(table *db.GetAllImportableTablesRow) *pb.Table
	ProtoToDB(table *pb.Table) *db.TablesMetadatum
	CreateRequestToRepoParams(req *pb.CreateTableRequest, schemaID uuid.UUID) *repository.CreateTableParams
	UpdateRequestToRepoParams(req *pb.UpdateTableRequest, tableID uuid.UUID) *repository.UpdateTableParams
}

// IColumnMapper defines column mapping interface
type IColumnMapper interface {
	DBToProto(column *db.ColumnsMetadatum) *pb.Column
	ProtoToDB(column *pb.Column) *db.ColumnsMetadatum
	CreateRequestToRepoParams(req *pb.CreateColumnRequest, tableID uuid.UUID) *repository.CreateColumnParams
	UpdateRequestToRepoParams(req *pb.UpdateColumnRequest, columnID uuid.UUID) *repository.UpdateColumnParams
	DataTypeDBToProto(dataType string) pb.DataType
	DataTypeProtoToDB(dataType pb.DataType) string
}

// IMappingMapper defines import mapping interface
type IMappingMapper interface {
	DBToProto(mapping *db.ImportMapping) *pb.ImportMapping
	DetailDBToProto(mapping *db.GetMappingByIDRow) *pb.ImportMapping
	ProtoToDB(mapping *pb.ImportMapping) *db.ImportMapping
	CreateRequestToRepoParams(req *pb.CreateMappingRequest) *repository.CreateMappingParams
	UpdateRequestToRepoParams(req *pb.UpdateMappingRequest, mappingID uuid.UUID) *repository.UpdateMappingParams
}

// IJobMapper defines import job mapping interface
type IJobMapper interface {
	DBToProto(job *db.ImportJob) *pb.ImportJob
	DetailDBToProto(job *db.GetJobByIDRow) *pb.ImportJob
	UserJobDBToProto(job *db.GetJobsByUserRow) *pb.ImportJob
	ProtoToDB(job *pb.ImportJob) *db.ImportJob
	CreateRequestToRepoParams(req *pb.CreateImportJobRequest) *repository.CreateJobParams
	StatusDBToProto(status string) pb.ImportJobStatus
	StatusProtoToDB(status pb.ImportJobStatus) string
}

// MapperManager combines all mapper interfaces
type MapperManager struct {
	Schema  ISchemaMapper
	Table   ITableMapper
	Column  IColumnMapper
	Mapping IMappingMapper
	Job     IJobMapper
}