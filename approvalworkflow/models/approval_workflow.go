package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ApprovalWorkflow defines the complete approval workflow configuration
type ApprovalWorkflow struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	EntityType  string    `json:"entity_type" db:"entity_type"` // e.g., 'purchase_order', 'expense', 'leave_request'
	Version     int       `json:"version" db:"version"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsDefault   bool      `json:"is_default" db:"is_default"`

	// Workflow configuration
	Steps          json.RawMessage `json:"steps" db:"steps"`                     // ApprovalStep[]
	Rules          json.RawMessage `json:"rules" db:"rules"`                     // ApprovalRule[]
	Escalations    json.RawMessage `json:"escalations" db:"escalations"`         // EscalationRule[]
	Notifications  json.RawMessage `json:"notifications" db:"notifications"`     // NotificationRule[]
	ConditionalLogic json.RawMessage `json:"conditional_logic" db:"conditional_logic"` // ConditionalRule[]

	// Metadata
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Metadata  json.RawMessage `json:"metadata" db:"metadata"`
}

// ApprovalStep represents a single step in the approval workflow
type ApprovalStep struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	StepType    StepType `json:"step_type"` // SEQUENTIAL, PARALLEL, CONDITIONAL, ESCALATION
	Order       int      `json:"order"`
	IsRequired  bool     `json:"is_required"`

	// Approval configuration
	Approvers        []Approver        `json:"approvers"`
	ApprovalCriteria ApprovalCriteria  `json:"approval_criteria"`
	TimeoutDuration  time.Duration     `json:"timeout_duration"`
	EscalationRules  []EscalationRule  `json:"escalation_rules"`

	// Conditional logic
	Conditions []Condition `json:"conditions"`

	// Step metadata
	Properties map[string]interface{} `json:"properties"`
}

type StepType string

const (
	StepTypeSequential  StepType = "SEQUENTIAL"
	StepTypeParallel    StepType = "PARALLEL"
	StepTypeConditional StepType = "CONDITIONAL"
	StepTypeEscalation  StepType = "ESCALATION"
	StepTypeAutomatic   StepType = "AUTOMATIC"
)

// Approver defines who can approve at each step
type Approver struct {
	ID           string       `json:"id"`
	Type         ApproverType `json:"type"`
	Value        string       `json:"value"` // user_id, role_name, etc.
	DisplayName  string       `json:"display_name"`
	IsAlternate  bool         `json:"is_alternate"`
	Priority     int          `json:"priority"`
	CanDelegate  bool         `json:"can_delegate"`
	MaxAmount    *float64     `json:"max_amount,omitempty"` // For amount-based approvals
}

type ApproverType string

const (
	ApproverTypeUser         ApproverType = "USER"
	ApproverTypeRole         ApproverType = "ROLE"
	ApproverTypeDepartment   ApproverType = "DEPARTMENT"
	ApproverTypeManager      ApproverType = "MANAGER" // Direct manager
	ApproverTypeSkipManager  ApproverType = "SKIP_MANAGER" // Skip-level manager
	ApproverTypeCustomQuery  ApproverType = "CUSTOM_QUERY" // Based on custom logic
)

// ApprovalCriteria defines how approvals are evaluated
type ApprovalCriteria struct {
	RequiredApprovals int                `json:"required_approvals"` // Number of approvals needed
	RequireAll        bool               `json:"require_all"`        // All approvers must approve
	RejectOnFirst     bool               `json:"reject_on_first"`    // Reject if first approver rejects
	VotingType        VotingType         `json:"voting_type"`
	WeightedVoting    []WeightedApprover `json:"weighted_voting,omitempty"`
}

type VotingType string

const (
	VotingTypeSimple     VotingType = "SIMPLE"     // Simple majority
	VotingTypeWeighted   VotingType = "WEIGHTED"   // Based on weights
	VotingTypeConsensus  VotingType = "CONSENSUS"  // All must agree
	VotingTypeThreshold  VotingType = "THRESHOLD"  // Percentage threshold
)

type WeightedApprover struct {
	ApproverID string  `json:"approver_id"`
	Weight     float64 `json:"weight"`
}

// EscalationRule defines escalation behavior
type EscalationRule struct {
	ID              string        `json:"id"`
	TriggerAfter    time.Duration `json:"trigger_after"`
	EscalationLevel int           `json:"escalation_level"`
	EscalateToType  ApproverType  `json:"escalate_to_type"`
	EscalateToValue string        `json:"escalate_to_value"`
	Action          EscalationAction `json:"action"`
	NotifyOriginal  bool          `json:"notify_original"`
	MaxEscalations  int           `json:"max_escalations"`
}

type EscalationAction string

const (
	EscalationActionNotify     EscalationAction = "NOTIFY"
	EscalationActionReassign   EscalationAction = "REASSIGN"
	EscalationActionAutoApprove EscalationAction = "AUTO_APPROVE"
	EscalationActionReject     EscalationAction = "REJECT"
)

// Condition for conditional workflows
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, gt, lt, in, contains
	Value    interface{} `json:"value"`
	LogicOp  string      `json:"logic_op"` // AND, OR
}

// ApprovalInstance represents a specific approval request
type ApprovalInstance struct {
	ID               uuid.UUID `json:"id" db:"id"`
	WorkflowID       uuid.UUID `json:"workflow_id" db:"workflow_id"`
	EntityType       string    `json:"entity_type" db:"entity_type"`
	EntityID         string    `json:"entity_id" db:"entity_id"`
	RequestedBy      uuid.UUID `json:"requested_by" db:"requested_by"`
	CurrentStep      string    `json:"current_step" db:"current_step"`
	Status           ApprovalStatus `json:"status" db:"status"`
	Priority         Priority  `json:"priority" db:"priority"`

	// Request data
	RequestData     json.RawMessage `json:"request_data" db:"request_data"`
	Comments        string          `json:"comments" db:"comments"`
	AttachmentURLs  []string        `json:"attachment_urls" db:"attachment_urls"`

	// Tracking
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	EscalatedAt     *time.Time      `json:"escalated_at" db:"escalated_at"`
	EscalationLevel int             `json:"escalation_level" db:"escalation_level"`

	// SLA tracking
	SLAStartTime    time.Time  `json:"sla_start_time" db:"sla_start_time"`
	SLAEndTime      *time.Time `json:"sla_end_time" db:"sla_end_time"`
	SLABreached     bool       `json:"sla_breached" db:"sla_breached"`

	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type ApprovalStatus string

const (
	StatusPending    ApprovalStatus = "PENDING"
	StatusInProgress ApprovalStatus = "IN_PROGRESS"
	StatusApproved   ApprovalStatus = "APPROVED"
	StatusRejected   ApprovalStatus = "REJECTED"
	StatusCancelled  ApprovalStatus = "CANCELLED"
	StatusEscalated  ApprovalStatus = "ESCALATED"
	StatusExpired    ApprovalStatus = "EXPIRED"
)

type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
	PriorityCritical Priority = "CRITICAL"
)

// ApprovalAction represents individual approval actions
type ApprovalAction struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	InstanceID      uuid.UUID    `json:"instance_id" db:"instance_id"`
	StepID          string       `json:"step_id" db:"step_id"`
	ApproverID      uuid.UUID    `json:"approver_id" db:"approver_id"`
	Action          ActionType   `json:"action" db:"action"`
	Comments        string       `json:"comments" db:"comments"`
	ActedAt         time.Time    `json:"acted_at" db:"acted_at"`
	IPAddress       string       `json:"ip_address" db:"ip_address"`
	UserAgent       string       `json:"user_agent" db:"user_agent"`
	DelegatedFrom   *uuid.UUID   `json:"delegated_from" db:"delegated_from"`
	AttachmentURLs  []string     `json:"attachment_urls" db:"attachment_urls"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type ActionType string

const (
	ActionApprove   ActionType = "APPROVE"
	ActionReject    ActionType = "REJECT"
	ActionDelegate  ActionType = "DELEGATE"
	ActionRequest   ActionType = "REQUEST_INFO"
	ActionWithdraw  ActionType = "WITHDRAW"
	ActionReassign  ActionType = "REASSIGN"
)

// ApprovalDelegate represents delegation relationships
type ApprovalDelegate struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	DelegatorID  uuid.UUID  `json:"delegator_id" db:"delegator_id"`
	DelegateID   uuid.UUID  `json:"delegate_id" db:"delegate_id"`
	EntityTypes  []string   `json:"entity_types" db:"entity_types"` // What can be delegated
	StartDate    time.Time  `json:"start_date" db:"start_date"`
	EndDate      *time.Time `json:"end_date" db:"end_date"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	Reason       string     `json:"reason" db:"reason"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	Metadata     json.RawMessage `json:"metadata" db:"metadata"`
}

// ApprovalReport for analytics and reporting
type ApprovalReport struct {
	ID                uuid.UUID `json:"id" db:"id"`
	WorkflowID        uuid.UUID `json:"workflow_id" db:"workflow_id"`
	EntityType        string    `json:"entity_type" db:"entity_type"`
	ReportDate        time.Time `json:"report_date" db:"report_date"`

	// Metrics
	TotalRequests     int           `json:"total_requests" db:"total_requests"`
	ApprovedCount     int           `json:"approved_count" db:"approved_count"`
	RejectedCount     int           `json:"rejected_count" db:"rejected_count"`
	PendingCount      int           `json:"pending_count" db:"pending_count"`
	EscalatedCount    int           `json:"escalated_count" db:"escalated_count"`

	// Performance metrics
	AvgProcessingTime time.Duration `json:"avg_processing_time" db:"avg_processing_time"`
	SLABreachCount    int           `json:"sla_breach_count" db:"sla_breach_count"`
	SLAComplianceRate float64       `json:"sla_compliance_rate" db:"sla_compliance_rate"`

	// Bottleneck analysis
	BottleneckSteps   json.RawMessage `json:"bottleneck_steps" db:"bottleneck_steps"`
	TopApprovers      json.RawMessage `json:"top_approvers" db:"top_approvers"`

	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	Metadata          json.RawMessage `json:"metadata" db:"metadata"`
}