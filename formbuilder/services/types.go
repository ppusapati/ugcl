// =============================================================================
// internal/services/types.go
// =============================================================================
package services

import (
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
	ID            string
	Label         string
	AssignedRole  string
	AssignedUsers []string
	Actions       []string
	Transitions   []WorkflowTransition
	Type          string
	Properties    map[string]string
	Validations   []StateValidation
}

type WorkflowTransition struct {
	Event     string
	Condition string
	NextState string
	Actions   []TransitionAction
	Metadata  map[string]string
}

type TransitionAction struct {
	Type   string
	Params map[string]string
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
