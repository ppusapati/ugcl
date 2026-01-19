package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"p9e.in/ugcl/insighthub/db/generated"
	"p9e.in/ugcl/insighthub/mappers"
	"p9e.in/ugcl/insighthub/models"
)

// ReportRepository implements the IReportRepository interface using SQLC
type ReportRepository struct {
	queries *generated.Queries
	mapper  *mappers.DomainMapper
}

// NewReportRepository creates a new report repository instance
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{
		queries: generated.New(db),
		mapper:  mappers.NewDomainMapper(),
	}
}

// =============================================================================
// Report CRUD Operations
// =============================================================================

func (r *ReportRepository) CreateReport(ctx context.Context, report *models.Report) (*models.Report, error) {
	sqlcReport := r.mapper.DomainReportToSQLC(report)

	createdReport, err := r.queries.CreateReport(ctx, generated.CreateReportParams{
		Name:        sqlcReport.Name,
		Description: sqlcReport.Description,
		Category:    sqlcReport.Category,
		Tags:        sqlcReport.Tags,
		IsTemplate:  sqlcReport.IsTemplate,
		CreatedBy:   sqlcReport.CreatedBy,
		Metadata:    sqlcReport.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return r.mapper.SQLCReportToDomain(createdReport), nil
}

func (r *ReportRepository) GetReportByID(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	sqlcReport, err := r.queries.GetReportByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	domainReport := r.mapper.SQLCReportToDomain(sqlcReport)

	// Load associated components
	if err := r.loadReportComponents(ctx, domainReport); err != nil {
		return nil, fmt.Errorf("failed to load report components: %w", err)
	}

	return domainReport, nil
}

func (r *ReportRepository) UpdateReport(ctx context.Context, report *models.Report) (*models.Report, error) {
	sqlcReport := r.mapper.DomainReportToSQLC(report)

	updatedReport, err := r.queries.UpdateReport(ctx, generated.UpdateReportParams{
		ID:          sqlcReport.ID,
		Name:        sqlcReport.Name,
		Description: sqlcReport.Description,
		Category:    sqlcReport.Category,
		Tags:        sqlcReport.Tags,
		UpdatedBy:   sqlcReport.UpdatedBy,
		Metadata:    sqlcReport.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return r.mapper.SQLCReportToDomain(updatedReport), nil
}

func (r *ReportRepository) DeleteReport(ctx context.Context, id uuid.UUID, deletedBy string) error {
	err := r.queries.DeleteReport(ctx, generated.DeleteReportParams{
		ID:        id,
		UpdatedBy: sql.NullString{String: deletedBy, Valid: deletedBy != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

func (r *ReportRepository) SearchReports(ctx context.Context, query string, limit, offset int32) ([]*models.Report, int32, error) {
	sqlcReports, err := r.queries.SearchReports(ctx, generated.SearchReportsParams{
		SearchQuery: query,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search reports: %w", err)
	}

	var reports []*models.Report
	for _, sqlcReport := range sqlcReports {
		reports = append(reports, r.mapper.SQLCReportToDomain(sqlcReport))
	}

	// Get total count
	totalCount, err := r.queries.CountSearchReports(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return reports, int32(totalCount), nil
}

func (r *ReportRepository) CloneReport(ctx context.Context, sourceID uuid.UUID, newName, createdBy string) (*models.Report, error) {
	clonedReport, err := r.queries.CloneReport(ctx, generated.CloneReportParams{
		ID:      sourceID,
		Name:    newName,
		Column3: createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone report: %w", err)
	}

	return r.mapper.SQLCReportToDomain(clonedReport), nil
}

// =============================================================================
// Report Field Operations
// =============================================================================

func (r *ReportRepository) AddReportField(ctx context.Context, field *models.ReportField) (*models.ReportField, error) {
	sqlcField := r.mapper.DomainReportFieldToSQLC(field)

	createdField, err := r.queries.AddReportField(ctx, generated.AddReportFieldParams{
		ReportID:          sqlcField.ReportID,
		FieldID:           sqlcField.FieldID,
		Alias:             sqlcField.Alias,
		AggregateFunction: sqlcField.AggregateFunction,
		IsVisible:         sqlcField.IsVisible,
		FormattingRules:   sqlcField.FormattingRules,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report field: %w", err)
	}

	return r.mapper.SQLCReportFieldToDomain(createdField), nil
}

func (r *ReportRepository) UpdateReportField(ctx context.Context, field *models.ReportField) (*models.ReportField, error) {
	sqlcField := r.mapper.DomainReportFieldToSQLC(field)

	updatedField, err := r.queries.UpdateReportField(ctx, generated.UpdateReportFieldParams{
		ID:                sqlcField.ID,
		Alias:             sqlcField.Alias,
		AggregateFunction: sqlcField.AggregateFunction,
		IsVisible:         sqlcField.IsVisible,
		FormattingRules:   sqlcField.FormattingRules,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report field: %w", err)
	}

	return r.mapper.SQLCReportFieldToDomain(updatedField), nil
}

func (r *ReportRepository) RemoveReportField(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RemoveReportField(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove report field: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportFields(ctx context.Context, reportID uuid.UUID) ([]*models.ReportField, error) {
	sqlcFields, err := r.queries.GetReportFields(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report fields: %w", err)
	}

	var fields []*models.ReportField
	for _, sqlcField := range sqlcFields {
		fields = append(fields, r.mapper.SQLCReportFieldToDomain(sqlcField))
	}

	return fields, nil
}

// =============================================================================
// Report Filter Operations
// =============================================================================

func (r *ReportRepository) AddReportFilter(ctx context.Context, filter *models.ReportFilter) (*models.ReportFilter, error) {
	sqlcFilter := r.mapper.DomainReportFilterToSQLC(filter)

	createdFilter, err := r.queries.AddReportFilter(ctx, generated.AddReportFilterParams{
		ReportID:         sqlcFilter.ReportID,
		FieldID:          sqlcFilter.FieldID,
		Operator:         sqlcFilter.Operator,
		Value:            sqlcFilter.Value,
		ValueType:        sqlcFilter.ValueType,
		LogicalOperator:  sqlcFilter.LogicalOperator,
		GroupIndex:       sqlcFilter.GroupIndex,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report filter: %w", err)
	}

	return r.mapper.SQLCReportFilterToDomain(createdFilter), nil
}

func (r *ReportRepository) UpdateReportFilter(ctx context.Context, filter *models.ReportFilter) (*models.ReportFilter, error) {
	sqlcFilter := r.mapper.DomainReportFilterToSQLC(filter)

	updatedFilter, err := r.queries.UpdateReportFilter(ctx, generated.UpdateReportFilterParams{
		ID:               sqlcFilter.ID,
		Operator:         sqlcFilter.Operator,
		Value:            sqlcFilter.Value,
		ValueType:        sqlcFilter.ValueType,
		LogicalOperator:  sqlcFilter.LogicalOperator,
		GroupIndex:       sqlcFilter.GroupIndex,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report filter: %w", err)
	}

	return r.mapper.SQLCReportFilterToDomain(updatedFilter), nil
}

func (r *ReportRepository) RemoveReportFilter(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RemoveReportFilter(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove report filter: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportFilters(ctx context.Context, reportID uuid.UUID) ([]*models.ReportFilter, error) {
	sqlcFilters, err := r.queries.GetReportFilters(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report filters: %w", err)
	}

	var filters []*models.ReportFilter
	for _, sqlcFilter := range sqlcFilters {
		filters = append(filters, r.mapper.SQLCReportFilterToDomain(sqlcFilter))
	}

	return filters, nil
}

// =============================================================================
// Report Group Operations
// =============================================================================

func (r *ReportRepository) AddReportGroup(ctx context.Context, group *models.ReportGroup) (*models.ReportGroup, error) {
	sqlcGroup := r.mapper.DomainReportGroupToSQLC(group)

	createdGroup, err := r.queries.AddReportGroup(ctx, generated.AddReportGroupParams{
		ReportID:         sqlcGroup.ReportID,
		FieldID:          sqlcGroup.FieldID,
		DateGranularity:  sqlcGroup.DateGranularity,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report group: %w", err)
	}

	return r.mapper.SQLCReportGroupToDomain(createdGroup), nil
}

func (r *ReportRepository) UpdateReportGroup(ctx context.Context, group *models.ReportGroup) (*models.ReportGroup, error) {
	sqlcGroup := r.mapper.DomainReportGroupToSQLC(group)

	updatedGroup, err := r.queries.UpdateReportGroup(ctx, generated.UpdateReportGroupParams{
		ID:               sqlcGroup.ID,
		DateGranularity:  sqlcGroup.DateGranularity,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report group: %w", err)
	}

	return r.mapper.SQLCReportGroupToDomain(updatedGroup), nil
}

func (r *ReportRepository) RemoveReportGroup(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RemoveReportGroup(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove report group: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportGroups(ctx context.Context, reportID uuid.UUID) ([]*models.ReportGroup, error) {
	sqlcGroups, err := r.queries.GetReportGroups(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report groups: %w", err)
	}

	var groups []*models.ReportGroup
	for _, sqlcGroup := range sqlcGroups {
		groups = append(groups, r.mapper.SQLCReportGroupToDomain(sqlcGroup))
	}

	return groups, nil
}

// =============================================================================
// Report Sort Operations
// =============================================================================

func (r *ReportRepository) AddReportSort(ctx context.Context, sort *models.ReportSort) (*models.ReportSort, error) {
	sqlcSort := r.mapper.DomainReportSortToSQLC(sort)

	createdSort, err := r.queries.AddReportSort(ctx, generated.AddReportSortParams{
		ReportID:  sqlcSort.ReportID,
		FieldID:   sqlcSort.FieldID,
		Direction: sqlcSort.Direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add report sort: %w", err)
	}

	return r.mapper.SQLCReportSortToDomain(createdSort), nil
}

func (r *ReportRepository) UpdateReportSort(ctx context.Context, sort *models.ReportSort) (*models.ReportSort, error) {
	sqlcSort := r.mapper.DomainReportSortToSQLC(sort)

	updatedSort, err := r.queries.UpdateReportSort(ctx, generated.UpdateReportSortParams{
		ID:        sqlcSort.ID,
		Direction: sqlcSort.Direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report sort: %w", err)
	}

	return r.mapper.SQLCReportSortToDomain(updatedSort), nil
}

func (r *ReportRepository) RemoveReportSort(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RemoveReportSort(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove report sort: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportSorts(ctx context.Context, reportID uuid.UUID) ([]*models.ReportSort, error) {
	sqlcSorts, err := r.queries.GetReportSorts(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report sorts: %w", err)
	}

	var sorts []*models.ReportSort
	for _, sqlcSort := range sqlcSorts {
		sorts = append(sorts, r.mapper.SQLCReportSortToDomain(sqlcSort))
	}

	return sorts, nil
}

// =============================================================================
// Report Chart Operations
// =============================================================================

func (r *ReportRepository) SetReportChart(ctx context.Context, chart *models.ReportChart) (*models.ReportChart, error) {
	sqlcChart := r.mapper.DomainReportChartToSQLC(chart)

	createdChart, err := r.queries.SetReportChart(ctx, generated.SetReportChartParams{
		ReportID:       sqlcChart.ReportID,
		ChartType:      sqlcChart.ChartType,
		XAxisFieldID:   sqlcChart.XAxisFieldID,
		YAxisFieldID:   sqlcChart.YAxisFieldID,
		SeriesFieldID:  sqlcChart.SeriesFieldID,
		ChartOptions:   sqlcChart.ChartOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set report chart: %w", err)
	}

	return r.mapper.SQLCReportChartToDomain(createdChart), nil
}

func (r *ReportRepository) RemoveReportChart(ctx context.Context, reportID uuid.UUID) error {
	err := r.queries.RemoveReportChart(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to remove report chart: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportChart(ctx context.Context, reportID uuid.UUID) (*models.ReportChart, error) {
	sqlcChart, err := r.queries.GetReportChart(ctx, reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No chart configured
		}
		return nil, fmt.Errorf("failed to get report chart: %w", err)
	}

	return r.mapper.SQLCReportChartToDomain(sqlcChart), nil
}

// =============================================================================
// Report Permission Operations
// =============================================================================

func (r *ReportRepository) GrantReportPermission(ctx context.Context, permission *models.ReportPermission) (*models.ReportPermission, error) {
	sqlcPermission := r.mapper.DomainReportPermissionToSQLC(permission)

	createdPermission, err := r.queries.GrantReportPermission(ctx, generated.GrantReportPermissionParams{
		ReportID:        sqlcPermission.ReportID,
		PrincipalType:   sqlcPermission.PrincipalType,
		PrincipalID:     sqlcPermission.PrincipalID,
		PermissionLevel: sqlcPermission.PermissionLevel,
		GrantedBy:       sqlcPermission.GrantedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to grant report permission: %w", err)
	}

	return r.mapper.SQLCReportPermissionToDomain(createdPermission), nil
}

func (r *ReportRepository) RevokeReportPermission(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RevokeReportPermission(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to revoke report permission: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetReportPermissions(ctx context.Context, reportID uuid.UUID) ([]*models.ReportPermission, error) {
	sqlcPermissions, err := r.queries.GetReportPermissions(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report permissions: %w", err)
	}

	var permissions []*models.ReportPermission
	for _, sqlcPermission := range sqlcPermissions {
		permissions = append(permissions, r.mapper.SQLCReportPermissionToDomain(sqlcPermission))
	}

	return permissions, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

func (r *ReportRepository) loadReportComponents(ctx context.Context, report *models.Report) error {
	// Load fields
	fields, err := r.GetReportFields(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report fields: %w", err)
	}
	report.Fields = fields

	// Load filters
	filters, err := r.GetReportFilters(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report filters: %w", err)
	}
	report.Filters = filters

	// Load groups
	groups, err := r.GetReportGroups(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report groups: %w", err)
	}
	report.Groups = groups

	// Load sorts
	sorts, err := r.GetReportSorts(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report sorts: %w", err)
	}
	report.Sorts = sorts

	// Load chart
	chart, err := r.GetReportChart(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report chart: %w", err)
	}
	report.Chart = chart

	// Load permissions
	permissions, err := r.GetReportPermissions(ctx, report.ID)
	if err != nil {
		return fmt.Errorf("failed to load report permissions: %w", err)
	}
	report.Permissions = permissions

	return nil
}