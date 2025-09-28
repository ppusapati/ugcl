package mappers

import (
	"encoding/json"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/insighthub/api/proto"
	"p9e.in/ugcl/insighthub/db/generated"
	"p9e.in/ugcl/insighthub/models"
)

// ReportMapper handles conversions between domain models and protobuf models
type ReportMapper struct{}

// NewReportMapper creates a new report mapper instance
func NewReportMapper() *ReportMapper {
	return &ReportMapper{}
}

// =============================================================================
// Report Mappings: Domain ↔ Proto
// =============================================================================

func (m *ReportMapper) DomainReportToProto(report *models.Report) *pb.Report {
	var metadata *structpb.Struct
	if report.Metadata != nil {
		metadata, _ = structpb.NewStruct(report.Metadata)
	}

	protoReport := &pb.Report{
		Id:          report.ID.String(),
		Name:        report.Name,
		Description: report.Description,
		Category:    report.Category,
		Tags:        report.Tags,
		IsTemplate:  report.IsTemplate,
		IsActive:    report.IsActive,
		CreatedBy:   report.CreatedBy,
		CreatedAt:   timestamppb.New(report.CreatedAt),
		UpdatedBy:   report.UpdatedBy,
		UpdatedAt:   timestamppb.New(report.UpdatedAt),
		Version:     report.Version,
		Metadata:    metadata,
	}

	// Convert associated components
	for _, field := range report.Fields {
		protoReport.Fields = append(protoReport.Fields, m.DomainReportFieldToProto(field))
	}

	for _, filter := range report.Filters {
		protoReport.Filters = append(protoReport.Filters, m.DomainReportFilterToProto(filter))
	}

	for _, group := range report.Groups {
		protoReport.Groups = append(protoReport.Groups, m.DomainReportGroupToProto(group))
	}

	for _, sort := range report.Sorts {
		protoReport.Sorts = append(protoReport.Sorts, m.DomainReportSortToProto(sort))
	}

	if report.Chart != nil {
		protoReport.Chart = m.DomainReportChartToProto(report.Chart)
	}

	return protoReport
}

func (m *ReportMapper) ProtoReportToDomain(proto *pb.Report) *models.Report {
	var metadata map[string]interface{}
	if proto.Metadata != nil {
		metadata = proto.Metadata.AsMap()
	}

	reportID, _ := uuid.Parse(proto.Id)

	return &models.Report{
		ID:          reportID,
		Name:        proto.Name,
		Description: proto.Description,
		Category:    proto.Category,
		Tags:        proto.Tags,
		IsTemplate:  proto.IsTemplate,
		IsActive:    proto.IsActive,
		CreatedBy:   proto.CreatedBy,
		CreatedAt:   proto.CreatedAt.AsTime(),
		UpdatedBy:   proto.UpdatedBy,
		UpdatedAt:   proto.UpdatedAt.AsTime(),
		Version:     proto.Version,
		Metadata:    metadata,
	}
}

// =============================================================================
// Report Field Mappings
// =============================================================================

func (m *ReportMapper) DomainReportFieldToProto(field *models.ReportField) *pb.ReportField {
	var formattingRules *structpb.Struct
	if field.FormattingRules != nil {
		formattingRules, _ = structpb.NewStruct(field.FormattingRules)
	}

	return &pb.ReportField{
		Id:                field.ID.String(),
		ReportId:          field.ReportID.String(),
		FieldId:           field.FieldID.String(),
		Alias:             field.Alias,
		AggregateFunction: field.AggregateFunction,
		OrderIndex:        field.OrderIndex,
		IsVisible:         field.IsVisible,
		FormattingRules:   formattingRules,
		CreatedAt:         timestamppb.New(field.CreatedAt),
	}
}

func (m *ReportMapper) ToProtoReportField(field generated.ReportField) *pb.ReportField {
	var formattingRules *structpb.Struct
	if field.FormattingRules != nil {
		var rulesMap map[string]interface{}
		if err := json.Unmarshal(field.FormattingRules, &rulesMap); err == nil {
			formattingRules, _ = structpb.NewStruct(rulesMap)
		}
	}

	return &pb.ReportField{
		Id:                field.ID.String(),
		ReportId:          field.ReportID.String(),
		FieldId:           field.FieldID.String(),
		Alias:             field.Alias.String,
		AggregateFunction: field.AggregateFunction.String,
		OrderIndex:        field.OrderIndex,
		IsVisible:         field.IsVisible,
		FormattingRules:   formattingRules,
		CreatedAt:         timestamppb.New(field.CreatedAt),
	}
}

// =============================================================================
// Report Filter Mappings
// =============================================================================

func (m *ReportMapper) DomainReportFilterToProto(filter *models.ReportFilter) *pb.ReportFilter {
	return &pb.ReportFilter{
		Id:              filter.ID.String(),
		ReportId:        filter.ReportID.String(),
		FieldId:         filter.FieldID.String(),
		Operator:        filter.Operator,
		Value:           filter.Value,
		ValueType:       filter.ValueType,
		LogicalOperator: filter.LogicalOperator,
		GroupIndex:      filter.GroupIndex,
		OrderIndex:      filter.OrderIndex,
		IsActive:        filter.IsActive,
		CreatedAt:       timestamppb.New(filter.CreatedAt),
	}
}

func (m *ReportMapper) ToProtoReportFilter(filter generated.ReportFilter) *pb.ReportFilter {
	return &pb.ReportFilter{
		Id:              filter.ID.String(),
		ReportId:        filter.ReportID.String(),
		FieldId:         filter.FieldID.String(),
		Operator:        filter.Operator,
		Value:           filter.Value,
		ValueType:       filter.ValueType,
		LogicalOperator: filter.LogicalOperator.String,
		GroupIndex:      filter.GroupIndex.Int32,
		OrderIndex:      filter.OrderIndex,
		IsActive:        filter.IsActive,
		CreatedAt:       timestamppb.New(filter.CreatedAt),
	}
}

// =============================================================================
// Report Group Mappings
// =============================================================================

func (m *ReportMapper) DomainReportGroupToProto(group *models.ReportGroup) *pb.ReportGroup {
	return &pb.ReportGroup{
		Id:              group.ID.String(),
		ReportId:        group.ReportID.String(),
		FieldId:         group.FieldID.String(),
		OrderIndex:      group.OrderIndex,
		DateGranularity: group.DateGranularity,
		CreatedAt:       timestamppb.New(group.CreatedAt),
	}
}

func (m *ReportMapper) ToProtoReportGroup(group generated.ReportGroup) *pb.ReportGroup {
	return &pb.ReportGroup{
		Id:              group.ID.String(),
		ReportId:        group.ReportID.String(),
		FieldId:         group.FieldID.String(),
		OrderIndex:      group.OrderIndex,
		DateGranularity: group.DateGranularity.String,
		CreatedAt:       timestamppb.New(group.CreatedAt),
	}
}

// =============================================================================
// Report Sort Mappings
// =============================================================================

func (m *ReportMapper) DomainReportSortToProto(sort *models.ReportSort) *pb.ReportSort {
	return &pb.ReportSort{
		Id:         sort.ID.String(),
		ReportId:   sort.ReportID.String(),
		FieldId:    sort.FieldID.String(),
		Direction:  sort.Direction,
		OrderIndex: sort.OrderIndex,
		CreatedAt:  timestamppb.New(sort.CreatedAt),
	}
}

func (m *ReportMapper) ToProtoReportSort(sort generated.ReportSort) *pb.ReportSort {
	return &pb.ReportSort{
		Id:         sort.ID.String(),
		ReportId:   sort.ReportID.String(),
		FieldId:    sort.FieldID.String(),
		Direction:  sort.Direction,
		OrderIndex: sort.OrderIndex,
		CreatedAt:  timestamppb.New(sort.CreatedAt),
	}
}

// =============================================================================
// Report Chart Mappings
// =============================================================================

func (m *ReportMapper) DomainReportChartToProto(chart *models.ReportChart) *pb.ReportChart {
	var chartOptions *structpb.Struct
	if chart.ChartOptions != nil {
		chartOptions, _ = structpb.NewStruct(chart.ChartOptions)
	}

	var xAxisFieldID, yAxisFieldID, seriesFieldID string
	if chart.XAxisFieldID != uuid.Nil {
		xAxisFieldID = chart.XAxisFieldID.String()
	}
	if chart.YAxisFieldID != uuid.Nil {
		yAxisFieldID = chart.YAxisFieldID.String()
	}
	if chart.SeriesFieldID != uuid.Nil {
		seriesFieldID = chart.SeriesFieldID.String()
	}

	return &pb.ReportChart{
		Id:            chart.ID.String(),
		ReportId:      chart.ReportID.String(),
		ChartType:     chart.ChartType,
		XAxisFieldId:  xAxisFieldID,
		YAxisFieldId:  yAxisFieldID,
		SeriesFieldId: seriesFieldID,
		ChartOptions:  chartOptions,
		CreatedAt:     timestamppb.New(chart.CreatedAt),
	}
}

func (m *ReportMapper) ToProtoReportChart(chart generated.ReportChart) *pb.ReportChart {
	var chartOptions *structpb.Struct
	if chart.ChartOptions != nil {
		var optionsMap map[string]interface{}
		if err := json.Unmarshal(chart.ChartOptions, &optionsMap); err == nil {
			chartOptions, _ = structpb.NewStruct(optionsMap)
		}
	}

	var xAxisFieldID, yAxisFieldID, seriesFieldID string
	if chart.XAxisFieldID.Valid {
		xAxisFieldID = chart.XAxisFieldID.UUID.String()
	}
	if chart.YAxisFieldID.Valid {
		yAxisFieldID = chart.YAxisFieldID.UUID.String()
	}
	if chart.SeriesFieldID.Valid {
		seriesFieldID = chart.SeriesFieldID.UUID.String()
	}

	return &pb.ReportChart{
		Id:            chart.ID.String(),
		ReportId:      chart.ReportID.String(),
		ChartType:     chart.ChartType,
		XAxisFieldId:  xAxisFieldID,
		YAxisFieldId:  yAxisFieldID,
		SeriesFieldId: seriesFieldID,
		ChartOptions:  chartOptions,
		CreatedAt:     timestamppb.New(chart.CreatedAt),
	}
}