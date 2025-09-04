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
