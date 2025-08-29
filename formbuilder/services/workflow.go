package services

import (
	"context"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/repository"

	"github.com/google/uuid"
)

type workflowService struct {
	workflowRepo repository.IWorkflowRepository
	formRepo     repository.IFormBuilderRepository
}

// NewWorkflowService creates a new form instance service
func NewWorkflowService(
	workflowRepo repository.IWorkflowRepository,
	formRepo repository.IFormBuilderRepository,
) IWorkflowService {
	return &workflowService{
		workflowRepo: workflowRepo,
		formRepo:     formRepo,
	}
}

// CreateEscalation implements IWorkflowService.
func (w *workflowService) CreateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	panic("unimplemented")
}

// CreateSLARule implements IWorkflowService.
func (w *workflowService) CreateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	panic("unimplemented")
}

// CreateWorkflow implements IWorkflowService.
func (w *workflowService) CreateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	panic("unimplemented")
}

// DeleteEscalation implements IWorkflowService.
func (w *workflowService) DeleteEscalation(ctx context.Context, escalationID string) error {
	panic("unimplemented")
}

// DeleteSLARule implements IWorkflowService.
func (w *workflowService) DeleteSLARule(ctx context.Context, ruleID string) error {
	panic("unimplemented")
}

// DeleteWorkflow implements IWorkflowService.
func (w *workflowService) DeleteWorkflow(ctx context.Context, workflowID string) error {
	panic("unimplemented")
}

// GetEscalations implements IWorkflowService.
func (w *workflowService) GetEscalations(ctx context.Context, fromState string) ([]*db.Escalation, error) {
	panic("unimplemented")
}

// GetSLARules implements IWorkflowService.
func (w *workflowService) GetSLARules(ctx context.Context, state string) ([]*db.SlaRule, error) {
	panic("unimplemented")
}

// GetWorkflow implements IWorkflowService.
func (w *workflowService) GetWorkflow(ctx context.Context, workflowID string) (*db.Workflow, error) {
	panic("unimplemented")
}

// GetWorkflowStates implements IWorkflowService.
func (w *workflowService) GetWorkflowStates(ctx context.Context, formID string) (*GetWorkflowStatesResult, error) {
	panic("unimplemented")
}

// TransitionWorkflow implements IWorkflowService.
func (w *workflowService) TransitionWorkflow(ctx context.Context, instanceID string, event string, userID uuid.UUID, context map[string]string) (*TransitionResult, error) {
	panic("unimplemented")
}

// TriggerExternalWorkflow implements IWorkflowService.
func (w *workflowService) TriggerExternalWorkflow(ctx context.Context, instanceID string, action string, variables map[string]interface{}) (*ExternalWorkflowResult, error) {
	panic("unimplemented")
}

// UpdateEscalation implements IWorkflowService.
func (w *workflowService) UpdateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	panic("unimplemented")
}

// UpdateSLARule implements IWorkflowService.
func (w *workflowService) UpdateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	panic("unimplemented")
}

// UpdateWorkflow implements IWorkflowService.
func (w *workflowService) UpdateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	panic("unimplemented")
}
