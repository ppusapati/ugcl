package models

import (
	"time"

	"github.com/google/uuid"
)

// Report represents a report definition in the domain layer
type Report struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Tags        []string
	IsTemplate  bool
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedBy   string
	UpdatedAt   time.Time
	Version     int32
	Metadata    map[string]interface{}

	// Associated components
	Fields      []*ReportField
	Filters     []*ReportFilter
	Groups      []*ReportGroup
	Sorts       []*ReportSort
	Joins       []*ReportJoin
	Chart       *ReportChart
	Parameters  []*ReportParameter
	Permissions []*ReportPermission
}

// ReportField represents a field in a report
type ReportField struct {
	ID                uuid.UUID
	ReportID          uuid.UUID
	FieldID           uuid.UUID // References masters.data_fields.id
	Alias             string
	AggregateFunction string
	OrderIndex        int32
	IsVisible         bool
	FormattingRules   map[string]interface{}
	CreatedAt         time.Time
}

// ReportFilter represents a filter condition in a report
type ReportFilter struct {
	ID              uuid.UUID
	ReportID        uuid.UUID
	FieldID         uuid.UUID // References masters.data_fields.id
	Operator        string    // =, !=, >, <, >=, <=, IN, LIKE, BETWEEN
	Value           string
	ValueType       string // string, number, date, boolean
	LogicalOperator string // AND, OR
	GroupIndex      int32
	OrderIndex      int32
	IsActive        bool
	CreatedAt       time.Time
}

// ReportGroup represents a grouping configuration in a report
type ReportGroup struct {
	ID              uuid.UUID
	ReportID        uuid.UUID
	FieldID         uuid.UUID // References masters.data_fields.id
	OrderIndex      int32
	DateGranularity string // year, month, day, hour for date fields
	CreatedAt       time.Time
}

// ReportSort represents a sorting configuration in a report
type ReportSort struct {
	ID         uuid.UUID
	ReportID   uuid.UUID
	FieldID    uuid.UUID // References masters.data_fields.id
	Direction  string    // ASC, DESC
	OrderIndex int32
	CreatedAt  time.Time
}

// ReportJoin represents a join configuration in a report
type ReportJoin struct {
	ID           uuid.UUID
	ReportID     uuid.UUID
	LeftTableID  uuid.UUID // References masters.data_tables.id
	RightTableID uuid.UUID // References masters.data_tables.id
	JoinType     string    // INNER, LEFT, RIGHT, FULL
	LeftFieldID  uuid.UUID // References masters.data_fields.id
	RightFieldID uuid.UUID // References masters.data_fields.id
	Alias        string
	CreatedAt    time.Time
}

// ReportChart represents a chart configuration in a report
type ReportChart struct {
	ID            uuid.UUID
	ReportID      uuid.UUID
	ChartType     string // line, bar, pie, scatter, area, etc.
	XAxisFieldID  uuid.UUID
	YAxisFieldID  uuid.UUID
	SeriesFieldID uuid.UUID
	ChartOptions  map[string]interface{} // ECharts configuration
	CreatedAt     time.Time
}

// ReportParameter represents a parameter in a report
type ReportParameter struct {
	ID            uuid.UUID
	ReportID      uuid.UUID
	Name          string
	DisplayName   string
	ParameterType string // string, number, date, boolean, select
	DefaultValue  string
	IsRequired    bool
	Options       map[string]interface{} // For select type parameters
	OrderIndex    int32
	CreatedAt     time.Time
}

// ReportPermission represents access control for a report
type ReportPermission struct {
	ID              uuid.UUID
	ReportID        uuid.UUID
	PrincipalType   string // user, role
	PrincipalID     string
	PermissionLevel string // read, write, admin
	IsActive        bool
	GrantedBy       string
	GrantedAt       time.Time
}