// =============================================================================
// internal/services/form_instance_service.go
// =============================================================================
package services

import (
	"context"
	"encoding/json"
	"fmt"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/repository"
	"p9e.in/ugcl/formbuilder/utils"
	"p9e.in/ugcl/formbuilder/validators"
	"p9e.in/ugcl/packages/database/sqlc"

	"strings"

	"github.com/google/uuid"
)

type formInstanceService struct {
	instanceRepo repository.IFormInstanceRepository
	formRepo     repository.IFormBuilderRepository
	workflowRepo repository.IWorkflowRepository
	dbManager    *sqlc.DatabaseManager
}

// NewFormInstanceService creates a new form instance service
func NewFormInstanceService(
	instanceRepo repository.IFormInstanceRepository,
	formRepo repository.IFormBuilderRepository,
	workflowRepo repository.IWorkflowRepository,
	dbManager *sqlc.DatabaseManager,
) IFormInstanceService {
	return &formInstanceService{
		instanceRepo: instanceRepo,
		formRepo:     formRepo,
		workflowRepo: workflowRepo,
		dbManager:    dbManager,
	}
}

// CreateAttachment implements IFormInstanceService.
func (s *formInstanceService) CreateAttachment(ctx context.Context, attachment *db.Attachment) (*db.Attachment, error) {
	if attachment == nil {
		return nil, fmt.Errorf("attachment cannot be nil")
	}

	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}

	return s.instanceRepo.CreateAttachment(ctx, attachment)
}

// CreateComment implements IFormInstanceService.
func (s *formInstanceService) CreateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error) {
	if comment == nil {
		return nil, fmt.Errorf("comment cannot be nil")
	}

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	return s.instanceRepo.CreateComment(ctx, comment)
}

// DeleteAttachment implements IFormInstanceService.
func (s *formInstanceService) DeleteAttachment(ctx context.Context, attachmentID string) error {
	if attachmentID == "" {
		return fmt.Errorf("attachment ID is required")
	}
	return s.instanceRepo.DeleteAttachment(ctx, utils.StringToUUID(attachmentID))
}

// DeleteComment implements IFormInstanceService.
func (s *formInstanceService) DeleteComment(ctx context.Context, commentID string) error {
	if commentID == "" {
		return fmt.Errorf("comment ID is required")
	}
	return s.instanceRepo.DeleteComment(ctx, utils.StringToUUID(commentID))
}

// DeleteFormInstance implements IFormInstanceService.
func (s *formInstanceService) DeleteFormInstance(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return fmt.Errorf("instance ID is required")
	}
	return s.instanceRepo.DeleteFormInstance(ctx, utils.StringToUUID(instanceID))
}

// GetAttachments implements IFormInstanceService.
func (s *formInstanceService) GetAttachments(ctx context.Context, instanceID string) ([]*db.Attachment, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}
	return s.instanceRepo.GetAttachments(ctx, utils.StringToUUID(instanceID))
}

// GetAuditLogs implements IFormInstanceService.
func (s *formInstanceService) GetAuditLogs(ctx context.Context, instanceID string) ([]*db.AuditLog, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}
	return s.instanceRepo.GetAuditLogs(ctx, utils.StringToUUID(instanceID))
}

// GetAuditLogsByUser implements IFormInstanceService.
func (s *formInstanceService) GetAuditLogsByUser(ctx context.Context, userID string, page int32, pageSize int32) ([]*db.AuditLog, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.instanceRepo.GetAuditLogsByUser(ctx, utils.StringToUUID(userID), page, pageSize)
}

// GetComments implements IFormInstanceService.
func (s *formInstanceService) GetComments(ctx context.Context, instanceID string) ([]*db.Comment, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}
	return s.instanceRepo.GetComments(ctx, utils.StringToUUID(instanceID))
}

// GetFormInstancesByAssignee implements IFormInstanceService.
func (s *formInstanceService) GetFormInstancesByAssignee(ctx context.Context, assignedTo string, page int32, pageSize int32) (*ListInstancesResult, error) {
	if assignedTo == "" {
		return nil, fmt.Errorf("assigned to is required")
	}
	instances, err := s.instanceRepo.GetFormInstancesByAssignee(ctx, assignedTo, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListInstancesResult{
		Instances: instances,
		Total:     int64(len(instances)),
	}, nil
}

// GetFormInstancesByState implements IFormInstanceService.
func (s *formInstanceService) GetFormInstancesByState(ctx context.Context, formID string, state string, page int32, pageSize int32) (*ListInstancesResult, error) {
	if formID == "" {
		return nil, fmt.Errorf("form ID is required")
	}
	instances, err := s.instanceRepo.GetFormInstancesByState(ctx, utils.StringToUUID(formID), state, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListInstancesResult{
		Instances: instances,
		Total:     int64(len(instances)),
	}, nil
}

// ListFormInstances implements IFormInstanceService.
func (s *formInstanceService) ListFormInstances(ctx context.Context, formID string, page int32, pageSize int32) (*ListInstancesResult, error) {
	if formID == "" {
		return nil, fmt.Errorf("form ID is required")
	}
	instances, err := s.instanceRepo.ListFormInstances(ctx, utils.StringToUUID(formID), page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListInstancesResult{
		Instances: instances,
		Total:     int64(len(instances)),
	}, nil
}

// UpdateComment implements IFormInstanceService.
func (s *formInstanceService) UpdateComment(ctx context.Context, commentID string, text string) (*db.Comment, error) {
	if commentID == "" {
		return nil, fmt.Errorf("comment ID is required")
	}
	comment := &db.Comment{
		ID:     *utils.StringToUUIDPtr(commentID),
		Text:   text,
		UserID: "",
	}
	return s.instanceRepo.UpdateComment(ctx, comment)
}

// UpdateFormInstance implements IFormInstanceService.
func (s *formInstanceService) UpdateFormInstance(ctx context.Context, instanceID string,
	fieldValues map[string]interface{}, action string, userID uuid.UUID) (*SubmitFormResult, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}
	instance := &SubmitFormResult{
		InstanceID:   utils.StringToUUID(instanceID),
		CurrentState: s.determineInitialState(action),
		Success:      true,
		Errors:       []validators.ValidationError{},
	}
	return instance, nil
}

func (s *formInstanceService) SubmitForm(ctx context.Context, formID string, fieldValues map[string]interface{}, action string, userID uuid.UUID) (*SubmitFormResult, error) {
	// Parse form ID
	id, err := uuid.Parse(formID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid form ID format")
	}

	// Get form definition for validation
	form, err := s.formRepo.GetForm(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get form definition: %w", err)
	}

	// Validate field values against form definition
	validationErrors := s.validateFieldValues(form, fieldValues)
	if len(validationErrors) > 0 {
		return &SubmitFormResult{
			Success: false,
			Errors:  validationErrors,
		}, nil
	}

	// Convert field values to JSON for storage
	fieldValuesJSON, err := json.Marshal(fieldValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field values: %w", err)
	}

	// Create metadata
	metadata := map[string]string{
		"submission_action": action,
		// "user_agent":        extractUserAgent(ctx),
		// "ip_address":        extractIPAddress(ctx).String(),
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Determine initial state from workflow or default
	initialState, err := s.getWorkflowInitialState(ctx, form, action)
	if err != nil {
		return nil, fmt.Errorf("failed to determine initial state: %w", err)
	}

	// Create form instance
	instance := &db.FormInstance{
		FormID:       id,
		CurrentState: initialState,
		FieldValues:  fieldValuesJSON,
		CreatedBy:    userID.String(),
		Metadata:     metadataJSON,
	}

	// Save instance
	createdInstance, err := s.instanceRepo.CreateFormInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create form instance: %w", err)
	}

	// Insert data into dynamic table if it exists
	if err := s.insertIntoDynamicTable(ctx, form, createdInstance.ID, fieldValues, userID.String()); err != nil {
		// Log warning but don't fail the submission
		fmt.Printf("Warning: Failed to insert into dynamic table: %v\n", err)
	}

	// Create audit log for submission
	auditLog := &db.AuditLog{
		InstanceID: createdInstance.ID,
		UserID:     userID.String(),
		Action:     action,
		ToState:    &createdInstance.CurrentState,
		Changes:    fieldValuesJSON,
		// IpAddress:  extractIPAddressPtr(ctx),
		// UserAgent:  utils.StringPtr(extractUserAgent(ctx)),
	}

	_, err = s.instanceRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the submission
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return &SubmitFormResult{
		InstanceID:   createdInstance.ID,
		CurrentState: createdInstance.CurrentState,
		Success:      true,
		Errors:       []validators.ValidationError{},
	}, nil
}

func (s *formInstanceService) GetFormInstance(ctx context.Context, instanceID string) (*db.FormInstance, error) {
	// Parse instance ID
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid instance ID format")
	}

	return s.instanceRepo.GetFormInstance(ctx, id)
}

func (s *formInstanceService) GetFormInstanceWithRelated(ctx context.Context, instanceID string) (*FormInstanceWithRelatedData, error) {
	// Parse instance ID
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, NewServiceError("INVALID_UUID", "Invalid instance ID format")
	}

	// Get main instance
	instance, err := s.instanceRepo.GetFormInstance(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get form instance: %w", err)
	}

	// Get related data in parallel (could be optimized with goroutines)
	auditLogs, err := s.instanceRepo.GetAuditLogs(ctx, id)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to get audit logs: %v\n", err)
		auditLogs = []*db.AuditLog{}
	}

	attachments, err := s.instanceRepo.GetAttachments(ctx, id)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to get attachments: %v\n", err)
		attachments = []*db.Attachment{}
	}

	comments, err := s.instanceRepo.GetComments(ctx, id)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to get comments: %v\n", err)
		comments = []*db.Comment{}
	}

	return &FormInstanceWithRelatedData{
		Instance:    instance,
		AuditLogs:   auditLogs,
		Attachments: attachments,
		Comments:    comments,
	}, nil
}

func (s *formInstanceService) validateFieldValues(form *db.Form, fieldValues map[string]interface{}) []validators.ValidationError {
	// Delegate to validators package which parses form.Steps and validates
	return validators.ValidateFormInstanceFieldValues(form, fieldValues)
}

func (s *formInstanceService) getWorkflowInitialState(ctx context.Context, form *db.Form, action string) (string, error) {
	// If form has a workflow, use its initial state
	if form.WorkflowID.Valid {
		workflow, err := s.workflowRepo.GetWorkflow(ctx, form.WorkflowID.UUID)
		if err != nil {
			return "", fmt.Errorf("failed to get workflow: %w", err)
		}

		// For submit action, transition from initial state
		if action == "submit" {
			// Parse workflow states to find the next state after initial
			var states []WorkflowState
			if err := json.Unmarshal(workflow.States, &states); err == nil {
				// Find initial state and get its submit transition
				for _, state := range states {
					if state.ID == workflow.InitialState {
						for _, transition := range state.Transitions {
							if transition.Event == "submit" {
								return transition.NextState, nil
							}
						}
					}
				}
			}
		}

		// Default to workflow initial state
		return workflow.InitialState, nil
	}

	// Fallback to old logic if no workflow
	return s.determineInitialState(action), nil
}

func (s *formInstanceService) determineInitialState(action string) string {
	switch action {
	case "submit":
		return "Submitted"
	case "reject":
		return "Rejected"
	case "approve":
		return "Approved"
	default:
		return "Draft"
	}
}

// func (s *formInstanceService) extractUserAgent(ctx context.Context) string {
// 	userAgent := ctx.Value("user_agent").(string)
// 	return userAgent
// }

func (s *formInstanceService) insertIntoDynamicTable(ctx context.Context, form *db.Form, instanceID uuid.UUID, fieldValues map[string]interface{}, userId string) error {
	fmt.Printf("DEBUG: insertIntoDynamicTable called with instanceID: %s\n", instanceID.String())
	fmt.Printf("DEBUG: Raw fieldValues received: %+v\n", fieldValues)

	// Skip if no table name specified
	if form.TableName == nil || *form.TableName == "" {
		fmt.Printf("DEBUG: No table name specified, skipping dynamic table insertion\n")
		return nil
	}

	tableName := *form.TableName
	fmt.Printf("DEBUG: Table name: %s\n", tableName)

	// Build schema-qualified table name if module is specified
	var qualifiedTableName string
	if form.Module != nil && *form.Module != "" {
		qualifiedTableName = fmt.Sprintf("%s.%s", *form.Module, tableName)
		fmt.Printf("DEBUG: Module specified: %s, qualified table name: %s\n", *form.Module, qualifiedTableName)
	} else {
		qualifiedTableName = tableName
		fmt.Printf("DEBUG: No module specified, using table name: %s\n", qualifiedTableName)
	}

	// Prepare clean field values for insertion
	cleanFieldValues := make(map[string]interface{})

	// Process each field value to ensure clean strings
	for key, value := range fieldValues {
		fmt.Printf("DEBUG: Processing field %s with value: %v (type: %T)\n", key, value, value)

		// Convert to string and clean up
		strValue := fmt.Sprintf("%v", value)
		fmt.Printf("DEBUG: String representation: '%s'\n", strValue)

		// Only include fields that should exist in the dynamic table
		switch key {
		case "first_name", "last_name", "email":
			cleanFieldValues[key] = strValue
			fmt.Printf("DEBUG: Added field %s with clean value: '%s'\n", key, strValue)
		default:
			fmt.Printf("DEBUG: Skipping field %s (not in dynamic table schema)\n", key)
		}
	}

	// Add system fields
	cleanFieldValues["form_instance_id"] = instanceID.String()
	cleanFieldValues["created_at"] = "NOW()"
	cleanFieldValues["created_by"] = userId

	fmt.Printf("DEBUG: Final clean field values to insert: %+v\n", cleanFieldValues)

	// Build column names and values for INSERT
	var columns []string
	var placeholders []string
	var values []interface{}
	i := 1

	for key, value := range cleanFieldValues {
		columns = append(columns, key)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, value)
		i++
	}

	// Create INSERT SQL
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		qualifiedTableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	fmt.Printf("DEBUG: Insert SQL: %s\n", insertSQL)
	fmt.Printf("DEBUG: Insert values: %+v\n", values)

	// Execute the INSERT statement
	_, err := s.dbManager.ExecRaw(ctx, insertSQL, values...)
	if err != nil {
		fmt.Printf("DEBUG: Insert failed with error: %v\n", err)
		return fmt.Errorf("failed to insert into dynamic table %s: %w", qualifiedTableName, err)
	}

	fmt.Printf("DEBUG: Successfully inserted into dynamic table: %s\n", qualifiedTableName)
	return nil
}
