package services

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
	forminstancev2 "p9e.in/ugcl/formbuilder/api/v2/form_instance"
	workflowv2 "p9e.in/ugcl/formbuilder/api/v2/workflow"

	"github.com/google/uuid"
)

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
