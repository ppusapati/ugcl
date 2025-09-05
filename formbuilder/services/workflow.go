package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/repository"
)

type workflowService struct {
	workflowRepo repository.IWorkflowRepository
	formRepo     repository.IFormBuilderRepository
	instanceRepo repository.IFormInstanceRepository
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(
	workflowRepo repository.IWorkflowRepository,
	formRepo repository.IFormBuilderRepository,
	instanceRepo repository.IFormInstanceRepository,
) IWorkflowService {
	return &workflowService{
		workflowRepo: workflowRepo,
		formRepo:     formRepo,
		instanceRepo: instanceRepo,
	}
}

// CreateEscalation implements IWorkflowService.
func (w *workflowService) CreateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	return w.workflowRepo.CreateEscalation(ctx, escalation)
}

// CreateSLARule implements IWorkflowService.
func (w *workflowService) CreateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	return w.workflowRepo.CreateSLARule(ctx, rule)
}

// CreateWorkflow implements IWorkflowService.
func (w *workflowService) CreateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	return w.workflowRepo.CreateWorkflow(ctx, workflow)
}

// DeleteEscalation implements IWorkflowService.
func (w *workflowService) DeleteEscalation(ctx context.Context, escalationID string) error {
	id, err := uuid.Parse(escalationID)
	if err != nil {
		return fmt.Errorf("invalid escalation ID: %w", err)
	}
	return w.workflowRepo.DeleteEscalation(ctx, id)
}

// DeleteSLARule implements IWorkflowService.
func (w *workflowService) DeleteSLARule(ctx context.Context, ruleID string) error {
	id, err := uuid.Parse(ruleID)
	if err != nil {
		return fmt.Errorf("invalid SLA rule ID: %w", err)
	}
	return w.workflowRepo.DeleteSLARule(ctx, id)
}

// DeleteWorkflow implements IWorkflowService.
func (w *workflowService) DeleteWorkflow(ctx context.Context, workflowID string) error {
	id, err := uuid.Parse(workflowID)
	if err != nil {
		return fmt.Errorf("invalid workflow ID: %w", err)
	}
	return w.workflowRepo.DeleteWorkflow(ctx, id)
}

// GetEscalations implements IWorkflowService.
func (w *workflowService) GetEscalations(ctx context.Context, fromState string) ([]*db.Escalation, error) {
	return w.workflowRepo.GetEscalationsByFromState(ctx, fromState)
}

// GetSLARules implements IWorkflowService.
func (w *workflowService) GetSLARules(ctx context.Context, state string) ([]*db.SlaRule, error) {
	if state == "" {
		return w.workflowRepo.GetSLARules(ctx)
	}
	return w.workflowRepo.GetSLARulesByState(ctx, state)
}

// GetWorkflow implements IWorkflowService.
func (w *workflowService) GetWorkflow(ctx context.Context, workflowID string) (*db.Workflow, error) {
	id, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID: %w", err)
	}
	return w.workflowRepo.GetWorkflow(ctx, id)
}

// GetWorkflowStates implements IWorkflowService.
func (w *workflowService) GetWorkflowStates(ctx context.Context, formID string) (*GetWorkflowStatesResult, error) {
	id, err := uuid.Parse(formID)
	if err != nil {
		return nil, fmt.Errorf("invalid form ID: %w", err)
	}

	// Get workflow associated with the form
	workflow, err := w.workflowRepo.GetWorkflowByFormID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow for form: %w", err)
	}

	// Parse workflow states from JSON
	var states []WorkflowState
	if err := json.Unmarshal(workflow.States, &states); err != nil {
		return nil, fmt.Errorf("failed to parse workflow states: %w", err)
	}

	return &GetWorkflowStatesResult{
		States:           states,
		CurrentState:     workflow.InitialState,
		AvailableActions: []string{}, // TODO: Calculate based on current state
	}, nil
}

// TransitionWorkflow implements IWorkflowService.
func (w *workflowService) TransitionWorkflow(ctx context.Context, instanceID string, event string, userID uuid.UUID, context map[string]string) (*TransitionResult, error) {
	// Parse instance ID
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, fmt.Errorf("invalid instance ID: %w", err)
	}

	// Get form instance to determine current state and workflow
	instance, err := w.instanceRepo.GetFormInstance(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get form instance: %w", err)
	}

	// Get form to find associated workflow
	form, err := w.formRepo.GetForm(ctx, instance.FormID)
	if err != nil {
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if !form.WorkflowID.Valid {
		return &TransitionResult{
			Success: false,
			Message: "No workflow associated with this form",
		}, nil
	}

	// Get workflow definition
	workflow, err := w.workflowRepo.GetWorkflow(ctx, form.WorkflowID.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Parse workflow states
	var states []WorkflowState
	if err := json.Unmarshal(workflow.States, &states); err != nil {
		return nil, fmt.Errorf("failed to parse workflow states: %w", err)
	}

	// Find current state and determine next state based on event
	currentState := instance.CurrentState
	nextState, actions := w.calculateNextState(states, currentState, event)

	if nextState == "" {
		// Debug: Log available states and transitions
		var availableStates []string
		var availableTransitions []string
		for _, state := range states {
			availableStates = append(availableStates, state.ID)
			if state.ID == currentState {
				for _, transition := range state.Transitions {
					availableTransitions = append(availableTransitions, transition.Event)
				}
			}
		}
		
		debugMsg := fmt.Sprintf("Invalid transition from state '%s' with event '%s'. Available states: %v. Available events from current state: %v", 
			currentState, event, availableStates, availableTransitions)
		
		return &TransitionResult{
			Success: false,
			Message: debugMsg,
		}, nil
	}

	// Update instance state
	instance.CurrentState = nextState
	_, err = w.instanceRepo.UpdateFormInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance state: %w", err)
	}

	// Start SLA tracking for the new state
	err = w.startSLATracking(ctx, id, nextState, userID)
	if err != nil {
		// Log error but don't fail the transition
		fmt.Printf("WARNING: Failed to start SLA tracking: %v\n", err)
	}

	// Create audit log entry for the transition
	fromState := currentState
	toState := nextState
	auditLog := &db.AuditLog{
		ID:         uuid.New(),
		InstanceID: id,
		UserID:     userID.String(),
		Action:     event,
		FromState:  &fromState,
		ToState:    &toState,
		Changes:    []byte("{}"), // Empty changes for workflow transitions
		Timestamp:  time.Now(),
		IpAddress:  nil, // Could be passed from context if needed
		UserAgent:  nil, // Could be passed from context if needed
	}

	_, err = w.instanceRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the transition
		fmt.Printf("WARNING: Failed to create audit log for transition: %v\n", err)
	}

	// Create comment if provided in context
	if comment, exists := context["comment"]; exists && comment != "" {
		isInternal := false
		if internalStr, exists := context["internal"]; exists && internalStr == "true" {
			isInternal = true
		}

		commentObj := &db.Comment{
			ID:         uuid.New(),
			InstanceID: id,
			UserID:     userID.String(),
			Text:       comment,
			CreatedAt:  time.Now(),
			Internal:   &isInternal,
		}

		_, err = w.instanceRepo.CreateComment(ctx, commentObj)
		if err != nil {
			// Log error but don't fail the transition
			fmt.Printf("WARNING: Failed to create comment for transition: %v\n", err)
		}
	}

	return &TransitionResult{
		Success:         true,
		NewState:        nextState,
		Message:         fmt.Sprintf("Transitioned from '%s' to '%s'", currentState, nextState),
		ExecutedActions: actions,
	}, nil
}

// calculateNextState determines the next state based on current state and event
func (w *workflowService) calculateNextState(states []WorkflowState, currentState, event string) (string, []*TransitionAction) {
	// Find current state definition
	for _, state := range states {
		if state.ID == currentState {
			// Debug: log the transitions for this state
			fmt.Printf("DEBUG: Found state '%s' with %d transitions\n", state.ID, len(state.Transitions))
			
			// Check transitions for matching event
			for i, transition := range state.Transitions {
				fmt.Printf("DEBUG: Transition %d - Event: '%s', NextState: '%s'\n", i, transition.Event, transition.NextState)
				
				if transition.Event == event {
					fmt.Printf("DEBUG: Match found! Event '%s' -> NextState '%s'\n", event, transition.NextState)
					
					// Convert transition actions
					actions := make([]*TransitionAction, len(transition.Actions))
					for i, action := range transition.Actions {
						actions[i] = &TransitionAction{
							Type:   action.Type,
							Params: action.Params,
						}
					}
					return transition.NextState, actions
				}
			}
			
			fmt.Printf("DEBUG: No matching transition found for event '%s' in state '%s'\n", event, currentState)
			break
		}
	}
	return "", nil
}

// startSLATracking initiates SLA tracking for a form instance in a specific state
func (w *workflowService) startSLATracking(ctx context.Context, instanceID uuid.UUID, state string, userID uuid.UUID) error {
	// Get SLA rules for this state
	slaRules, err := w.GetSLARules(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get SLA rules for state %s: %w", state, err)
	}

	// No SLA rules for this state
	if len(slaRules) == 0 {
		return nil
	}

	now := time.Now()
	
	for _, rule := range slaRules {
		if rule.Active == nil || !*rule.Active {
			continue
		}

		// Parse duration from JSON
		var duration struct {
			Value int32  `json:"value"`
			Unit  string `json:"unit"`
		}
		if err := json.Unmarshal(rule.Duration, &duration); err != nil {
			fmt.Printf("WARNING: Failed to parse duration for SLA rule %s: %v\n", rule.ID, err)
			continue
		}

		// Calculate due time based on rule duration
		var dueTime time.Time
		switch duration.Unit {
		case "MINUTES":
			dueTime = now.Add(time.Duration(duration.Value) * time.Minute)
		case "HOURS":
			dueTime = now.Add(time.Duration(duration.Value) * time.Hour)
		case "DAYS":
			dueTime = now.Add(time.Duration(duration.Value) * 24 * time.Hour)
		case "WEEKS":
			dueTime = now.Add(time.Duration(duration.Value) * 7 * 24 * time.Hour)
		case "MONTHS":
			dueTime = now.AddDate(0, int(duration.Value), 0)
		default:
			dueTime = now.Add(time.Duration(duration.Value) * time.Hour) // Default to hours
		}

		// Create SLA instance record
		slaInstanceParams := db.CreateSLAInstanceParams{
			ID:         uuid.New(),
			InstanceID: instanceID,
			SlaRuleID:  rule.ID,
			State:      state,
			StartTime:  now,
			DueTime:    dueTime,
			Status:     "active",
			AssignedTo: nil, // Could be set based on workflow state assignment
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		_, err = w.workflowRepo.CreateSLAInstance(ctx, slaInstanceParams)
		if err != nil {
			fmt.Printf("WARNING: Failed to create SLA instance: %v\n", err)
			continue
		}
		
		fmt.Printf("SLA tracking created: Instance=%s, Rule=%s, Due=%s\n", 
			instanceID, rule.ID, dueTime.Format(time.RFC3339))
	}

	return nil
}

// TriggerExternalWorkflow implements IWorkflowService.
func (w *workflowService) TriggerExternalWorkflow(ctx context.Context, instanceID string, action string, variables map[string]interface{}) (*ExternalWorkflowResult, error) {
	// TODO: Implement external workflow integration (BPMN engines like Camunda, Flowable, Zeebe)
	return &ExternalWorkflowResult{
		ProcessInstanceID: fmt.Sprintf("ext-%s-%s", instanceID, action),
		Success:           true,
		Message:           "External workflow triggered successfully",
	}, nil
}

// UpdateEscalation implements IWorkflowService.
func (w *workflowService) UpdateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error) {
	return w.workflowRepo.UpdateEscalation(ctx, escalation)
}

// UpdateSLARule implements IWorkflowService.
func (w *workflowService) UpdateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error) {
	return w.workflowRepo.UpdateSLARule(ctx, rule)
}

// UpdateWorkflow implements IWorkflowService.
func (w *workflowService) UpdateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error) {
	return w.workflowRepo.UpdateWorkflow(ctx, workflow)
}
