package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"p9e.in/ugcl/approvalworkflow/models"
	"p9e.in/ugcl/approvalworkflow/repository"
)

// ApprovalEngine handles the core approval workflow logic
type ApprovalEngine struct {
	repo         repository.ApprovalRepository
	notifyService NotificationService
	logger       *zap.Logger
}

// NotificationService interface for sending notifications
type NotificationService interface {
	SendApprovalNotification(ctx context.Context, notification *models.ApprovalNotification) error
}

func NewApprovalEngine(repo repository.ApprovalRepository, notifyService NotificationService, logger *zap.Logger) *ApprovalEngine {
	return &ApprovalEngine{
		repo:         repo,
		notifyService: notifyService,
		logger:       logger,
	}
}

// StartApprovalProcess initiates a new approval workflow
func (e *ApprovalEngine) StartApprovalProcess(ctx context.Context, req *StartApprovalRequest) (*models.ApprovalInstance, error) {
	// Get the workflow definition
	workflow, err := e.repo.GetWorkflowByEntityType(ctx, req.EntityType, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow for entity type %s: %w", req.EntityType, err)
	}

	// Parse workflow steps
	var steps []models.ApprovalStep
	if err := json.Unmarshal(workflow.Steps, &steps); err != nil {
		return nil, fmt.Errorf("failed to parse workflow steps: %w", err)
	}

	// Determine initial step based on conditional logic
	initialStep := e.determineInitialStep(steps, req.RequestData)

	// Calculate SLA end time
	slaEndTime := e.calculateSLAEndTime(workflow, req.Priority)

	// Create approval instance
	instance := &models.ApprovalInstance{
		ID:              uuid.New(),
		WorkflowID:      workflow.ID,
		EntityType:      req.EntityType,
		EntityID:        req.EntityID,
		RequestedBy:     req.RequestedBy,
		CurrentStep:     initialStep,
		Status:          models.StatusPending,
		Priority:        req.Priority,
		RequestData:     req.RequestData,
		Comments:        req.Comments,
		AttachmentURLs:  req.AttachmentURLs,
		SLAStartTime:    time.Now(),
		SLAEndTime:      slaEndTime,
		Metadata:        req.Metadata,
	}

	// Save to database
	if err := e.repo.CreateApprovalInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create approval instance: %w", err)
	}

	// Send initial notifications
	if err := e.sendStepNotifications(ctx, instance, initialStep, workflow); err != nil {
		e.logger.Error("Failed to send initial notifications", zap.Error(err), zap.String("instance_id", instance.ID.String()))
	}

	e.logger.Info("Started approval process",
		zap.String("instance_id", instance.ID.String()),
		zap.String("entity_type", req.EntityType),
		zap.String("entity_id", req.EntityID))

	return instance, nil
}

// ProcessApprovalAction handles approval actions (approve, reject, delegate, etc.)
func (e *ApprovalEngine) ProcessApprovalAction(ctx context.Context, req *ApprovalActionRequest) error {
	// Get current approval instance
	instance, err := e.repo.GetApprovalInstance(ctx, req.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get approval instance: %w", err)
	}

	// Validate action is allowed
	if err := e.validateApprovalAction(ctx, instance, req); err != nil {
		return fmt.Errorf("invalid approval action: %w", err)
	}

	// Record the action
	action := &models.ApprovalAction{
		ID:             uuid.New(),
		InstanceID:     req.InstanceID,
		StepID:         instance.CurrentStep,
		ApproverID:     req.ApproverID,
		Action:         req.Action,
		Comments:       req.Comments,
		ActedAt:        time.Now(),
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		DelegatedFrom:  req.DelegatedFrom,
		AttachmentURLs: req.AttachmentURLs,
		Metadata:       req.Metadata,
	}

	if err := e.repo.CreateApprovalAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create approval action: %w", err)
	}

	// Process the action based on type
	switch req.Action {
	case models.ActionApprove:
		return e.processApprovalDecision(ctx, instance, action, true)
	case models.ActionReject:
		return e.processApprovalDecision(ctx, instance, action, false)
	case models.ActionDelegate:
		return e.processDelegation(ctx, instance, action, req.DelegateTo)
	case models.ActionRequest:
		return e.processInformationRequest(ctx, instance, action)
	default:
		return fmt.Errorf("unsupported action type: %s", req.Action)
	}
}

// processApprovalDecision handles approve/reject decisions
func (e *ApprovalEngine) processApprovalDecision(ctx context.Context, instance *models.ApprovalInstance, action *models.ApprovalAction, approved bool) error {
	// Get workflow definition
	workflow, err := e.repo.GetWorkflow(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	var steps []models.ApprovalStep
	if err := json.Unmarshal(workflow.Steps, &steps); err != nil {
		return fmt.Errorf("failed to parse workflow steps: %w", err)
	}

	currentStep := e.findStepByID(steps, instance.CurrentStep)
	if currentStep == nil {
		return fmt.Errorf("current step not found: %s", instance.CurrentStep)
	}

	// Check if step is complete based on approval criteria
	stepComplete, stepApproved := e.evaluateStepCompletion(ctx, instance, currentStep, approved)

	if !stepComplete {
		// Step is not complete, update instance status but keep in current step
		instance.Status = models.StatusInProgress
		return e.repo.UpdateApprovalInstance(ctx, instance)
	}

	if !stepApproved {
		// Step was rejected, end workflow
		instance.Status = models.StatusRejected
		instance.CompletedAt = &action.ActedAt
		if err := e.repo.UpdateApprovalInstance(ctx, instance); err != nil {
			return err
		}

		// Send rejection notifications
		return e.sendCompletionNotifications(ctx, instance, workflow, false)
	}

	// Step was approved, move to next step or complete workflow
	nextStep := e.determineNextStep(steps, instance.CurrentStep, instance.RequestData)
	if nextStep == "" {
		// Workflow complete
		instance.Status = models.StatusApproved
		instance.CompletedAt = &action.ActedAt
		if err := e.repo.UpdateApprovalInstance(ctx, instance); err != nil {
			return err
		}

		// Send completion notifications
		return e.sendCompletionNotifications(ctx, instance, workflow, true)
	}

	// Move to next step
	instance.CurrentStep = nextStep
	instance.Status = models.StatusInProgress
	if err := e.repo.UpdateApprovalInstance(ctx, instance); err != nil {
		return err
	}

	// Send notifications for new step
	return e.sendStepNotifications(ctx, instance, nextStep, workflow)
}

// evaluateStepCompletion determines if a step is complete and approved
func (e *ApprovalEngine) evaluateStepCompletion(ctx context.Context, instance *models.ApprovalInstance, step *models.ApprovalStep, latestApproved bool) (complete bool, approved bool) {
	// Get all actions for this step
	actions, err := e.repo.GetApprovalActionsByStep(ctx, instance.ID, step.ID)
	if err != nil {
		e.logger.Error("Failed to get approval actions", zap.Error(err))
		return false, false
	}

	approvals := 0
	rejections := 0

	for _, action := range actions {
		switch action.Action {
		case models.ActionApprove:
			approvals++
		case models.ActionReject:
			rejections++
		}
	}

	criteria := step.ApprovalCriteria

	// Check rejection criteria first
	if criteria.RejectOnFirst && rejections > 0 {
		return true, false
	}

	// Check approval criteria based on voting type
	switch criteria.VotingType {
	case models.VotingTypeSimple:
		if criteria.RequireAll {
			// All approvers must approve
			totalApprovers := len(step.Approvers)
			return approvals >= totalApprovers, approvals >= totalApprovers
		} else {
			// Simple majority or required count
			requiredApprovals := criteria.RequiredApprovals
			if requiredApprovals == 0 {
				requiredApprovals = (len(step.Approvers) / 2) + 1 // Simple majority
			}
			return approvals >= requiredApprovals, approvals >= requiredApprovals
		}

	case models.VotingTypeWeighted:
		return e.evaluateWeightedVoting(step, actions)

	case models.VotingTypeConsensus:
		totalApprovers := len(step.Approvers)
		return approvals >= totalApprovers, approvals >= totalApprovers

	case models.VotingTypeThreshold:
		// Implementation would depend on threshold configuration
		return approvals >= criteria.RequiredApprovals, approvals >= criteria.RequiredApprovals

	default:
		// Default to simple majority
		required := (len(step.Approvers) / 2) + 1
		return approvals >= required, approvals >= required
	}
}

// evaluateWeightedVoting handles weighted voting logic
func (e *ApprovalEngine) evaluateWeightedVoting(step *models.ApprovalStep, actions []models.ApprovalAction) (complete bool, approved bool) {
	if len(step.ApprovalCriteria.WeightedVoting) == 0 {
		return false, false
	}

	weightMap := make(map[string]float64)
	for _, weighted := range step.ApprovalCriteria.WeightedVoting {
		weightMap[weighted.ApproverID] = weighted.Weight
	}

	approvalWeight := 0.0
	rejectionWeight := 0.0
	totalWeight := 0.0

	for _, weighted := range step.ApprovalCriteria.WeightedVoting {
		totalWeight += weighted.Weight
	}

	for _, action := range actions {
		weight := weightMap[action.ApproverID.String()]
		switch action.Action {
		case models.ActionApprove:
			approvalWeight += weight
		case models.ActionReject:
			rejectionWeight += weight
		}
	}

	threshold := totalWeight * 0.5 // 50% threshold by default

	if rejectionWeight > threshold {
		return true, false
	}

	if approvalWeight > threshold {
		return true, true
	}

	return false, false
}

// ProcessEscalations handles escalation logic
func (e *ApprovalEngine) ProcessEscalations(ctx context.Context) error {
	// Get instances that need escalation
	instances, err := e.repo.GetInstancesForEscalation(ctx)
	if err != nil {
		return fmt.Errorf("failed to get instances for escalation: %w", err)
	}

	for _, instance := range instances {
		if err := e.escalateInstance(ctx, instance); err != nil {
			e.logger.Error("Failed to escalate instance",
				zap.Error(err),
				zap.String("instance_id", instance.ID.String()))
		}
	}

	return nil
}

// Helper functions

func (e *ApprovalEngine) determineInitialStep(steps []models.ApprovalStep, requestData json.RawMessage) string {
	// Find the first step based on conditions or order
	for _, step := range steps {
		if step.Order == 1 || len(step.Conditions) == 0 {
			return step.ID
		}

		// Evaluate conditions if present
		if e.evaluateConditions(step.Conditions, requestData) {
			return step.ID
		}
	}

	// Default to first step if found
	if len(steps) > 0 {
		return steps[0].ID
	}

	return ""
}

func (e *ApprovalEngine) determineNextStep(steps []models.ApprovalStep, currentStepID string, requestData json.RawMessage) string {
	currentStep := e.findStepByID(steps, currentStepID)
	if currentStep == nil {
		return ""
	}

	// Find next step based on order and conditions
	nextOrder := currentStep.Order + 1
	for _, step := range steps {
		if step.Order == nextOrder {
			if len(step.Conditions) == 0 || e.evaluateConditions(step.Conditions, requestData) {
				return step.ID
			}
		}
	}

	return ""
}

func (e *ApprovalEngine) findStepByID(steps []models.ApprovalStep, stepID string) *models.ApprovalStep {
	for _, step := range steps {
		if step.ID == stepID {
			return &step
		}
	}
	return nil
}

func (e *ApprovalEngine) evaluateConditions(conditions []models.Condition, requestData json.RawMessage) bool {
	if len(conditions) == 0 {
		return true
	}

	var data map[string]interface{}
	if err := json.Unmarshal(requestData, &data); err != nil {
		return false
	}

	// Simple condition evaluation - would be more sophisticated in production
	for _, condition := range conditions {
		value, exists := data[condition.Field]
		if !exists {
			return false
		}

		switch condition.Operator {
		case "eq":
			if value != condition.Value {
				return false
			}
		case "gt":
			if v, ok := value.(float64); ok {
				if target, ok := condition.Value.(float64); ok && v <= target {
					return false
				}
			}
		case "lt":
			if v, ok := value.(float64); ok {
				if target, ok := condition.Value.(float64); ok && v >= target {
					return false
				}
			}
		}
	}

	return true
}

func (e *ApprovalEngine) calculateSLAEndTime(workflow *models.ApprovalWorkflow, priority models.Priority) *time.Time {
	// Default SLA times based on priority
	var hours int
	switch priority {
	case models.PriorityCritical:
		hours = 4
	case models.PriorityHigh:
		hours = 24
	case models.PriorityMedium:
		hours = 72
	case models.PriorityLow:
		hours = 168 // 1 week
	default:
		hours = 72
	}

	slaTime := time.Now().Add(time.Duration(hours) * time.Hour)
	return &slaTime
}

// Request/Response types

type StartApprovalRequest struct {
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	RequestedBy    uuid.UUID              `json:"requested_by"`
	Priority       models.Priority        `json:"priority"`
	RequestData    json.RawMessage        `json:"request_data"`
	Comments       string                 `json:"comments"`
	AttachmentURLs []string               `json:"attachment_urls"`
	Metadata       json.RawMessage        `json:"metadata"`
}

type ApprovalActionRequest struct {
	InstanceID     uuid.UUID              `json:"instance_id"`
	ApproverID     uuid.UUID              `json:"approver_id"`
	Action         models.ActionType      `json:"action"`
	Comments       string                 `json:"comments"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	DelegatedFrom  *uuid.UUID             `json:"delegated_from,omitempty"`
	DelegateTo     *uuid.UUID             `json:"delegate_to,omitempty"`
	AttachmentURLs []string               `json:"attachment_urls"`
	Metadata       json.RawMessage        `json:"metadata"`
}

// Additional helper methods would be implemented for:
// - sendStepNotifications
// - sendCompletionNotifications
// - validateApprovalAction
// - processDelegation
// - processInformationRequest
// - escalateInstance