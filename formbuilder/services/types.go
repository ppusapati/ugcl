// =============================================================================
// internal/services/types.go
// =============================================================================
package services

import (
	"encoding/json"
	"fmt"
	"time"

	db "p9e.in/ugcl/formbuilder/db/generated"
	validators "p9e.in/ugcl/formbuilder/validators"

	"github.com/google/uuid"
)

// Form Builder Service Types
type CreateFormResult struct {
	FormID    uuid.UUID
	TableName string
	Success   bool
	Message   string
}

type ListFormsResult struct {
	Forms    []*db.Form
	Total    int64
	Page     int32
	PageSize int32
}

type MigrateFormResult struct {
	Success         bool
	RecordsMigrated int32
	Warnings        []string
}

type FieldOption struct {
	Value    string
	Label    string
	Disabled bool
	Metadata map[string]string
}

type GetFieldOptionsResult struct {
	Options   []*FieldOption
	FromCache bool
	CachedAt  *time.Time
}

// Form Instance Service Types
type SubmitFormResult struct {
	InstanceID   uuid.UUID
	CurrentState string
	Success      bool
	Errors       []validators.ValidationError
}

type ListInstancesResult struct {
	Instances []*db.FormInstance
	Total     int64
	Page      int32
	PageSize  int32
}

type FormInstanceWithRelatedData struct {
	Instance    *db.FormInstance
	AuditLogs   []*db.AuditLog
	Attachments []*db.Attachment
	Comments    []*db.Comment
}

// Workflow Service Types
type GetWorkflowStatesResult struct {
	States           []WorkflowState
	CurrentState     string
	AvailableActions []string
}

type WorkflowState struct {
	ID            string               `json:"id"`
	Label         string               `json:"label"`
	AssignedRole  string               `json:"assigned_role"`
	AssignedUsers []string             `json:"assigned_users"`
	Actions       []string             `json:"actions"`
	Transitions   []WorkflowTransition `json:"transitions"`
	Type          StateType            `json:"type"`
	Properties    map[string]string    `json:"properties"`
	Validations   []StateValidation    `json:"validations"`
}

// StateType handles both numeric and string values during JSON unmarshaling
type StateType string

const (
	StateTypeStart    StateType = "START"
	StateTypeApproval StateType = "APPROVAL"
	StateTypeEnd      StateType = "END"
)

// UnmarshalJSON handles both numeric and string values for StateType
func (st *StateType) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first (database enum values)
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		switch num {
		case 1:
			*st = StateTypeStart
		case 2:
			*st = StateTypeEnd
		case 3:
			*st = StateTypeApproval
		default:
			*st = StateTypeStart // Default fallback
		}
		return nil
	}
	
	// Try to unmarshal as string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*st = StateType(str)
		return nil
	}
	
	return fmt.Errorf("cannot unmarshal StateType from %s", string(data))
}

type WorkflowTransition struct {
	Event     string            `json:"event"`
	Condition string            `json:"condition"`
	NextState string            `json:"next_state"` // Handle snake_case from database
	Actions   []TransitionAction `json:"actions"`
	Metadata  map[string]string  `json:"metadata"`
}

type TransitionAction struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type StateValidation struct {
	Expression string
	Message    string
	Blocking   bool
}

type TransitionResult struct {
	Success         bool
	NewState        string
	Message         string
	ExecutedActions []*TransitionAction
}

type ExternalWorkflowResult struct {
	ProcessInstanceID string
	Success           bool
	Message           string
}

// Analytics Service Types
type FormStatistics struct {
	FormID                   uuid.UUID
	FormTitle                string
	TotalInstances           int64
	CompletedInstances       int64
	DraftInstances           int64
	InReviewInstances        int64
	AvgCompletionTimeSeconds float64
}

type UserWorkload struct {
	AssignedTo           string
	AssignedCount        int64
	PendingApprovalCount int64
	InReviewCount        int64
}

type OverdueInstance struct {
	InstanceID   uuid.UUID
	FormTitle    string
	CurrentState string
	CreatedAt    time.Time
	AssignedTo   *string
}

type EscalationCandidate struct {
	InstanceID   uuid.UUID
	FormTitle    string
	CurrentState string
	UpdatedAt    time.Time
	AssignedTo   *string
}

type ActivityEntry struct {
	ID         uuid.UUID
	Action     string
	Timestamp  time.Time
	UserID     string
	InstanceID uuid.UUID
	FormTitle  string
}

type DashboardMetrics struct {
	TotalForms            int64
	TotalInstances        int64
	PendingInstances      int64
	CompletedToday        int64
	OverdueInstances      int64
	UserAssignedInstances int64
	RecentActivity        []*ActivityEntry
}

// Error types
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

type ServiceError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new service error
func NewServiceError(code, message string, details ...map[string]interface{}) *ServiceError {
	var detailsMap map[string]interface{}
	if len(details) > 0 {
		detailsMap = details[0]
	}

	return &ServiceError{
		Code:    code,
		Message: message,
		Details: detailsMap,
	}
}

// =============================================================================
// Approval Service Types
// =============================================================================

// Approval workflow requests
type StartApprovalRequest struct {
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	RequestedBy    string                 `json:"requested_by"`
	Priority       string                 `json:"priority"`
	RequestData    map[string]interface{} `json:"request_data"`
	Comments       string                 `json:"comments"`
	AttachmentURLs []string               `json:"attachment_urls"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type ProcessApprovalActionRequest struct {
	InstanceID     uuid.UUID              `json:"instance_id"`
	ApproverID     string                 `json:"approver_id"`
	Action         string                 `json:"action"` // APPROVE, REJECT, DELEGATE, REQUEST_INFO, WITHDRAW, REASSIGN
	Comments       string                 `json:"comments"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	DelegatedFrom  *string                `json:"delegated_from"`
	DelegateTo     *string                `json:"delegate_to"`
	AttachmentURLs []string               `json:"attachment_urls"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type CreateApprovalActionRequest struct {
	FormInstanceID uuid.UUID              `json:"form_instance_id"`
	ApproverID     string                 `json:"approver_id"`
	Action         string                 `json:"action"`
	Comments       string                 `json:"comments"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	DelegatedFrom  *string                `json:"delegated_from"`
	AttachmentURLs []string               `json:"attachment_urls"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Delegation requests
type CreateDelegateRequest struct {
	DelegatorID string    `json:"delegator_id"`
	DelegateID  string    `json:"delegate_id"`
	EntityTypes []string  `json:"entity_types"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Reason      string    `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type UpdateDelegateRequest struct {
	DelegateID  string     `json:"delegate_id"`
	EntityTypes []string   `json:"entity_types"`
	EndDate     *time.Time `json:"end_date"`
	IsActive    bool       `json:"is_active"`
	Reason      string     `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Reporting and metrics requests
type ApprovalMetricsRequest struct {
	EntityType string     `json:"entity_type"`
	WorkflowID *uuid.UUID `json:"workflow_id"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
}

type CreateApprovalReportRequest struct {
	WorkflowID        *uuid.UUID             `json:"workflow_id"`
	EntityType        string                 `json:"entity_type"`
	ReportDate        time.Time              `json:"report_date"`
	TotalRequests     int32                  `json:"total_requests"`
	ApprovedCount     int32                  `json:"approved_count"`
	RejectedCount     int32                  `json:"rejected_count"`
	PendingCount      int32                  `json:"pending_count"`
	EscalatedCount    int32                  `json:"escalated_count"`
	AvgProcessingTime float64                `json:"avg_processing_time_hours"`
	SLABreachCount    int32                  `json:"sla_breach_count"`
	SLAComplianceRate float64                `json:"sla_compliance_rate"`
	BottleneckSteps   []string               `json:"bottleneck_steps"`
	TopApprovers      []string               `json:"top_approvers"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// Response types
type ApprovalMetrics struct {
	TotalRequests          int32   `json:"total_requests"`
	ApprovedCount          int32   `json:"approved_count"`
	RejectedCount          int32   `json:"rejected_count"`
	PendingCount           int32   `json:"pending_count"`
	EscalatedCount         int32   `json:"escalated_count"`
	AvgProcessingTimeHours float64 `json:"avg_processing_time_hours"`
	SLAComplianceRate      float64 `json:"sla_compliance_rate"`
}

type PendingApprovalWithForm struct {
	Instance       *db.FormInstance `json:"instance"`
	FormTitle      string           `json:"form_title"`
	FormDescription string          `json:"form_description"`
}

type FormInstanceWithTitle struct {
	Instance  *db.FormInstance `json:"instance"`
	FormTitle string           `json:"form_title"`
}

type ApprovalHistoryEntry struct {
	ApprovalAction *db.ApprovalAction `json:"approval_action"`
	AuditAction    string             `json:"audit_action"`
	Changes        json.RawMessage    `json:"changes"`
	AuditTimestamp time.Time          `json:"audit_timestamp"`
}
