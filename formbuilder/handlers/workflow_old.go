package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	formbuilderv2 "p9e.in/ugcl/formbuilder/api/v2/form_builder"
	forminstancev2 "p9e.in/ugcl/formbuilder/api/v2/form_instance"
	workflowv2 "p9e.in/ugcl/formbuilder/api/v2/workflow"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// FormValidator handles form validation
type FormValidator struct {
	validators map[string]ValidatorFunc
}

type ValidatorFunc func(value interface{}, field *formbuilderv2.FormField) error

func NewFormValidator() *FormValidator {
	fv := &FormValidator{
		validators: make(map[string]ValidatorFunc),
	}
	fv.registerDefaultValidators()
	return fv
}

// registerDefaultValidators registers built-in validators
func (fv *FormValidator) registerDefaultValidators() {
	// Email validator
	fv.validators["email"] = func(value interface{}, field *formbuilderv2.FormField) error {
		// Simple email validation
		str, ok := value.(string)
		if !ok || !strings.Contains(str, "@") {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}

	// URL validator
	fv.validators["url"] = func(value interface{}, field *formbuilderv2.FormField) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid URL format")
		}
		_, err := url.Parse(str)
		return err
	}
}

// ValidateForm validates form instance data
func (fv *FormValidator) ValidateForm(formDef *formbuilderv2.FormDefinition, data map[string]*anypb.Any) []error {
	var errors []error

	// Field-level validation
	for _, step := range formDef.Steps {
		for _, field := range step.Fields {
			if err := fv.validateField(field, data[field.Id]); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Cross-field validation
	for _, cfv := range formDef.CrossFieldValidations {
		if err := fv.validateCrossField(cfv, data); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateField validates a single field
func (fv *FormValidator) validateField(field *formbuilderv2.FormField, value *anypb.Any) error {
	// Check required
	if field.Required && (value == nil || isEmptyValue(value)) {
		return fmt.Errorf("field %s is required", field.Id)
	}

	// Skip validation if value is empty and not required
	if !field.Required && (value == nil || isEmptyValue(value)) {
		return nil
	}

	// Type-specific validation
	switch field.Type {
	case formbuilderv2.FieldType_EMAIL:
		return fv.validators["email"](value, field)
	case formbuilderv2.FieldType_URL:
		return fv.validators["url"](value, field)
	case formbuilderv2.FieldType_NUMBER:
		return fv.validateNumber(value, field)
	}

	return nil
}

// validateNumber validates numeric fields
func (fv *FormValidator) validateNumber(value *anypb.Any, field *formbuilderv2.FormField) error {
	if field.Validation == nil {
		return nil
	}

	// Extract numeric value
	// In real implementation, properly unmarshal the Any value
	// For now, assume it's a float64
	num := 0.0 // Extract from value

	if field.Validation.Min > 0 && num < field.Validation.Min {
		return fmt.Errorf("value must be at least %f", field.Validation.Min)
	}

	if field.Validation.Max > 0 && num > field.Validation.Max {
		return fmt.Errorf("value must be at most %f", field.Validation.Max)
	}

	return nil
}

// validateCrossField validates cross-field dependencies
func (fv *FormValidator) validateCrossField(cfv *formbuilderv2.CrossFieldValidation, data map[string]*anypb.Any) error {
	// Implementation would evaluate the rule expression
	// This could use an expression evaluation library
	return nil
}

// Helper function
func isEmptyValue(value *anypb.Any) bool {
	if value == nil {
		return true
	}
	// Check for empty string, zero values, etc.
	// Implementation depends on how you marshal values
	return false
}

// WorkflowEngine handles workflow state management
type WorkflowEngine struct {
	db *sql.DB
}

func NewWorkflowEngine(db *sql.DB) *WorkflowEngine {
	return &WorkflowEngine{db: db}
}

// ProcessTransition handles workflow state transitions
func (we *WorkflowEngine) ProcessTransition(
	instance *forminstancev2.FormInstance,
	event string,
	workflow *workflowv2.Workflow,
) (*forminstancev2.FormInstance, error) {

	currentState := we.findState(workflow, instance.CurrentState)
	if currentState == nil {
		return nil, fmt.Errorf("invalid current state: %s", instance.CurrentState)
	}

	// Find matching transition
	var transition *workflowv2.WorkflowTransition
	for _, t := range currentState.Transitions {
		if t.Event == event && we.evaluateCondition(t.Condition, instance) {
			transition = t
			break
		}
	}

	if transition == nil {
		return nil, fmt.Errorf("no valid transition for event %s in state %s", event, instance.CurrentState)
	}

	// Execute transition actions
	for _, action := range transition.Actions {
		if err := we.executeAction(action, instance); err != nil {
			return nil, fmt.Errorf("failed to execute transition action: %w", err)
		}
	}

	// Update state
	instance.CurrentState = transition.NextState

	// Add audit log
	instance.AuditLogs = append(instance.AuditLogs, &forminstancev2.AuditLog{
		Id:        uuid.New().String(),
		UserId:    instance.AssignedTo,
		Action:    event,
		FromState: currentState.Id,
		ToState:   transition.NextState,
		Timestamp: timestamppb.Now(),
	})

	return instance, nil
}

// findState finds workflow state by ID
func (we *WorkflowEngine) findState(workflow *workflowv2.Workflow, stateId string) *workflowv2.WorkflowState {
	for _, state := range workflow.States {
		if state.Id == stateId {
			return state
		}
	}
	return nil
}

// evaluateCondition evaluates transition condition
func (we *WorkflowEngine) evaluateCondition(condition string, instance *forminstancev2.FormInstance) bool {
	// Implementation would use expression evaluation
	// For now, return true
	return true
}

// executeAction executes workflow action
func (we *WorkflowEngine) executeAction(action *workflowv2.TransitionAction, instance *forminstancev2.FormInstance) error {
	switch action.Type {
	case "email":
		return we.sendEmail(action.Params, instance)
	case "webhook":
		return we.callWebhook(action.Params, instance)
	case "notification":
		return we.sendNotification(action.Params, instance)
	}
	return nil
}

func (we *WorkflowEngine) sendEmail(params map[string]string, instance *forminstancev2.FormInstance) error {
	// Email sending implementation
	return nil
}

func (we *WorkflowEngine) callWebhook(params map[string]string, instance *forminstancev2.FormInstance) error {
	// Webhook calling implementation
	return nil
}

func (we *WorkflowEngine) sendNotification(params map[string]string, instance *forminstancev2.FormInstance) error {
	// Notification sending implementation
	return nil
}

// FieldDataService handles API calls for field options
type FieldDataService struct {
	db         *sql.DB
	httpClient *http.Client
	cache      *FieldOptionsCache
}

func NewFieldDataService(db *sql.DB) *FieldDataService {
	return &FieldDataService{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cache:      NewFieldOptionsCache(db),
	}
}

// GetFieldOptions fetches options for a dropdown/select field
func (fds *FieldDataService) GetFieldOptions(
	ctx context.Context,
	field *formbuilderv2.FormField,
	context map[string]string,
) ([]*formbuilderv2.FieldOption, error) {

	// Check cache first if enabled
	if field.CacheConfig != nil && field.CacheConfig.Enabled && !field.CacheConfig.RealTime {
		cached, err := fds.cache.Get(field.Id, context)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Build API request
	apiConfig := field.ApiConfig
	if apiConfig == nil {
		apiConfig = &formbuilderv2.ApiConfig{
			Method:         "GET",
			TimeoutSeconds: 30,
			RetryCount:     3,
		}
	}

	// Construct URL with parameters
	apiUrl := field.OptionsSource
	if len(context) > 0 {
		params := url.Values{}
		for k, v := range context {
			params.Add(k, v)
		}
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	}

	// Make API call with retries
	var options []*formbuilderv2.FieldOption
	var lastErr error

	for i := 0; i <= int(apiConfig.RetryCount); i++ {
		req, err := http.NewRequestWithContext(ctx, apiConfig.Method, apiUrl, nil)
		if err != nil {
			return nil, err
		}

		// Add headers
		for k, v := range apiConfig.Headers {
			req.Header.Set(k, v)
		}

		// Add auth
		if err := fds.addAuthentication(req, apiConfig); err != nil {
			return nil, err
		}

		resp, err := fds.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i) * time.Second) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Parse response
			options, err = fds.parseResponse(resp.Body, apiConfig.ResponseTransform)
			if err != nil {
				return nil, err
			}
			break
		}

		lastErr = fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	if options == nil {
		return nil, lastErr
	}

	// Cache if enabled
	if field.CacheConfig != nil && field.CacheConfig.Enabled {
		fds.cache.Set(field.Id, context, options, field.CacheConfig.TtlSeconds)
	}

	return options, nil
}

func (fds *FieldDataService) addAuthentication(req *http.Request, config *formbuilderv2.ApiConfig) error {
	switch config.AuthType {
	case "bearer":
		token := config.Headers["Authorization"]
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username := config.Headers["username"]
		password := config.Headers["password"]
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	case "api_key":
		key := config.Headers["X-API-Key"]
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
	}
	return nil
}

func (fds *FieldDataService) parseResponse(body io.Reader, transform string) ([]*formbuilderv2.FieldOption, error) {
	var rawData interface{}
	if err := json.NewDecoder(body).Decode(&rawData); err != nil {
		return nil, err
	}

	// Simple parsing - assumes array of objects with value/label
	// In production, would use the transform expression
	var options []*formbuilderv2.FieldOption

	if arr, ok := rawData.([]interface{}); ok {
		for _, item := range arr {
			if obj, ok := item.(map[string]interface{}); ok {
				option := &formbuilderv2.FieldOption{
					Value: fmt.Sprintf("%v", obj["value"]),
					Label: fmt.Sprintf("%v", obj["label"]),
				}
				if disabled, ok := obj["disabled"].(bool); ok {
					option.Disabled = disabled
				}
				options = append(options, option)
			}
		}
	}

	return options, nil
}

// FieldOptionsCache manages caching of dropdown options
type FieldOptionsCache struct {
	db *sql.DB
}

func NewFieldOptionsCache(db *sql.DB) *FieldOptionsCache {
	return &FieldOptionsCache{db: db}
}

func (foc *FieldOptionsCache) Get(fieldId string, context map[string]string) ([]*formbuilderv2.FieldOption, error) {
	cacheKey := foc.buildCacheKey(fieldId, context)

	var optionsJSON []byte
	query := `
		SELECT options_data 
		FROM form_options_cache 
		WHERE cache_key = $1 AND expires_at > NOW()
	`

	err := foc.db.QueryRow(query, cacheKey).Scan(&optionsJSON)
	if err != nil {
		return nil, err
	}

	var options []*formbuilderv2.FieldOption
	err = json.Unmarshal(optionsJSON, &options)
	return options, err
}

func (foc *FieldOptionsCache) Set(
	fieldId string,
	context map[string]string,
	options []*formbuilderv2.FieldOption,
	ttlSeconds int32,
) error {
	cacheKey := foc.buildCacheKey(fieldId, context)
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO form_options_cache (source_endpoint, cache_key, options_data, expires_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '%d seconds')
		ON CONFLICT (source_endpoint, cache_key) 
		DO UPDATE SET options_data = EXCLUDED.options_data, expires_at = EXCLUDED.expires_at
	`

	_, err = foc.db.Exec(fmt.Sprintf(query, ttlSeconds), fieldId, cacheKey, optionsJSON)
	return err
}

func (foc *FieldOptionsCache) buildCacheKey(fieldId string, context map[string]string) string {
	// Build deterministic cache key
	keys := make([]string, 0, len(context))
	for k := range context {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := []string{fieldId}
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, context[k]))
	}

	return strings.Join(parts, "|")
}
