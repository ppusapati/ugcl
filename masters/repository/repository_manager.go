// =============================================================================
// repository/repository_manager.go - Repository manager for Masters module
// =============================================================================
package repository

import (
	db "masters/db/generated"
)

type RepositoryManager struct {
	Schema ISchemaRepository
	Table  ITableRepository
	Column IColumnRepository
}

func NewRepositoryManager(queries *db.Queries) *RepositoryManager {
	return &RepositoryManager{
		Schema: NewSchemaRepository(queries),
		Table:  NewTableRepository(queries),
		Column: NewColumnRepository(queries),
	}
}