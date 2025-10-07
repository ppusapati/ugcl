package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/repository"
)

type approvalService struct {
	formRepo     repository.IFormBuilderRepository
	instanceRepo repository.IFormInstanceRepository
	workflowRepo repository.IWorkflowRepository
}

// NewApprovalService creates a new approval service
func NewApprovalService(
	formRepo repository.IFormBuilderRepository,
	instanceRepo repository.IFormInstanceRepository,
	workflowRepo repository.IWorkflowRepository,
) IApprovalService {
	return &approvalService{
		formRepo:     formRepo,
		instanceRepo: instanceRepo,
		workflowRepo: workflowRepo,
	}
}

// StartApprovalProcess initiates a new approval workflow for a form instance
func (s *approvalService) StartApprovalProcess(ctx context.Context, req *StartApprovalRequest) (*db.FormInstance, error) {
	// Create form instance with approval workflow state
	instanceID := uuid.New()

	// Convert request data to JSON
	requestDataJson, err := json.Marshal(req.RequestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Convert metadata to JSON
	metadataJson, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create the form instance
	instance := &db.FormInstance{
		ID:           instanceID,
		FormID:       uuid.MustParse(req.EntityID), // Assuming EntityID is the form ID
		CurrentState: "pending_approval",
		FieldValues:  requestDataJson,
		CreatedBy:    req.RequestedBy,
		AssignedTo:   nil, // Will be assigned based on workflow rules
		Metadata:     metadataJson,
	}

	createdInstance, err := s.instanceRepo.CreateFormInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create form instance: %w", err)
	}

	// TODO: Determine approver based on workflow rules and assign
	// This would involve looking up the form's workflow configuration
	// and determining the next approver in the approval chain

	return createdInstance, nil
}

// ProcessApprovalAction processes an approval action (approve, reject, delegate, etc.)
func (s *approvalService) ProcessApprovalAction(ctx context.Context, req *ProcessApprovalActionRequest) (*db.ApprovalAction, *db.FormInstance, error) {
	// Get the current instance
	instance, err := s.instanceRepo.GetFormInstance(ctx, req.InstanceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get form instance: %w", err)
	}

	// Create the approval action record
	actionReq := &CreateApprovalActionRequest{
		FormInstanceID: req.InstanceID,
		ApproverID:     req.ApproverID,
		Action:         req.Action,
		Comments:       req.Comments,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		DelegatedFrom:  req.DelegatedFrom,
		AttachmentURLs: req.AttachmentURLs,
		Metadata:       req.Metadata,
	}

	action, err := s.CreateApprovalAction(ctx, actionReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create approval action: %w", err)
	}

	// Update instance state based on action
	var newState string
	switch req.Action {
	case "APPROVE":
		// TODO: Check if this is the final approval step
		newState = "approved"
	case "REJECT":
		newState = "rejected"
	case "DELEGATE":
		// TODO: Handle delegation logic
		if req.DelegateTo != nil {
			instance.AssignedTo = req.DelegateTo
		}
		newState = instance.CurrentState // Keep current state but change assignee
	case "REQUEST_INFO":
		newState = "info_requested"
	case "WITHDRAW":
		newState = "withdrawn"
	case "REASSIGN":
		if req.DelegateTo != nil {
			instance.AssignedTo = req.DelegateTo
		}
		newState = instance.CurrentState
	default:
		return nil, nil, fmt.Errorf("unsupported action: %s", req.Action)
	}

	// Update instance state
	instance.CurrentState = newState
	updatedInstance, err := s.instanceRepo.UpdateFormInstance(ctx, instance)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update form instance: %w", err)
	}

	return action, updatedInstance, nil
}

// CreateApprovalAction creates a new approval action record
func (s *approvalService) CreateApprovalAction(ctx context.Context, req *CreateApprovalActionRequest) (*db.ApprovalAction, error) {
	// Convert attachment URLs to JSON
	attachmentUrlsJson, err := json.Marshal(req.AttachmentURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attachment URLs: %w", err)
	}

	// Convert metadata to JSON
	metadataJson, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create approval action
	action := &db.ApprovalAction{
		ID:             uuid.New(),
		FormInstanceID: pgtype.UUID{Bytes: req.FormInstanceID, Valid: true},
		ApproverID:     req.ApproverID,
		Action:         req.Action,
		Comments:       &req.Comments,
		ActedAt:        time.Now(),
		UserAgent:      &req.UserAgent,
		DelegatedFrom:  req.DelegatedFrom,
		AttachmentUrls: attachmentUrlsJson,
		Metadata:       metadataJson,
		CreatedAt:      time.Now(),
	}

	// TODO: Implement actual database creation via repository
	// For now, return the action (would need to add this to repository interface)
	return action, nil
}

// GetApprovalActionsByInstance gets all approval actions for a form instance
func (s *approvalService) GetApprovalActionsByInstance(ctx context.Context, instanceID uuid.UUID) ([]*db.ApprovalAction, error) {
	// TODO: Implement via repository
	return []*db.ApprovalAction{}, nil
}

// GetApprovalActionsByApprover gets approval actions by approver with pagination
func (s *approvalService) GetApprovalActionsByApprover(ctx context.Context, approverID string, limit, offset int32) ([]*db.ApprovalAction, error) {
	// TODO: Implement via repository
	return []*db.ApprovalAction{}, nil
}

// CreateDelegate creates a new delegation relationship
func (s *approvalService) CreateDelegate(ctx context.Context, req *CreateDelegateRequest) (*db.ApprovalDelegate, error) {
	// Convert entity types to JSON
	entityTypesJson, err := json.Marshal(req.EntityTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity types: %w", err)
	}

	// Convert metadata to JSON
	metadataJson, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	delegate := &db.ApprovalDelegate{
		ID:          uuid.New(),
		DelegatorID: req.DelegatorID,
		DelegateID:  req.DelegateID,
		EntityTypes: entityTypesJson,
		StartDate:   req.StartDate,
		EndDate:     func() sql.NullTime { if req.EndDate != nil { return sql.NullTime{Time: *req.EndDate, Valid: true} }; return sql.NullTime{Valid: false} }(),
		IsActive:    &[]bool{true}[0],
		Reason:      &req.Reason,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    metadataJson,
	}

	// TODO: Implement actual database creation via repository
	return delegate, nil
}

// UpdateDelegate updates an existing delegation
func (s *approvalService) UpdateDelegate(ctx context.Context, delegateID uuid.UUID, req *UpdateDelegateRequest) (*db.ApprovalDelegate, error) {
	// TODO: Implement via repository
	return nil, fmt.Errorf("not implemented")
}

// DeleteDelegate deletes a delegation (sets inactive)
func (s *approvalService) DeleteDelegate(ctx context.Context, delegateID uuid.UUID) error {
	// TODO: Implement via repository
	return fmt.Errorf("not implemented")
}

// GetDelegatesByDelegator gets delegations by delegator
func (s *approvalService) GetDelegatesByDelegator(ctx context.Context, delegatorID string) ([]*db.ApprovalDelegate, error) {
	// TODO: Implement via repository
	return []*db.ApprovalDelegate{}, nil
}

// GetDelegatesByDelegate gets delegations by delegate
func (s *approvalService) GetDelegatesByDelegate(ctx context.Context, delegateID string) ([]*db.ApprovalDelegate, error) {
	// TODO: Implement via repository
	return []*db.ApprovalDelegate{}, nil
}

// GetActiveDelegation gets active delegation for a delegator and entity type
func (s *approvalService) GetActiveDelegation(ctx context.Context, delegatorID, entityType string) (*db.ApprovalDelegate, error) {
	// TODO: Implement via repository
	return nil, nil
}

// GetPendingApprovalsForUser gets pending approvals for a user
func (s *approvalService) GetPendingApprovalsForUser(ctx context.Context, userID string, limit, offset int32) ([]*PendingApprovalWithForm, error) {
	// TODO: Implement via repository using the SQL query we defined
	return []*PendingApprovalWithForm{}, nil
}

// GetApprovalHistory gets the approval history for an instance
func (s *approvalService) GetApprovalHistory(ctx context.Context, instanceID uuid.UUID) ([]*ApprovalHistoryEntry, error) {
	// TODO: Implement via repository using the SQL query we defined
	return []*ApprovalHistoryEntry{}, nil
}

// GetApprovalMetrics gets approval metrics
func (s *approvalService) GetApprovalMetrics(ctx context.Context, req *ApprovalMetricsRequest) (*ApprovalMetrics, error) {
	// TODO: Implement via repository using the SQL query we defined
	return &ApprovalMetrics{}, nil
}

// CreateApprovalReport creates a new approval report
func (s *approvalService) CreateApprovalReport(ctx context.Context, req *CreateApprovalReportRequest) (*db.ApprovalReport, error) {
	// TODO: Implement via repository
	return nil, fmt.Errorf("not implemented")
}

// GetApprovalReportsByWorkflow gets approval reports by workflow
func (s *approvalService) GetApprovalReportsByWorkflow(ctx context.Context, workflowID uuid.UUID, startDate, endDate *time.Time) ([]*db.ApprovalReport, error) {
	// TODO: Implement via repository
	return []*db.ApprovalReport{}, nil
}

// GetApprovalReportsByEntityType gets approval reports by entity type
func (s *approvalService) GetApprovalReportsByEntityType(ctx context.Context, entityType string, startDate, endDate *time.Time) ([]*db.ApprovalReport, error) {
	// TODO: Implement via repository
	return []*db.ApprovalReport{}, nil
}

// GetEscalatedInstances gets escalated instances
func (s *approvalService) GetEscalatedInstances(ctx context.Context, limit, offset int32) ([]*FormInstanceWithTitle, error) {
	// TODO: Implement via repository using the SQL query we defined
	return []*FormInstanceWithTitle{}, nil
}

// ProcessEscalations processes escalations (background job)
func (s *approvalService) ProcessEscalations(ctx context.Context) (int32, []string, error) {
	// TODO: Implement escalation processing logic
	// This would:
	// 1. Find instances that have exceeded their SLA
	// 2. Escalate them according to escalation rules
	// 3. Send notifications
	// 4. Return count of processed instances and list of escalated instance IDs
	return 0, []string{}, nil
}