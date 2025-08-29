// =============================================================================
// internal/services/interfaces.go
// =============================================================================
package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "p9e.in/ugcl/formbuilder/api/v2/form_builder"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// IFormBuilderService defines form builder service interface
type IFormBuilderService interface {
	CreateForm(ctx context.Context, form *pb.FormDefinition, createTable bool) (*CreateFormResult, error)
	GetForm(ctx context.Context, formID string, version *string) (*db.Form, error)
	UpdateForm(ctx context.Context, form *db.Form) (*CreateFormResult, error)
	DeleteForm(ctx context.Context, formID string) error
	ListForms(ctx context.Context, page, pageSize int32, filter, sortBy string) (*ListFormsResult, error)
	MigrateFormVersion(ctx context.Context, formID, fromVersion, toVersion string, migrateData bool) (*MigrateFormResult, error)
	GetFieldOptions(ctx context.Context, formID, fieldID string, context map[string]string) (*GetFieldOptionsResult, error)
	RefreshFieldCache(ctx context.Context, formID, fieldID string) error
}

// IFormInstanceService defines form instance service interface
type IFormInstanceService interface {
	SubmitForm(ctx context.Context, formID string, fieldValues map[string]interface{}, action string, userID uuid.UUID) (*SubmitFormResult, error)
	GetFormInstance(ctx context.Context, instanceID string) (*db.FormInstance, error)
	GetFormInstanceWithRelated(ctx context.Context, instanceID string) (*FormInstanceWithRelatedData, error)
	UpdateFormInstance(ctx context.Context, instanceID string, fieldValues map[string]interface{}, action string, userID uuid.UUID) (*SubmitFormResult, error)
	DeleteFormInstance(ctx context.Context, instanceID string) error
	ListFormInstances(ctx context.Context, formID string, page, pageSize int32) (*ListInstancesResult, error)
	GetFormInstancesByState(ctx context.Context, formID, state string, page, pageSize int32) (*ListInstancesResult, error)
	GetFormInstancesByAssignee(ctx context.Context, assignedTo string, page, pageSize int32) (*ListInstancesResult, error)

	// Audit log methods
	GetAuditLogs(ctx context.Context, instanceID string) ([]*db.AuditLog, error)
	GetAuditLogsByUser(ctx context.Context, userID string, page, pageSize int32) ([]*db.AuditLog, error)

	// Attachment methods
	CreateAttachment(ctx context.Context, attachment *db.Attachment) (*db.Attachment, error)
	GetAttachments(ctx context.Context, instanceID string) ([]*db.Attachment, error)
	DeleteAttachment(ctx context.Context, attachmentID string) error

	// Comment methods
	CreateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error)
	GetComments(ctx context.Context, instanceID string) ([]*db.Comment, error)
	UpdateComment(ctx context.Context, commentID, text string) (*db.Comment, error)
	DeleteComment(ctx context.Context, commentID string) error
}

// IWorkflowService defines workflow service interface
type IWorkflowService interface {
	GetWorkflowStates(ctx context.Context, formID string) (*GetWorkflowStatesResult, error)
	TransitionWorkflow(ctx context.Context, instanceID, event string, userID uuid.UUID, context map[string]string) (*TransitionResult, error)
	TriggerExternalWorkflow(ctx context.Context, instanceID, action string, variables map[string]interface{}) (*ExternalWorkflowResult, error)

	// Workflow management
	CreateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error)
	GetWorkflow(ctx context.Context, workflowID string) (*db.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error)
	DeleteWorkflow(ctx context.Context, workflowID string) error

	// SLA management
	CreateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error)
	GetSLARules(ctx context.Context, state string) ([]*db.SlaRule, error)
	UpdateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error)
	DeleteSLARule(ctx context.Context, ruleID string) error

	// Escalation management
	CreateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error)
	GetEscalations(ctx context.Context, fromState string) ([]*db.Escalation, error)
	UpdateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error)
	DeleteEscalation(ctx context.Context, escalationID string) error
}

// IAnalyticsService defines analytics service interface
type IAnalyticsService interface {
	GetFormStatistics(ctx context.Context, formID string) (*FormStatistics, error)
	GetUserWorkload(ctx context.Context) ([]*UserWorkload, error)
	GetOverdueInstances(ctx context.Context) ([]*OverdueInstance, error)
	GetInstancesRequiringEscalation(ctx context.Context) ([]*EscalationCandidate, error)
	GetRecentActivity(ctx context.Context, since time.Time, page, pageSize int32) ([]*ActivityEntry, error)
	GetDashboardMetrics(ctx context.Context, userID string) (*DashboardMetrics, error)
}
