package mappers

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"p9e.in/ugcl/insighthub/db/generated"
	"p9e.in/ugcl/insighthub/models"
)

// DomainMapper handles conversions between domain models and SQLC models
type DomainMapper struct{}

// NewDomainMapper creates a new domain mapper instance
func NewDomainMapper() *DomainMapper {
	return &DomainMapper{}
}

// =============================================================================
// Report Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportToSQLC(report *models.Report) generated.Report {
	var metadataBytes []byte
	if report.Metadata != nil {
		metadataBytes, _ = json.Marshal(report.Metadata)
	}

	return generated.Report{
		ID:          report.ID,
		Name:        report.Name,
		Description: sql.NullString{String: report.Description, Valid: report.Description != ""},
		Category:    sql.NullString{String: report.Category, Valid: report.Category != ""},
		Tags:        report.Tags,
		IsTemplate:  report.IsTemplate,
		IsActive:    report.IsActive,
		CreatedBy:   report.CreatedBy,
		CreatedAt:   report.CreatedAt,
		UpdatedBy:   sql.NullString{String: report.UpdatedBy, Valid: report.UpdatedBy != ""},
		UpdatedAt:   report.UpdatedAt,
		Version:     report.Version,
		Metadata:    metadataBytes,
	}
}

func (m *DomainMapper) SQLCReportToDomain(report generated.Report) *models.Report {
	var metadata map[string]interface{}
	if report.Metadata != nil {
		_ = json.Unmarshal(report.Metadata, &metadata)
	}

	return &models.Report{
		ID:          report.ID,
		Name:        report.Name,
		Description: report.Description.String,
		Category:    report.Category.String,
		Tags:        report.Tags,
		IsTemplate:  report.IsTemplate,
		IsActive:    report.IsActive,
		CreatedBy:   report.CreatedBy,
		CreatedAt:   report.CreatedAt,
		UpdatedBy:   report.UpdatedBy.String,
		UpdatedAt:   report.UpdatedAt,
		Version:     report.Version,
		Metadata:    metadata,
	}
}

// =============================================================================
// Report Field Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportFieldToSQLC(field *models.ReportField) generated.ReportField {
	var formattingRulesBytes []byte
	if field.FormattingRules != nil {
		formattingRulesBytes, _ = json.Marshal(field.FormattingRules)
	}

	return generated.ReportField{
		ID:                field.ID,
		ReportID:          field.ReportID,
		FieldID:           field.FieldID,
		Alias:             sql.NullString{String: field.Alias, Valid: field.Alias != ""},
		AggregateFunction: sql.NullString{String: field.AggregateFunction, Valid: field.AggregateFunction != ""},
		OrderIndex:        field.OrderIndex,
		IsVisible:         field.IsVisible,
		FormattingRules:   formattingRulesBytes,
		CreatedAt:         field.CreatedAt,
	}
}

func (m *DomainMapper) SQLCReportFieldToDomain(field generated.ReportField) *models.ReportField {
	var formattingRules map[string]interface{}
	if field.FormattingRules != nil {
		_ = json.Unmarshal(field.FormattingRules, &formattingRules)
	}

	return &models.ReportField{
		ID:                field.ID,
		ReportID:          field.ReportID,
		FieldID:           field.FieldID,
		Alias:             field.Alias.String,
		AggregateFunction: field.AggregateFunction.String,
		OrderIndex:        field.OrderIndex,
		IsVisible:         field.IsVisible,
		FormattingRules:   formattingRules,
		CreatedAt:         field.CreatedAt,
	}
}

// =============================================================================
// Report Filter Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportFilterToSQLC(filter *models.ReportFilter) generated.ReportFilter {
	return generated.ReportFilter{
		ID:              filter.ID,
		ReportID:        filter.ReportID,
		FieldID:         filter.FieldID,
		Operator:        filter.Operator,
		Value:           filter.Value,
		ValueType:       filter.ValueType,
		LogicalOperator: sql.NullString{String: filter.LogicalOperator, Valid: filter.LogicalOperator != ""},
		GroupIndex:      sql.NullInt32{Int32: filter.GroupIndex, Valid: filter.GroupIndex > 0},
		OrderIndex:      filter.OrderIndex,
		IsActive:        filter.IsActive,
		CreatedAt:       filter.CreatedAt,
	}
}

func (m *DomainMapper) SQLCReportFilterToDomain(filter generated.ReportFilter) *models.ReportFilter {
	return &models.ReportFilter{
		ID:              filter.ID,
		ReportID:        filter.ReportID,
		FieldID:         filter.FieldID,
		Operator:        filter.Operator,
		Value:           filter.Value,
		ValueType:       filter.ValueType,
		LogicalOperator: filter.LogicalOperator.String,
		GroupIndex:      filter.GroupIndex.Int32,
		OrderIndex:      filter.OrderIndex,
		IsActive:        filter.IsActive,
		CreatedAt:       filter.CreatedAt,
	}
}

// =============================================================================
// Report Group Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportGroupToSQLC(group *models.ReportGroup) generated.ReportGroup {
	return generated.ReportGroup{
		ID:              group.ID,
		ReportID:        group.ReportID,
		FieldID:         group.FieldID,
		OrderIndex:      group.OrderIndex,
		DateGranularity: sql.NullString{String: group.DateGranularity, Valid: group.DateGranularity != ""},
		CreatedAt:       group.CreatedAt,
	}
}

func (m *DomainMapper) SQLCReportGroupToDomain(group generated.ReportGroup) *models.ReportGroup {
	return &models.ReportGroup{
		ID:              group.ID,
		ReportID:        group.ReportID,
		FieldID:         group.FieldID,
		OrderIndex:      group.OrderIndex,
		DateGranularity: group.DateGranularity.String,
		CreatedAt:       group.CreatedAt,
	}
}

// =============================================================================
// Report Sort Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportSortToSQLC(sort *models.ReportSort) generated.ReportSort {
	return generated.ReportSort{
		ID:         sort.ID,
		ReportID:   sort.ReportID,
		FieldID:    sort.FieldID,
		Direction:  sort.Direction,
		OrderIndex: sort.OrderIndex,
		CreatedAt:  sort.CreatedAt,
	}
}

func (m *DomainMapper) SQLCReportSortToDomain(sort generated.ReportSort) *models.ReportSort {
	return &models.ReportSort{
		ID:         sort.ID,
		ReportID:   sort.ReportID,
		FieldID:    sort.FieldID,
		Direction:  sort.Direction,
		OrderIndex: sort.OrderIndex,
		CreatedAt:  sort.CreatedAt,
	}
}

// =============================================================================
// Report Chart Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportChartToSQLC(chart *models.ReportChart) generated.ReportChart {
	var chartOptionsBytes []byte
	if chart.ChartOptions != nil {
		chartOptionsBytes, _ = json.Marshal(chart.ChartOptions)
	}

	return generated.ReportChart{
		ID:            chart.ID,
		ReportID:      chart.ReportID,
		ChartType:     chart.ChartType,
		XAxisFieldID:  uuid.NullUUID{UUID: chart.XAxisFieldID, Valid: chart.XAxisFieldID != uuid.Nil},
		YAxisFieldID:  uuid.NullUUID{UUID: chart.YAxisFieldID, Valid: chart.YAxisFieldID != uuid.Nil},
		SeriesFieldID: uuid.NullUUID{UUID: chart.SeriesFieldID, Valid: chart.SeriesFieldID != uuid.Nil},
		ChartOptions:  chartOptionsBytes,
		CreatedAt:     chart.CreatedAt,
	}
}

func (m *DomainMapper) SQLCReportChartToDomain(chart generated.ReportChart) *models.ReportChart {
	var chartOptions map[string]interface{}
	if chart.ChartOptions != nil {
		_ = json.Unmarshal(chart.ChartOptions, &chartOptions)
	}

	var xAxisFieldID, yAxisFieldID, seriesFieldID uuid.UUID
	if chart.XAxisFieldID.Valid {
		xAxisFieldID = chart.XAxisFieldID.UUID
	}
	if chart.YAxisFieldID.Valid {
		yAxisFieldID = chart.YAxisFieldID.UUID
	}
	if chart.SeriesFieldID.Valid {
		seriesFieldID = chart.SeriesFieldID.UUID
	}

	return &models.ReportChart{
		ID:            chart.ID,
		ReportID:      chart.ReportID,
		ChartType:     chart.ChartType,
		XAxisFieldID:  xAxisFieldID,
		YAxisFieldID:  yAxisFieldID,
		SeriesFieldID: seriesFieldID,
		ChartOptions:  chartOptions,
		CreatedAt:     chart.CreatedAt,
	}
}

// =============================================================================
// Report Permission Mappings: Domain ↔ SQLC
// =============================================================================

func (m *DomainMapper) DomainReportPermissionToSQLC(permission *models.ReportPermission) generated.ReportPermission {
	return generated.ReportPermission{
		ID:              permission.ID,
		ReportID:        permission.ReportID,
		PrincipalType:   permission.PrincipalType,
		PrincipalID:     permission.PrincipalID,
		PermissionLevel: permission.PermissionLevel,
		IsActive:        permission.IsActive,
		GrantedBy:       permission.GrantedBy,
		GrantedAt:       permission.GrantedAt,
	}
}

func (m *DomainMapper) SQLCReportPermissionToDomain(permission generated.ReportPermission) *models.ReportPermission {
	return &models.ReportPermission{
		ID:              permission.ID,
		ReportID:        permission.ReportID,
		PrincipalType:   permission.PrincipalType,
		PrincipalID:     permission.PrincipalID,
		PermissionLevel: permission.PermissionLevel,
		IsActive:        permission.IsActive,
		GrantedBy:       permission.GrantedBy,
		GrantedAt:       permission.GrantedAt,
	}
}