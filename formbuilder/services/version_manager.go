package services

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/anypb"
)

// VersionManager handles form versioning
type VersionManager struct {
	db *sql.DB
}

func NewVersionManager(db *sql.DB) *VersionManager {
	return &VersionManager{db: db}
}

// GetTableColumns gets the actual columns in the table
func (vm *VersionManager) GetTableColumns(tableName string) ([]string, error) {
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1 
		AND column_name NOT IN ('id', 'form_id', 'form_version', 'current_state', 
			'created_by', 'assigned_to', 'created_at', 'updated_at', 'metadata', 'extra_data')
	`

	rows, err := vm.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// SplitCoreAndExtraFields separates fields into core (table columns) and extra (JSONB)
func (vm *VersionManager) SplitCoreAndExtraFields(
	submitted map[string]*anypb.Any,
	coreFields []string,
) (map[string]*anypb.Any, map[string]*anypb.Any) {
	core := make(map[string]*anypb.Any)
	extra := make(map[string]*anypb.Any)

	for k, v := range submitted {
		if contains(coreFields, k) {
			core[k] = v
		} else {
			extra[k] = v
		}
	}

	return core, extra
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
