package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	pb "p9e.in/ugcl/insighthub/api/proto"
	"p9e.in/ugcl/insighthub/mappers"
	"p9e.in/ugcl/insighthub/models"
	"p9e.in/ugcl/insighthub/repository"
)

// ReportService implements the report service using clean architecture
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

	// Get existing fields for the report to determine the current field
	reportFields, err := s.repoManager.Report.GetReportFields(ctx, uuid.Nil) // We need the field ID only
	if err != nil {
		return nil, fmt.Errorf("failed to get report fields: %w", err)
	}

	var existingField *models.ReportField
	for _, field := range reportFields {
		if field.ID == fieldID {
			existingField = field
			break
		}
	}

	if existingField == nil {
		return nil, fmt.Errorf("report field not found")
	}

	var formattingRules map[string]interface{}
	if req.FormattingRules != nil {
		formattingRules = req.FormattingRules.AsMap()
	}

	// Update fields
	existingField.Alias = req.Alias
	existingField.AggregateFunction = req.AggregateFunction
	existingField.IsVisible = req.IsVisible
	existingField.FormattingRules = formattingRules

	updatedField, err := s.repoManager.Report.UpdateReportField(ctx, existingField)
	if err != nil {
		return nil, fmt.Errorf("failed to update report field: %w", err)
	}

	return s.protoMapper.DomainReportFieldToProto(updatedField), nil
}

func (s *ReportService) RemoveReportField(ctx context.Context, fieldID uuid.UUID) error {
	err := s.repoManager.Report.RemoveReportField(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("failed to remove report field: %w", err)
	}

	return nil
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

	filter := &models.ReportFilter{
		ID:              uuid.New(),
		ReportID:        reportID,
		FieldID:         fieldID,
		Operator:        req.Operator,
		Value:           req.Value,
		ValueType:       req.ValueType,
		LogicalOperator: req.LogicalOperator,
		GroupIndex:      req.GroupIndex,
		OrderIndex:      0, // Will be set by repository
		IsActive:        true,
		CreatedAt:       time.Now(),
	}

	createdFilter, err := s.repoManager.Report.AddReportFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to add report filter: %w", err)
	}

	return s.protoMapper.DomainReportFilterToProto(createdFilter), nil
}

func (s *ReportService) UpdateReportFilter(ctx context.Context, req *pb.UpdateReportFilterRequest) (*pb.ReportFilter, error) {
	filterID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid filter ID: %w", err)
	}

	// Get existing filters to find the one to update
	reportFilters, err := s.repoManager.Report.GetReportFilters(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get report filters: %w", err)
	}

	var existingFilter *models.ReportFilter
	for _, filter := range reportFilters {
		if filter.ID == filterID {
			existingFilter = filter
			break
		}
	}

	if existingFilter == nil {
		return nil, fmt.Errorf("report filter not found")
	}

	// Update fields
	existingFilter.Operator = req.Operator
	existingFilter.Value = req.Value
	existingFilter.ValueType = req.ValueType
	existingFilter.LogicalOperator = req.LogicalOperator
	existingFilter.GroupIndex = req.GroupIndex

	updatedFilter, err := s.repoManager.Report.UpdateReportFilter(ctx, existingFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to update report filter: %w", err)
	}

	return s.protoMapper.DomainReportFilterToProto(updatedFilter), nil
}

func (s *ReportService) RemoveReportFilter(ctx context.Context, filterID uuid.UUID) error {
	err := s.repoManager.Report.RemoveReportFilter(ctx, filterID)
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

	group := &models.ReportGroup{
		ID:              uuid.New(),
		ReportID:        reportID,
		FieldID:         fieldID,
		OrderIndex:      0, // Will be set by repository
		DateGranularity: req.DateGranularity,
		CreatedAt:       time.Now(),
	}

	createdGroup, err := s.repoManager.Report.AddReportGroup(ctx, group)
	if err != nil {
		return nil, fmt.Errorf("failed to add report group: %w", err)
	}

	return s.protoMapper.DomainReportGroupToProto(createdGroup), nil
}

func (s *ReportService) UpdateReportGroup(ctx context.Context, req *pb.UpdateReportGroupRequest) (*pb.ReportGroup, error) {
	groupID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	// Get existing groups to find the one to update
	reportGroups, err := s.repoManager.Report.GetReportGroups(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get report groups: %w", err)
	}

	var existingGroup *models.ReportGroup
	for _, group := range reportGroups {
		if group.ID == groupID {
			existingGroup = group
			break
		}
	}

	if existingGroup == nil {
		return nil, fmt.Errorf("report group not found")
	}

	// Update fields
	existingGroup.DateGranularity = req.DateGranularity

	updatedGroup, err := s.repoManager.Report.UpdateReportGroup(ctx, existingGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to update report group: %w", err)
	}

	return s.protoMapper.DomainReportGroupToProto(updatedGroup), nil
}

func (s *ReportService) RemoveReportGroup(ctx context.Context, groupID uuid.UUID) error {
	err := s.repoManager.Report.RemoveReportGroup(ctx, groupID)
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

	sort := &models.ReportSort{
		ID:         uuid.New(),
		ReportID:   reportID,
		FieldID:    fieldID,
		Direction:  req.Direction,
		OrderIndex: 0, // Will be set by repository
		CreatedAt:  time.Now(),
	}

	createdSort, err := s.repoManager.Report.AddReportSort(ctx, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to add report sort: %w", err)
	}

	return s.protoMapper.DomainReportSortToProto(createdSort), nil
}

func (s *ReportService) UpdateReportSort(ctx context.Context, req *pb.UpdateReportSortRequest) (*pb.ReportSort, error) {
	sortID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid sort ID: %w", err)
	}

	// Get existing sorts to find the one to update
	reportSorts, err := s.repoManager.Report.GetReportSorts(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get report sorts: %w", err)
	}

	var existingSort *models.ReportSort
	for _, sort := range reportSorts {
		if sort.ID == sortID {
			existingSort = sort
			break
		}
	}

	if existingSort == nil {
		return nil, fmt.Errorf("report sort not found")
	}

	// Update fields
	existingSort.Direction = req.Direction

	updatedSort, err := s.repoManager.Report.UpdateReportSort(ctx, existingSort)
	if err != nil {
		return nil, fmt.Errorf("failed to update report sort: %w", err)
	}

	return s.protoMapper.DomainReportSortToProto(updatedSort), nil
}

func (s *ReportService) RemoveReportSort(ctx context.Context, sortID uuid.UUID) error {
	err := s.repoManager.Report.RemoveReportSort(ctx, sortID)
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

	var xAxisFieldID, yAxisFieldID, seriesFieldID uuid.UUID

	if req.XAxisFieldId != "" {
		xAxisFieldID, err = uuid.Parse(req.XAxisFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid X axis field ID: %w", err)
		}
	}

	if req.YAxisFieldId != "" {
		yAxisFieldID, err = uuid.Parse(req.YAxisFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid Y axis field ID: %w", err)
		}
	}

	if req.SeriesFieldId != "" {
		seriesFieldID, err = uuid.Parse(req.SeriesFieldId)
		if err != nil {
			return nil, fmt.Errorf("invalid series field ID: %w", err)
		}
	}

	var chartOptions map[string]interface{}
	if req.ChartOptions != nil {
		chartOptions = req.ChartOptions.AsMap()
	}

	chart := &models.ReportChart{
		ID:            uuid.New(),
		ReportID:      reportID,
		ChartType:     req.ChartType,
		XAxisFieldID:  xAxisFieldID,
		YAxisFieldID:  yAxisFieldID,
		SeriesFieldID: seriesFieldID,
		ChartOptions:  chartOptions,
		CreatedAt:     time.Now(),
	}

	createdChart, err := s.repoManager.Report.SetReportChart(ctx, chart)
	if err != nil {
		return nil, fmt.Errorf("failed to set report chart: %w", err)
	}

	return s.protoMapper.DomainReportChartToProto(createdChart), nil
}

func (s *ReportService) RemoveReportChart(ctx context.Context, reportID uuid.UUID) error {
	err := s.repoManager.Report.RemoveReportChart(ctx, reportID)
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
		GeneratedSql:      "SELECT * FROM sample_table WHERE 1=1 LIMIT 10",
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

	permission := &models.ReportPermission{
		ID:              uuid.New(),
		ReportID:        reportID,
		PrincipalType:   req.PrincipalType,
		PrincipalID:     req.PrincipalId,
		PermissionLevel: req.PermissionLevel,
		IsActive:        true,
		GrantedBy:       req.GrantedBy,
		GrantedAt:       time.Now(),
	}

	createdPermission, err := s.repoManager.Report.GrantReportPermission(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to grant report permission: %w", err)
	}

	return &pb.ReportPermission{
		Id:              createdPermission.ID.String(),
		ReportId:        createdPermission.ReportID.String(),
		PrincipalType:   createdPermission.PrincipalType,
		PrincipalId:     createdPermission.PrincipalID,
		PermissionLevel: createdPermission.PermissionLevel,
		IsActive:        createdPermission.IsActive,
		GrantedBy:       createdPermission.GrantedBy,
		GrantedAt:       nil, // Will be converted properly by proto mapper
	}, nil
}

func (s *ReportService) RevokeReportPermission(ctx context.Context, permissionID uuid.UUID) error {
	err := s.repoManager.Report.RevokeReportPermission(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke report permission: %w", err)
	}

	return nil
}

func (s *ReportService) GetReportPermissions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportPermission, error) {
	permissions, err := s.repoManager.Report.GetReportPermissions(ctx, reportID)
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
			GrantedAt:       nil, // Will be converted properly by proto mapper
		})
	}

	return result, nil
}
