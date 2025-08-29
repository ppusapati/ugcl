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

	"github.com/google/uuid"
)

type formInstanceService struct {
	instanceRepo repository.IFormInstanceRepository
	formRepo     repository.IFormBuilderRepository
}

// NewFormInstanceService creates a new form instance service
func NewFormInstanceService(
	instanceRepo repository.IFormInstanceRepository,
	formRepo repository.IFormBuilderRepository,
) IFormInstanceService {
	return &formInstanceService{
		instanceRepo: instanceRepo,
		formRepo:     formRepo,
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

	// Create form instance
	instance := &db.FormInstance{
		FormID:       id,
		CurrentState: s.determineInitialState(action),
		FieldValues:  fieldValuesJSON,
		CreatedBy:    userID.String(),
		Metadata:     metadataJSON,
	}

	// Save instance
	createdInstance, err := s.instanceRepo.CreateFormInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create form instance: %w", err)
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

func (s *formInstanceService) extractUserAgent(ctx context.Context) string {
	userAgent := ctx.Value("user_agent").(string)
	return userAgent
}
