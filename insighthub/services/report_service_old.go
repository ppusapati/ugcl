package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/insighthub/api/proto"
	"p9e.in/ugcl/insighthub/mappers"
	"p9e.in/ugcl/insighthub/models"
	"p9e.in/ugcl/insighthub/repository"
)

// ReportService implements the IReportService interface
type ReportService struct {
	repoManager *repository.RepositoryManager
	protoMapper *mappers.ReportMapper
}

// NewReportService creates a new report service instance
func NewReportService(repoManager *repository.RepositoryManager) *ReportService {
	return &ReportService{
		repoManager: repoManager,
		protoMapper: mappers.NewReportMapper(),
	}
}

// =============================================================================
// Report CRUD Operations
// =============================================================================

func (s *ReportService) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.Report, error) {
	var metadata map[string]interface{}
	if req.Metadata != nil {
		metadata = req.Metadata.AsMap()
	}

	report := &models.Report{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		IsTemplate:  req.IsTemplate,
		IsActive:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
		Metadata:    metadata,
	}

	createdReport, err := s.repoManager.Report.CreateReport(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return s.protoMapper.DomainReportToProto(createdReport), nil
}

func (s *ReportService) GetReport(ctx context.Context, reportID uuid.UUID, userID string, userRoles []string) (*pb.Report, error) {
	report, err := s.repoManager.Report.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// TODO: Implement permission checking based on userID and userRoles
	// For now, return the report if it exists and is active

	return s.protoMapper.DomainReportToProto(report), nil
}

func (s *ReportService) UpdateReport(ctx context.Context, req *pb.UpdateReportRequest) (*pb.Report, error) {
	reportID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	// Get existing report
	existingReport, err := s.repoManager.Report.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing report: %w", err)
	}

	var metadata map[string]interface{}
	if req.Metadata != nil {
		metadata = req.Metadata.AsMap()
	}

	// Update fields
	existingReport.Name = req.Name
	existingReport.Description = req.Description
	existingReport.Category = req.Category
	existingReport.Tags = req.Tags
	existingReport.UpdatedBy = req.UpdatedBy
	existingReport.UpdatedAt = time.Now()
	existingReport.Metadata = metadata

	updatedReport, err := s.repoManager.Report.UpdateReport(ctx, existingReport)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return s.protoMapper.DomainReportToProto(updatedReport), nil
}

func (s *ReportService) DeleteReport(ctx context.Context, reportID uuid.UUID, deletedBy string) error {
	err := s.repoManager.Report.DeleteReport(ctx, reportID, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

func (s *ReportService) SearchReports(ctx context.Context, query string, userID string, userRoles []string, limit, offset int32) ([]*pb.Report, int32, error) {
	reports, totalCount, err := s.repoManager.Report.SearchReports(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search reports: %w", err)
	}

	// TODO: Implement permission filtering based on userID and userRoles

	var result []*pb.Report
	for _, report := range reports {
		result = append(result, s.protoMapper.DomainReportToProto(report))
	}

	return result, totalCount, nil
}

func (s *ReportService) CloneReport(ctx context.Context, req *pb.CloneReportRequest) (*pb.Report, error) {
	sourceReportID, err := uuid.Parse(req.SourceReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid source report ID: %w", err)
	}

	report, err := s.repoManager.Report.CloneReport(ctx, sourceReportID, req.NewName, req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to clone report: %w", err)
	}

	return s.protoMapper.DomainReportToProto(report), nil
}

// =============================================================================
// Report Field Management
// =============================================================================

func (s *ReportService) AddReportField(ctx context.Context, req *pb.AddReportFieldRequest) (*pb.ReportField, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	fieldID, err := uuid.Parse(req.FieldId)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	var formattingRules map[string]interface{}
	if req.FormattingRules != nil {
		formattingRules = req.FormattingRules.AsMap()
	}

	field := &models.ReportField{
		ID:                uuid.New(),
		ReportID:          reportID,
		FieldID:           fieldID,
		Alias:             req.Alias,
		AggregateFunction: req.AggregateFunction,
		OrderIndex:        0, // Will be set by repository
		IsVisible:         req.IsVisible,
		FormattingRules:   formattingRules,
		CreatedAt:         time.Now(),
	}

	createdField, err := s.repoManager.Report.AddReportField(ctx, field)
	if err != nil {
		return nil, fmt.Errorf("failed to add report field: %w", err)
	}

	return s.protoMapper.DomainReportFieldToProto(createdField), nil
}

func (s *ReportService) UpdateReportField(ctx context.Context, req *pb.UpdateReportFieldRequest) (*pb.ReportField, error) {
	fieldID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	var formattingRules []byte
	if req.FormattingRules != nil {
		rulesBytes, err := json.Marshal(req.FormattingRules.AsMap())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal formatting rules: %w", err)
		}
		formattingRules = rulesBytes
	}

	field, err := s.queries.UpdateReportField(ctx, generated.UpdateReportFieldParams{
		ID:                fieldID,
		Alias:             sql.NullString{String: req.Alias, Valid: req.Alias != ""},
		AggregateFunction: sql.NullString{String: req.AggregateFunction, Valid: req.AggregateFunction != ""},
		IsVisible:         req.IsVisible,
		FormattingRules:   formattingRules,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report field: %w", err)
	}

	return s.mapper.ToProtoReportField(field), nil
}

func (s *ReportService) RemoveReportField(ctx context.Context, fieldID uuid.UUID) error {
	err := s.queries.RemoveReportField(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("failed to remove report field: %w", err)
	}

	return nil
}

func (s *ReportService) ReorderReportFields(ctx context.Context, req *pb.ReorderReportFieldsRequest) error {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}

	// TODO: Implement field reordering
	// This would require updating the order_index for multiple fields
	// For now, return a placeholder implementation

	_ = reportID // Use the variable to avoid compiler warning
	return fmt.Errorf("field reordering not yet implemented")
}

// =============================================================================
// Report Filter Management
// =============================================================================

func (s *ReportService) AddReportFilter(ctx context.Context, req *pb.AddReportFilterRequest) (*pb.ReportFilter, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	fieldID, err := uuid.Parse(req.FieldId)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	filter, err := s.queries.AddReportFilter(ctx, generated.AddReportFilterParams{
		ReportID:         reportID,
		FieldID:          fieldID,
		Operator:         req.Operator,
		Value:            req.Value,
		ValueType:        req.ValueType,
		LogicalOperator:  sql.NullString{String: req.LogicalOperator, Valid: req.LogicalOperator != ""},
		GroupIndex:       sql.NullInt32{Int32: req.GroupIndex, Valid: req.GroupIndex > 0},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report filter: %w", err)
	}

	return s.mapper.ToProtoReportFilter(filter), nil
}

func (s *ReportService) UpdateReportFilter(ctx context.Context, req *pb.UpdateReportFilterRequest) (*pb.ReportFilter, error) {
	filterID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid filter ID: %w", err)
	}

	filter, err := s.queries.UpdateReportFilter(ctx, generated.UpdateReportFilterParams{
		ID:               filterID,
		Operator:         req.Operator,
		Value:            req.Value,
		ValueType:        req.ValueType,
		LogicalOperator:  sql.NullString{String: req.LogicalOperator, Valid: req.LogicalOperator != ""},
		GroupIndex:       sql.NullInt32{Int32: req.GroupIndex, Valid: req.GroupIndex > 0},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report filter: %w", err)
	}

	return s.mapper.ToProtoReportFilter(filter), nil
}

func (s *ReportService) RemoveReportFilter(ctx context.Context, filterID uuid.UUID) error {
	err := s.queries.RemoveReportFilter(ctx, filterID)
	if err != nil {
		return fmt.Errorf("failed to remove report filter: %w", err)
	}

	return nil
}

// =============================================================================
// Report Group Management
// =============================================================================

func (s *ReportService) AddReportGroup(ctx context.Context, req *pb.AddReportGroupRequest) (*pb.ReportGroup, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	fieldID, err := uuid.Parse(req.FieldId)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	group, err := s.queries.AddReportGroup(ctx, generated.AddReportGroupParams{
		ReportID:         reportID,
		FieldID:          fieldID,
		DateGranularity:  sql.NullString{String: req.DateGranularity, Valid: req.DateGranularity != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report group: %w", err)
	}

	return s.mapper.ToProtoReportGroup(group), nil
}

func (s *ReportService) UpdateReportGroup(ctx context.Context, req *pb.UpdateReportGroupRequest) (*pb.ReportGroup, error) {
	groupID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	group, err := s.queries.UpdateReportGroup(ctx, generated.UpdateReportGroupParams{
		ID:               groupID,
		DateGranularity:  sql.NullString{String: req.DateGranularity, Valid: req.DateGranularity != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report group: %w", err)
	}

	return s.mapper.ToProtoReportGroup(group), nil
}

func (s *ReportService) RemoveReportGroup(ctx context.Context, groupID uuid.UUID) error {
	err := s.queries.RemoveReportGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to remove report group: %w", err)
	}

	return nil
}

// =============================================================================
// Report Sort Management
// =============================================================================

func (s *ReportService) AddReportSort(ctx context.Context, req *pb.AddReportSortRequest) (*pb.ReportSort, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	fieldID, err := uuid.Parse(req.FieldId)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	sort, err := s.queries.AddReportSort(ctx, generated.AddReportSortParams{
		ReportID:  reportID,
		FieldID:   fieldID,
		Direction: req.Direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report sort: %w", err)
	}

	return s.mapper.ToProtoReportSort(sort), nil
}

func (s *ReportService) UpdateReportSort(ctx context.Context, req *pb.UpdateReportSortRequest) (*pb.ReportSort, error) {
	sortID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid sort ID: %w", err)
	}

	sort, err := s.queries.UpdateReportSort(ctx, generated.UpdateReportSortParams{
		ID:        sortID,
		Direction: req.Direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report sort: %w", err)
	}

	return s.mapper.ToProtoReportSort(sort), nil
}

func (s *ReportService) RemoveReportSort(ctx context.Context, sortID uuid.UUID) error {
	err := s.queries.RemoveReportSort(ctx, sortID)
	if err != nil {
		return fmt.Errorf("failed to remove report sort: %w", err)
	}

	return nil
}

// =============================================================================
// Report Chart Management
// =============================================================================

func (s *ReportService) SetReportChart(ctx context.Context, req *pb.SetReportChartRequest) (*pb.ReportChart, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	var xAxisFieldID, yAxisFieldID, seriesFieldID uuid.NullUUID

	if req.XAxisFieldId != "" {
		id, err := uuid.Parse(req.XAxisFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid X axis field ID: %w", err)
		}
		xAxisFieldID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if req.YAxisFieldId != "" {
		id, err := uuid.Parse(req.YAxisFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid Y axis field ID: %w", err)
		}
		yAxisFieldID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if req.SeriesFieldId != "" {
		id, err := uuid.Parse(req.SeriesFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid series field ID: %w", err)
		}
		seriesFieldID = uuid.NullUUID{UUID: id, Valid: true}
	}

	var chartOptions []byte
	if req.ChartOptions != nil {
		optionsBytes, err := json.Marshal(req.ChartOptions.AsMap())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chart options: %w", err)
		}
		chartOptions = optionsBytes
	}

	chart, err := s.queries.SetReportChart(ctx, generated.SetReportChartParams{
		ReportID:       reportID,
		ChartType:      req.ChartType,
		XAxisFieldID:   xAxisFieldID,
		YAxisFieldID:   yAxisFieldID,
		SeriesFieldID:  seriesFieldID,
		ChartOptions:   chartOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set report chart: %w", err)
	}

	return s.mapper.ToProtoReportChart(chart), nil
}

func (s *ReportService) RemoveReportChart(ctx context.Context, reportID uuid.UUID) error {
	err := s.queries.RemoveReportChart(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to remove report chart: %w", err)
	}

	return nil
}

// =============================================================================
// Report Validation and Preview
// =============================================================================

func (s *ReportService) ValidateReport(ctx context.Context, reportID uuid.UUID) (*pb.ValidateReportResponse, error) {
	// TODO: Implement report validation logic
	// This would check for:
	// - Required fields are present
	// - Valid field references
	// - Valid filter operators
	// - Valid chart configuration

	return &pb.ValidateReportResponse{
		IsValid:  true,
		Errors:   []*pb.ValidationError{},
		Warnings: []*pb.ValidationWarning{},
	}, nil
}

func (s *ReportService) PreviewReportQuery(ctx context.Context, req *pb.PreviewReportQueryRequest) (*pb.PreviewReportQueryResponse, error) {
	// TODO: Implement query preview generation
	// This would:
	// - Build SQL query from report definition
	// - Execute with LIMIT for preview
	// - Return sample data and metadata

	return &pb.PreviewReportQueryResponse{
		GeneratedSql:       "SELECT * FROM sample_table WHERE 1=1 LIMIT 10",
		Columns:           []*pb.ColumnInfo{},
		SampleRows:        []*structpb.Struct{},
		EstimatedRowCount: 0,
	}, nil
}

// =============================================================================
// Report Permissions
// =============================================================================

func (s *ReportService) GrantReportPermission(ctx context.Context, req *pb.GrantReportPermissionRequest) (*pb.ReportPermission, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	permission, err := s.queries.GrantReportPermission(ctx, generated.GrantReportPermissionParams{
		ReportID:        reportID,
		PrincipalType:   req.PrincipalType,
		PrincipalID:     req.PrincipalId,
		PermissionLevel: req.PermissionLevel,
		GrantedBy:       req.GrantedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to grant report permission: %w", err)
	}

	return &pb.ReportPermission{
		Id:              permission.ID.String(),
		ReportId:        permission.ReportID.String(),
		PrincipalType:   permission.PrincipalType,
		PrincipalId:     permission.PrincipalID,
		PermissionLevel: permission.PermissionLevel,
		IsActive:        permission.IsActive,
		GrantedBy:       permission.GrantedBy,
		GrantedAt:       timestamppb.New(permission.GrantedAt),
	}, nil
}

func (s *ReportService) RevokeReportPermission(ctx context.Context, permissionID uuid.UUID) error {
	err := s.queries.RevokeReportPermission(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke report permission: %w", err)
	}

	return nil
}

func (s *ReportService) GetReportPermissions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportPermission, error) {
	permissions, err := s.queries.GetReportPermissions(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report permissions: %w", err)
	}

	var result []*pb.ReportPermission
	for _, permission := range permissions {
		result = append(result, &pb.ReportPermission{
			Id:              permission.ID.String(),
			ReportId:        permission.ReportID.String(),
			PrincipalType:   permission.PrincipalType,
			PrincipalId:     permission.PrincipalID,
			PermissionLevel: permission.PermissionLevel,
			IsActive:        permission.IsActive,
			GrantedBy:       permission.GrantedBy,
			GrantedAt:       timestamppb.New(permission.GrantedAt),
		})
	}

	return result, nil
}