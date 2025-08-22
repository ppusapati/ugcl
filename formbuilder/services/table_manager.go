package services

import (
	"database/sql"
	"fmt"
	"strings"

	formbuilderv2 "p9e.in/ugcl/formbuilder/api/v2/form_builder"
)

// TableManager handles dynamic table creation
type TableManager struct {
	db *sql.DB
}

func NewTableManager(db *sql.DB) *TableManager {
	return &TableManager{db: db}
}

// CreateFormTable creates a new table for the form
func (tm *TableManager) CreateFormTable(formDef *formbuilderv2.FormDefinition) error {
	tableName := tm.sanitizeTableName(formDef.Metadata.TableName)
	if tableName == "" {
		tableName = tm.sanitizeTableName(formDef.Metadata.FormId)
	}

	// Build CREATE TABLE statement
	var columns []string
	var coreFields []string

	// Add system columns
	columns = append(columns,
		"id UUID PRIMARY KEY DEFAULT uuid_generate_v7()",
		"form_id VARCHAR(255) NOT NULL",
		"form_version INT NOT NULL",
		"current_state VARCHAR(100)",
		"created_by VARCHAR(255)",
		"assigned_to VARCHAR(255)",
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"metadata JSONB",
		"extra_data JSONB", // For version-specific fields
	)

	// Add columns for core fields only
	for _, step := range formDef.Steps {
		for _, field := range step.Fields {
			if field.IsCoreField {
				columnDef := tm.getColumnDefinition(field)
				if columnDef != "" {
					columns = append(columns, columnDef)
					coreFields = append(coreFields, field.Id)
				}
			}
		}
	}

	// Store core fields in metadata
	formDef.Metadata.CoreFields = coreFields

	// Create table
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			%s
		)`, tableName, strings.Join(columns, ",\n"))

	_, err := tm.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create form table: %w", err)
	}

	// Create indexes
	if err := tm.createIndexes(tableName, formDef); err != nil {
		return err
	}

	// Create audit table
	if err := tm.createAuditTable(tableName); err != nil {
		return err
	}

	return nil
}

// getColumnDefinition returns SQL column definition for a field
func (tm *TableManager) getColumnDefinition(field *formbuilderv2.FormField) string {
	columnName := tm.sanitizeColumnName(field.Id)

	var columnType string
	switch field.Type {
	case formbuilderv2.FieldType_TEXT, formbuilderv2.FieldType_EMAIL, formbuilderv2.FieldType_URL, formbuilderv2.FieldType_PHONE:
		maxLen := 255
		if field.Validation != nil && field.Validation.MaxLength > 0 {
			maxLen = int(field.Validation.MaxLength)
		}
		columnType = fmt.Sprintf("VARCHAR(%d)", maxLen)
	case formbuilderv2.FieldType_TEXTAREA:
		columnType = "TEXT"
	case formbuilderv2.FieldType_NUMBER:
		columnType = "NUMERIC"
	case formbuilderv2.FieldType_CURRENCY:
		columnType = "DECIMAL(15,2)"
	case formbuilderv2.FieldType_DATE:
		columnType = "DATE"
	case formbuilderv2.FieldType_DATETIME:
		columnType = "TIMESTAMP"
	case formbuilderv2.FieldType_CHECKBOX:
		columnType = "BOOLEAN"
	case formbuilderv2.FieldType_DROPDOWN, formbuilderv2.FieldType_RADIO:
		columnType = "VARCHAR(255)"
	case formbuilderv2.FieldType_MULTI_SELECT, formbuilderv2.FieldType_ARRAY:
		columnType = "JSONB"
	case formbuilderv2.FieldType_FILE:
		columnType = "JSONB" // Store file metadata
	case formbuilderv2.FieldType_JSON, formbuilderv2.FieldType_NESTED_FORM:
		columnType = "JSONB"
	default:
		columnType = "TEXT"
	}

	// Add constraints
	constraints := []string{}
	if field.Required {
		constraints = append(constraints, "NOT NULL")
	}

	if field.Validation != nil && len(field.Validation.AllowedValues) > 0 {
		checkValues := []string{}
		for _, v := range field.Validation.AllowedValues {
			checkValues = append(checkValues, fmt.Sprintf("'%s'", v))
		}
		constraints = append(constraints,
			fmt.Sprintf("CHECK (%s IN (%s))", columnName, strings.Join(checkValues, ",")))
	}

	constraintStr := ""
	if len(constraints) > 0 {
		constraintStr = " " + strings.Join(constraints, " ")
	}

	return fmt.Sprintf("%s %s%s", columnName, columnType, constraintStr)
}

// createIndexes creates necessary indexes for the form table
func (tm *TableManager) createIndexes(tableName string, formDef *formbuilderv2.FormDefinition) error {
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_form_id ON %s(form_id)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_state ON %s(current_state)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_assigned_to ON %s(assigned_to)", tableName, tableName),
	}

	for _, idx := range indexes {
		_, err := tm.db.Exec(idx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// createAuditTable creates an audit table for tracking changes
func (tm *TableManager) createAuditTable(tableName string) error {
	auditTableName := tableName + "_audit"

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
			record_id UUID NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			action VARCHAR(50) NOT NULL,
			from_state VARCHAR(100),
			to_state VARCHAR(100),
			changes JSONB,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address VARCHAR(45),
			user_agent TEXT
		)`, auditTableName)

	_, err := tm.db.Exec(query)
	return err
}

// sanitizeTableName ensures table name is SQL-safe
func (tm *TableManager) sanitizeTableName(name string) string {
	// Replace non-alphanumeric characters with underscores
	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}

// sanitizeColumnName ensures column name is SQL-safe
func (tm *TableManager) sanitizeColumnName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}
