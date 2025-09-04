package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/formbuilder/api/v2/workflow"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// ProtoToWorkflow converts protobuf Workflow to SQLC Workflow struct
func ProtoToWorkflow(proto *pb.Workflow) (*db.Workflow, error) {
	if proto == nil {
		return nil, fmt.Errorf("invalid workflow")
	}

	// Convert protobuf states to Go structs with proper field mapping
	goStates := make([]map[string]interface{}, len(proto.States))
	for i, protoState := range proto.States {
		// Convert transitions
		transitions := make([]map[string]interface{}, len(protoState.Transitions))
		for j, protoTransition := range protoState.Transitions {
			// Convert actions
			actions := make([]map[string]interface{}, len(protoTransition.Actions))
			for k, protoAction := range protoTransition.Actions {
				actions[k] = map[string]interface{}{
					"type":   protoAction.Type,
					"params": protoAction.Params,
				}
			}
			
			transitions[j] = map[string]interface{}{
				"event":      protoTransition.Event,
				"condition":  protoTransition.Condition,
				"next_state": protoTransition.NextState,
				"actions":    actions,
				"metadata":   protoTransition.Metadata,
			}
		}
		
		goStates[i] = map[string]interface{}{
			"id":             protoState.Id,
			"label":          protoState.Label,
			"type":           int(protoState.Type), // Convert enum to int
			"assigned_role":  protoState.AssignedRole,
			"assigned_users": protoState.AssignedUsers,
			"actions":        protoState.Actions,
			"transitions":    transitions,
			"properties":     protoState.Properties,
		}
	}

	// Marshal the converted states
	statesJSON, err := json.Marshal(goStates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal states: %w", err)
	}

	globalTransitionsJSON, err := json.Marshal(proto.GlobalTransitions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal global transitions: %w", err)
	}

	metadataJSON, err := json.Marshal(proto.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return &db.Workflow{
		ID:                uuid.New(),
		InitialState:      proto.InitialState,
		States:            statesJSON,
		GlobalTransitions: globalTransitionsJSON,
		Metadata:          metadataJSON,
	}, nil
}

// WorkflowToProto converts SQLC Workflow struct to protobuf Workflow
func WorkflowToProto(workflow *db.Workflow) (*pb.Workflow, error) {
	if workflow == nil {
		return nil, fmt.Errorf("invalid workflow model")
	}

	// Unmarshal JSONB fields
	var states []*pb.WorkflowState
	if err := json.Unmarshal(workflow.States, &states); err != nil {
		return nil, fmt.Errorf("failed to unmarshal states: %w", err)
	}

	var globalTransitions []*pb.WorkflowTransition
	if err := json.Unmarshal(workflow.GlobalTransitions, &globalTransitions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global transitions: %w", err)
	}

	var metadata map[string]string
	if err := json.Unmarshal(workflow.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &pb.Workflow{
		InitialState:      workflow.InitialState,
		States:            states,
		GlobalTransitions: globalTransitions,
		Metadata:          metadata,
	}, nil
}

// WorkflowStateToProto converts a workflow state to protobuf WorkflowState
func WorkflowStateToProto(state interface{}) (*pb.WorkflowState, error) {
	// Convert interface{} to JSON then to protobuf
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	var protoState pb.WorkflowState
	if err := json.Unmarshal(stateJSON, &protoState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to proto state: %w", err)
	}

	return &protoState, nil
}

// SLARuleToProto converts SQLC SlaRule struct to protobuf SLARule
func SLARuleToProto(rule *db.SlaRule) (*pb.SLARule, error) {
	if rule == nil {
		return nil, fmt.Errorf("invalid SLA rule model")
	}

	// Unmarshal JSONB fields
	var duration pb.Duration
	if err := json.Unmarshal(rule.Duration, &duration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal duration: %w", err)
	}

	var escalationLevels []*pb.EscalationLevel
	if err := json.Unmarshal(rule.EscalationLevels, &escalationLevels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal escalation levels: %w", err)
	}

	var applicableRoles []string
	if err := json.Unmarshal(rule.ApplicableRoles, &applicableRoles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal applicable roles: %w", err)
	}

	return &pb.SLARule{
		Id:               rule.ID.String(),
		Name:             rule.Name,
		State:            rule.State,
		Duration:         &duration,
		EscalationLevels: escalationLevels,
		Active:           boolValue(rule.Active),
		ApplicableRoles:  applicableRoles,
		Condition:        stringValue(rule.Condition),
	}, nil
}

// EscalationToProto converts SQLC Escalation struct to protobuf Escalation
func EscalationToProto(escalation *db.Escalation) (*pb.Escalation, error) {
	if escalation == nil {
		return nil, fmt.Errorf("invalid escalation model")
	}

	// Unmarshal JSONB fields
	var fromStates []string
	if err := json.Unmarshal(escalation.FromStates, &fromStates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal from states: %w", err)
	}

	var notifyRoles []string
	if err := json.Unmarshal(escalation.NotifyRoles, &notifyRoles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notify roles: %w", err)
	}

	var afterDuration pb.Duration
	if err := json.Unmarshal(escalation.AfterDuration, &afterDuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal after duration: %w", err)
	}

	return &pb.Escalation{
		Id:                escalation.ID.String(),
		Name:              escalation.Name,
		TriggerCondition:  stringValue(escalation.TriggerCondition),
		FromStates:        fromStates,
		ToState:           escalation.ToState,
		NotifyRoles:       notifyRoles,
		EscalationMessage: stringValue(escalation.EscalationMessage),
		AutoEscalate:      boolValue(escalation.AutoEscalate),
		AfterDuration:     &afterDuration,
	}, nil
}

// ProtoToSLARule converts protobuf SLARule to SQLC SlaRule struct
func ProtoToSLARule(proto *pb.SLARule) (*db.SlaRule, error) {
	if proto == nil {
		return nil, fmt.Errorf("invalid SLA rule")
	}

	// Marshal JSONB fields
	durationJSON, err := json.Marshal(proto.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal duration: %w", err)
	}

	escalationLevelsJSON, err := json.Marshal(proto.EscalationLevels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal escalation levels: %w", err)
	}

	applicableRolesJSON, err := json.Marshal(proto.ApplicableRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal applicable roles: %w", err)
	}

	return &db.SlaRule{
		ID:               uuid.New(),
		Name:             proto.Name,
		State:            proto.State,
		Duration:         durationJSON,
		EscalationLevels: escalationLevelsJSON,
		Active:           &proto.Active,
		ApplicableRoles:  applicableRolesJSON,
		Condition:        &proto.Condition,
	}, nil
}
