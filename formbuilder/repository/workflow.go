// =============================================================================
// internal/repository/workflow_repository.go
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/utils"

	"github.com/google/uuid"
)

type workflowRepository struct {
	queries *db.Queries
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(queries *db.Queries) IWorkflowRepository {
	return &workflowRepository{
		queries: queries,
	}
}

func (r *workflowRepository) CreateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	if workflow.ID == uuid.Nil {
		workflow.ID = utils.GenerateUUID()
	}

	params := db.CreateWorkflowParams{
		ID:                workflow.ID,
		InitialState:      workflow.InitialState,
		States:            workflow.States,
		GlobalTransitions: workflow.GlobalTransitions,
		Metadata:          workflow.Metadata,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return r.queries.CreateWorkflow(ctx, params)
}

func (r *workflowRepository) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*db.Workflow, error) {
	workflow, err := r.queries.GetWorkflow(ctx, db.GetWorkflowParams{ID: workflowID})
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (r *workflowRepository) GetWorkflowByFormID(ctx context.Context, formID uuid.UUID) (*db.Workflow, error) {
	workflow, err := r.queries.GetWorkflowByFormID(ctx, db.GetWorkflowByFormIDParams{ID: formID})
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	params := db.UpdateWorkflowParams{
		ID:                workflow.ID,
		InitialState:      workflow.InitialState,
		States:            workflow.States,
		GlobalTransitions: workflow.GlobalTransitions,
		Metadata:          workflow.Metadata,
		UpdatedAt:         time.Now(),
	}

	return r.queries.UpdateWorkflow(ctx, params)
}

func (r *workflowRepository) DeleteWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	return r.queries.DeleteWorkflow(ctx, db.DeleteWorkflowParams{ID: workflowID})
}

func (r *workflowRepository) ListWorkflows(ctx context.Context, limit, offset int32) ([]*db.Workflow, error) {
	params := db.ListWorkflowsParams{
		Limit:  limit,
		Offset: offset,
	}

	workflows, err := r.queries.ListWorkflows(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Workflow, len(workflows))
	for i, workflow := range workflows {
		result[i] = workflow
	}
	return result, nil
}

func (r *workflowRepository) CountWorkflows(ctx context.Context) (int64, error) {
	return r.queries.CountWorkflows(ctx)
}

// SLA rule methods
func (r *workflowRepository) CreateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	if rule.ID == uuid.Nil {
		rule.ID = utils.GenerateUUID()
	}

	params := db.CreateSLARuleParams{
		ID:               rule.ID,
		Name:             rule.Name,
		State:            rule.State,
		Duration:         rule.Duration,
		EscalationLevels: rule.EscalationLevels,
		Active:           rule.Active,
		ApplicableRoles:  rule.ApplicableRoles,
		Condition:        rule.Condition,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return r.queries.CreateSLARule(ctx, params)
}

func (r *workflowRepository) GetSLARule(ctx context.Context, ruleID uuid.UUID) (*db.SlaRule, error) {
	rule, err := r.queries.GetSLARule(ctx, db.GetSLARuleParams{ID: ruleID})
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *workflowRepository) GetSLARules(ctx context.Context) ([]*db.SlaRule, error) {
	rules, err := r.queries.GetSLARules(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SlaRule, len(rules))
	for i, rule := range rules {
		result[i] = rule
	}
	return result, nil
}

func (r *workflowRepository) GetSLARulesByState(ctx context.Context, state string) ([]*db.SlaRule, error) {
	rules, err := r.queries.GetSLARulesByState(ctx, db.GetSLARulesByStateParams{State: state})
	if err != nil {
		return nil, err
	}

	result := make([]*db.SlaRule, len(rules))
	for i, rule := range rules {
		result[i] = rule
	}
	return result, nil
}

func (r *workflowRepository) GetActiveSLARules(ctx context.Context) ([]*db.SlaRule, error) {
	rules, err := r.queries.GetActiveSLARules(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SlaRule, len(rules))
	for i, rule := range rules {
		result[i] = rule
	}
	return result, nil
}

func (r *workflowRepository) UpdateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	params := db.UpdateSLARuleParams{
		ID:               rule.ID,
		Name:             rule.Name,
		State:            rule.State,
		Duration:         rule.Duration,
		EscalationLevels: rule.EscalationLevels,
		Active:           rule.Active,
		ApplicableRoles:  rule.ApplicableRoles,
		Condition:        rule.Condition,
		UpdatedAt:        time.Now(),
	}

	return r.queries.UpdateSLARule(ctx, params)
}

func (r *workflowRepository) DeleteSLARule(ctx context.Context, ruleID uuid.UUID) error {
	return r.queries.DeleteSLARule(ctx, db.DeleteSLARuleParams{ID: ruleID})
}

func (r *workflowRepository) ToggleSLARule(ctx context.Context, ruleID uuid.UUID) (*db.SlaRule, error) {
	return r.queries.ToggleSLARule(ctx, db.ToggleSLARuleParams{ID: ruleID})
}

// Escalation methods
func (r *workflowRepository) CreateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	if escalation.ID == uuid.Nil {
		escalation.ID = utils.GenerateUUID()
	}

	params := db.CreateEscalationParams{
		ID:                escalation.ID,
		Name:              escalation.Name,
		TriggerCondition:  escalation.TriggerCondition,
		FromStates:        escalation.FromStates,
		ToState:           escalation.ToState,
		NotifyRoles:       escalation.NotifyRoles,
		EscalationMessage: escalation.EscalationMessage,
		AutoEscalate:      escalation.AutoEscalate,
		AfterDuration:     escalation.AfterDuration,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return r.queries.CreateEscalation(ctx, params)
}

func (r *workflowRepository) GetEscalation(ctx context.Context, escalationID uuid.UUID) (*db.Escalation, error) {
	escalation, err := r.queries.GetEscalation(ctx, db.GetEscalationParams{ID: escalationID})
	if err != nil {
		return nil, err
	}
	return escalation, nil
}

func (r *workflowRepository) GetEscalations(ctx context.Context) ([]*db.Escalation, error) {
	escalations, err := r.queries.GetEscalations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Escalation, len(escalations))
	copy(result, escalations)
	return result, nil
}

func (r *workflowRepository) GetEscalationsByFromState(ctx context.Context, fromState string) ([]*db.Escalation, error) {
	// Create JSONB parameter for the query
	jsonbState := []byte(`"` + fromState + `"`)

	escalations, err := r.queries.GetEscalationsByFromState(ctx, db.GetEscalationsByFromStateParams{Column1: jsonbState})
	if err != nil {
		return nil, err
	}

	result := make([]*db.Escalation, len(escalations))
	copy(result, escalations)
	return result, nil
}

func (r *workflowRepository) GetAutoEscalations(ctx context.Context) ([]*db.Escalation, error) {
	escalations, err := r.queries.GetAutoEscalations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Escalation, len(escalations))
	copy(result, escalations)
	return result, nil
}

func (r *workflowRepository) UpdateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	params := db.UpdateEscalationParams{
		ID:                escalation.ID,
		Name:              escalation.Name,
		TriggerCondition:  escalation.TriggerCondition,
		FromStates:        escalation.FromStates,
		ToState:           escalation.ToState,
		NotifyRoles:       escalation.NotifyRoles,
		EscalationMessage: escalation.EscalationMessage,
		AutoEscalate:      escalation.AutoEscalate,
		AfterDuration:     escalation.AfterDuration,
		UpdatedAt:         time.Now(),
	}

	return r.queries.UpdateEscalation(ctx, params)
}

func (r *workflowRepository) DeleteEscalation(ctx context.Context, escalationID uuid.UUID) error {
	return r.queries.DeleteEscalation(ctx, db.DeleteEscalationParams{ID: escalationID})
}

func (r *workflowRepository) ToggleAutoEscalation(ctx context.Context, escalationID uuid.UUID) (*db.Escalation, error) {
	return r.queries.ToggleAutoEscalation(ctx, db.ToggleAutoEscalationParams{ID: escalationID})
}
