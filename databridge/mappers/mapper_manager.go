// =============================================================================
// mappers/mapper_manager.go - Central mapper manager
// =============================================================================
package mappers

import (
	"go.uber.org/fx"
)

// NewMapperManager creates a new mapper manager instance
func NewMapperManager(
	schemaMapper ISchemaMapper,
	tableMapper ITableMapper,
	columnMapper IColumnMapper,
	mappingMapper IMappingMapper,
	jobMapper IJobMapper,
) *MapperManager {
	return &MapperManager{
		Schema:  schemaMapper,
		Table:   tableMapper,
		Column:  columnMapper,
		Mapping: mappingMapper,
		Job:     jobMapper,
	}
}

// MapperModule provides dependency injection for mappers
var MapperModule = fx.Module("mappers",
	fx.Provide(
		fx.Annotate(
			NewSchemaMapper,
			fx.As(new(ISchemaMapper)),
		),
		fx.Annotate(
			NewTableMapper,
			fx.As(new(ITableMapper)),
		),
		fx.Annotate(
			NewColumnMapper,
			fx.As(new(IColumnMapper)),
		),
		fx.Annotate(
			NewMappingMapper,
			fx.As(new(IMappingMapper)),
		),
		fx.Annotate(
			NewJobMapper,
			fx.As(new(IJobMapper)),
		),
		NewMapperManager,
	),
)