// =============================================================================
// internal/services/form_builder_service.go
// =============================================================================
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	pb "p9e.in/ugcl/formbuilder/api/v2/form_builder"
	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/mappers"
	"p9e.in/ugcl/formbuilder/repository"
	"p9e.in/ugcl/formbuilder/utils"
	validators "p9e.in/ugcl/formbuilder/validators"
	"p9e.in/ugcl/packages/database/sqlc"
)

type formBuilderService struct {
	formRepo     repository.IFormBuilderRepository
	workflowRepo repository.IWorkflowRepository
	dbManager    *sqlc.DatabaseManager
}

// NewFormBuilderService creates a new form builder service
func NewFormBuilderService(formRepo repository.IFormBuilderRepository, workflowRepo repository.IWorkflowRepository, dbManager *sqlc.DatabaseManager) IFormBuilderService {
	return &formBuilderService{
		formRepo:     formRepo,
		workflowRepo: workflowRepo,
		dbManager:    dbManager,
	}
}

func (s *formBuilderService) CreateForm(ctx context.Context, formpb *pb.FormDefinition, createTable bool) (*CreateFormResult, error) {
	form, err := mappers.ProtoToFormDefinition(formpb)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	// Validate form definition
	if err := validators.ValidateForm(form); err != nil {
		return &CreateFormResult{
			Success: false,
			Message: fmt.Sprintf("Validation failed: %s", err),
		}, nil
	}

	// Set default values if not provided
	if form.Version == "" {
		form.Version = "1.0.0"
	}
	fmt.Println("Module Name", &form.Module)
	if form.TableName == nil || *form.TableName == "" {
		tableName := fmt.Sprintf("form_%s", strings.ReplaceAll(form.ID.String(), "-", "_"))
		form.TableName = &tableName
	}

	if form.SchemaVersion == nil {
		schemaVersion := int32(1)
		form.SchemaVersion = &schemaVersion
	}

	if form.Audit == nil {
		audit := false
		form.Audit = &audit
	}
	if formpb.Workflow != nil {
		wf, err := mappers.ProtoToWorkflow(formpb.Workflow)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		createdWorkflow, err := s.workflowRepo.CreateWorkflow(ctx, wf)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		form.WorkflowID = uuid.NullUUID{UUID: createdWorkflow.ID, Valid: true}
	}
	// Create form in repository
	createdForm, err := s.formRepo.CreateForm(ctx, form)
	if err != nil {
		return nil, fmt.Errorf("failed to create form: %w", err)
	}
	createdForm.Module = &formpb.Metadata.Module
	// Create dynamic table if requested
	if createTable {
		if err := s.createDynamicTable(ctx, createdForm); err != nil {
			// Log warning but don't fail the form creation
			fmt.Printf("Warning: Failed to create dynamic table: %v\n", err)
		}
	}

	return &CreateFormResult{
		FormID:    createdForm.ID,
		TableName: utils.StringValue(createdForm.TableName),
		Success:   true,
		Message:   "Form created successfully",
	}, nil
}

func (s *formBuilderService) GetForm(ctx context.Context, formID string, version *string) (*db.Form, error) {
	// Parse UUID
	id, err := uuid.Parse(formID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Get form by version or latest
	if version != nil && *version != "" {
		return s.formRepo.GetFormByVersion(ctx, id, *version)
	}

	return s.formRepo.GetForm(ctx, id)
}

func (s *formBuilderService) UpdateForm(ctx context.Context, form *db.Form) (*CreateFormResult, error) {
	// Validate form definition
	if err := s.validateForm(form); err != nil {
		return &CreateFormResult{
			Success: false,
			Message: fmt.Sprintf("Validation failed: %s", err.Error()),
		}, nil
	}

	// Check if form exists
	existingForm, err := s.formRepo.GetForm(ctx, form.ID)
	if err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	// Preserve certain fields from existing form
	form.CreatedAt = existingForm.CreatedAt
	form.CreatedBy = existingForm.CreatedBy

	// Update form
	updatedForm, err := s.formRepo.UpdateForm(ctx, form)
	if err != nil {
		return nil, fmt.Errorf("failed to update form: %w", err)
	}

	return &CreateFormResult{
		FormID:  updatedForm.ID,
		Success: true,
		Message: "Form updated successfully",
	}, nil
}

func (s *formBuilderService) DeleteForm(ctx context.Context, formID string) error {
	// Parse UUID
	id, err := uuid.Parse(formID)
	if err != nil {
		return NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Check if form exists
	_, err = s.formRepo.GetForm(ctx, id)
	if err != nil {
		return fmt.Errorf("form not found: %w", err)
	}

	// Soft delete the form
	return s.formRepo.DeleteForm(ctx, id)
}

func (s *formBuilderService) ListForms(ctx context.Context, page, pageSize int32, filter, sortBy string) (*ListFormsResult, error) {
	// Validate pagination parameters
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	var forms []*db.Form
	var total int64
	var err error

	// Apply filter if provided
	if filter != "" {
		forms, err = s.formRepo.SearchForms(ctx, filter, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to search forms: %w", err)
		}
		// For simplicity, use total count (in production, implement filtered count)
		total, err = s.formRepo.CountForms(ctx)
	} else {
		forms, err = s.formRepo.ListForms(ctx, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list forms: %w", err)
		}
		total, err = s.formRepo.CountForms(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count forms: %w", err)
	}

	return &ListFormsResult{
		Forms:    forms,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *formBuilderService) MigrateFormVersion(ctx context.Context, formID, fromVersion, toVersion string, migrateData bool) (*MigrateFormResult, error) {
	// Parse UUID
	id, err := uuid.Parse(formID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Get source and target forms
	sourceForm, err := s.formRepo.GetFormByVersion(ctx, id, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("source form version not found: %w", err)
	}

	targetForm, err := s.formRepo.GetFormByVersion(ctx, id, toVersion)
	if err != nil {
		return nil, fmt.Errorf("target form version not found: %w", err)
	}

	warnings := []string{}

	// TODO: Implement actual migration logic
	// This would involve:
	// 1. Comparing form schemas
	// 2. Mapping field changes
	// 3. Migrating existing data if requested
	// 4. Updating form instances

	_ = sourceForm
	_ = targetForm
	_ = migrateData

	return &MigrateFormResult{
		Success:         true,
		RecordsMigrated: 0,
		Warnings:        warnings,
	}, nil
}

func (s *formBuilderService) GetFieldOptions(ctx context.Context, formID, fieldID string, context map[string]string) (*GetFieldOptionsResult, error) {
	// Parse UUID
	id, err := uuid.Parse(formID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Get form definition
	form, err := s.formRepo.GetForm(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	// Find field in form steps
	field, err := s.findFieldInForm(form, fieldID)
	if err != nil {
		return nil, err
	}

	// Generate options based on field configuration
	options, err := s.generateFieldOptions(field, context)
	if err != nil {
		return nil, fmt.Errorf("failed to generate field options: %w", err)
	}

	return &GetFieldOptionsResult{
		Options:   options,
		FromCache: false, // TODO: Implement caching
		CachedAt:  nil,
	}, nil
}

func (s *formBuilderService) RefreshFieldCache(ctx context.Context, formID, fieldID string) error {
	// Parse UUID
	id, err := uuid.Parse(formID)
	if err != nil {
		return NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Verify form and field exist
	form, err := s.formRepo.GetForm(ctx, id)
	if err != nil {
		return fmt.Errorf("form not found: %w", err)
	}

	_, err = s.findFieldInForm(form, fieldID)
	if err != nil {
		return err
	}

	// TODO: Implement cache refresh logic
	fmt.Printf("Cache refreshed for form %s, field %s\n", formID, fieldID)

	return nil
}

func (s *formBuilderService) generateFieldOptions(field map[string]interface{}, context map[string]string) ([]*FieldOption, error) {
	// Check if field has options source
	optionsSource, hasSource := field["options_source"].(string)
	if !hasSource || optionsSource == "" {
		// Return static options if available
		if options, hasOptions := field["options"].([]interface{}); hasOptions {
			result := make([]*FieldOption, len(options))
			for i, opt := range options {
				if optMap, ok := opt.(map[string]interface{}); ok {
					result[i] = &FieldOption{
						Value: utils.StringValue(utils.StringPtr(fmt.Sprintf("%v", optMap["value"]))),
						Label: utils.StringValue(utils.StringPtr(fmt.Sprintf("%v", optMap["label"]))),
					}
				}
			}
			return result, nil
		}

		return []*FieldOption{}, nil
	}

	// TODO: Implement dynamic options loading from API
	// This would involve:
	// 1. Making HTTP request to options_source
	// 2. Applying any field dependencies from context
	// 3. Transforming response to FieldOption format
	// 4. Caching results if configured

	fmt.Printf("Would load options from: %s with context: %+v\n", optionsSource, context)

	// Return mock options for now
	return []*FieldOption{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}, nil
}

// sanitizeColumnName sanitizes field IDs to be valid PostgreSQL column names
func (s *formBuilderService) sanitizeColumnName(fieldID string) string {
	// Convert to lowercase and replace invalid characters with underscores
	sanitized := strings.ToLower(fieldID)

	// Replace spaces, hyphens, and other invalid characters with underscores
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")

	// Remove any characters that aren't alphanumeric or underscore
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}

	sanitized = result.String()

	// Ensure it starts with a letter or underscore (PostgreSQL requirement)
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "_" + sanitized
	}

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "field"
	}

	return sanitized
}

// getPostgreSQLType maps form field types to PostgreSQL data types
func (s *formBuilderService) getPostgreSQLType(fieldType string) string {
	switch strings.ToLower(fieldType) {
	case "text", "textarea", "email", "url", "phone":
		return "TEXT"
	case "number", "integer":
		return "INTEGER"
	case "decimal", "float":
		return "DECIMAL(10,2)"
	case "boolean", "checkbox":
		return "BOOLEAN"
	case "date":
		return "DATE"
	case "datetime", "timestamp":
		return "TIMESTAMP WITH TIME ZONE"
	case "time":
		return "TIME"
	case "select", "radio", "dropdown":
		return "VARCHAR(255)"
	case "multiselect", "checkboxgroup":
		return "TEXT[]" // PostgreSQL array type
	case "file", "image":
		return "VARCHAR(500)" // Store file paths/URLs
	case "json", "object":
		return "JSONB"
	case "uuid":
		return "UUID"
	default:
		// Default to TEXT for unknown types
		return "TEXT"
	}
}

func (s *formBuilderService) validateForm(form *db.Form) error {
	// Basic validation
	if form.Title == "" {
		return fmt.Errorf("form title is required")
	}

	if form.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Parse and validate steps from JSON
	var steps []validators.FormStep
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		return fmt.Errorf("invalid steps JSON: %w", err)
	}

	if len(steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}

	// Validate each step and field
	for i, step := range steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: ID is required", i)
		}

		if len(step.Fields) == 0 {
			return fmt.Errorf("step %d: at least one field is required", i)
		}

		for j, field := range step.Fields {
			if field.ID == "" {
				return fmt.Errorf("step %d, field %d: ID is required", i, j)
			}
			if field.Label == "" {
				return fmt.Errorf("step %d, field %d: label is required", i, j)
			}
		}
	}

	return nil
}

func (s *formBuilderService) createDynamicTable(ctx context.Context, form *db.Form) error {
	// Skip if no table name specified
	if form.TableName == nil || *form.TableName == "" {
		return nil
	}

	tableName := *form.TableName
	fmt.Println("Module Name", &form.Module)

	// Create schema if module is specified
	if form.Module != nil && *form.Module != "" {
		createSchemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", *form.Module)
		if _, err := s.dbManager.ExecRaw(ctx, createSchemaSQL); err != nil {
			return fmt.Errorf("failed to create schema %s: %w", *form.Module, err)
		}
	}

	// Build schema-qualified table name if module is specified
	var qualifiedTableName string
	if form.Module != nil && *form.Module != "" {
		qualifiedTableName = fmt.Sprintf("%s.%s", *form.Module, tableName)
	} else {
		qualifiedTableName = tableName
	}

	// Parse form steps to extract field definitions
	var steps []validators.FormStep
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		return fmt.Errorf("failed to parse form steps: %w", err)
	}

	// Build CREATE TABLE statement
	var columns []string

	// Add standard columns
	columns = append(columns, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
	columns = append(columns, "form_instance_id UUID NOT NULL")
	columns = append(columns, "created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()")
	columns = append(columns, "updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()")
	columns = append(columns, "created_by VARCHAR(255) NOT NULL")

	// Add dynamic columns based on form fields
	for _, step := range steps {
		for _, field := range step.Fields {
			columnType := s.getPostgreSQLType(field.Type.String())
			columnName := s.sanitizeColumnName(field.ID)

			column := fmt.Sprintf("%s %s", columnName, columnType)
			if field.Required {
				column += " NOT NULL"
			}
			columns = append(columns, column)
		}
	}

	// Create the table
	createTableSQL := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s)",
		qualifiedTableName,
		strings.Join(columns, ", "),
	)

	// Execute the CREATE TABLE statement
	_, err := s.dbManager.ExecRaw(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create dynamic table %s: %w", qualifiedTableName, err)
	}

	// Create indexes for better performance
	indexSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_form_instance_id ON %s(form_instance_id)", tableName, qualifiedTableName)
	if _, err := s.dbManager.ExecRaw(ctx, indexSQL); err != nil {
		return fmt.Errorf("failed to create index on %s: %w", qualifiedTableName, err)
	}

	indexSQL = fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at)", tableName, qualifiedTableName)
	if _, err := s.dbManager.ExecRaw(ctx, indexSQL); err != nil {
		return fmt.Errorf("failed to create created_at index on %s: %w", qualifiedTableName, err)
	}

	return nil
}

func (s *formBuilderService) findFieldInForm(form *db.Form, fieldID string) (map[string]interface{}, error) {
	// Parse steps from JSON
	var steps []map[string]interface{}
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		return nil, fmt.Errorf("invalid form structure: %w", err)
	}

	// Search for field in all steps
	for _, step := range steps {
		if fields, ok := step["fields"].([]interface{}); ok {
			for _, f := range fields {
				if field, ok := f.(map[string]interface{}); ok {
					if id, exists := field["id"]; exists && id == fieldID {
						return field, nil
					}
				}
			}
		}
	}

	return nil, NewServiceError("FIELD_NOT_FOUND", "Field not found in form")
}
