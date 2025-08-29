// =============================================================================
// internal/repository/interfaces.go
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/formbuilder/db/generated"

	"github.com/google/uuid"
)

// IFormBuilderRepository defines form builder repository interface
type IFormBuilderRepository interface {
	CreateForm(ctx context.Context, form *db.Form) (*db.Form, error)
	GetForm(ctx context.Context, formID uuid.UUID) (*db.Form, error)
	GetFormByVersion(ctx context.Context, formID uuid.UUID, version string) (*db.Form, error)
	UpdateForm(ctx context.Context, form *db.Form) (*db.Form, error)
	DeleteForm(ctx context.Context, formID uuid.UUID) error
	ListForms(ctx context.Context, limit, offset int32) ([]*db.Form, error)
	CountForms(ctx context.Context) (int64, error)
	SearchForms(ctx context.Context, query string, limit, offset int32) ([]*db.Form, error)
	GetFormsByCreator(ctx context.Context, createdBy string, limit, offset int32) ([]*db.Form, error)
}

// IFormInstanceRepository defines form instance repository interface
type IFormInstanceRepository interface {
	CreateFormInstance(ctx context.Context, instance *db.FormInstance) (*db.FormInstance, error)
	GetFormInstance(ctx context.Context, instanceID uuid.UUID) (*db.FormInstance, error)
	UpdateFormInstance(ctx context.Context, instance *db.FormInstance) (*db.FormInstance, error)
	DeleteFormInstance(ctx context.Context, instanceID uuid.UUID) error
	ListFormInstances(ctx context.Context, formID uuid.UUID, limit, offset int32) ([]*db.FormInstance, error)
	CountFormInstances(ctx context.Context, formID uuid.UUID) (int64, error)
	GetFormInstancesByState(ctx context.Context, formID uuid.UUID, state string, limit, offset int32) ([]*db.FormInstance, error)
	GetFormInstancesByAssignee(ctx context.Context, assignedTo string, limit, offset int32) ([]*db.FormInstance, error)
	GetFormInstancesByCreator(ctx context.Context, createdBy string, limit, offset int32) ([]*db.FormInstance, error)
	UpdateFormInstanceState(ctx context.Context, instanceID uuid.UUID, state string, assignedTo *string) (*db.FormInstance, error)

	// Audit log methods
	CreateAuditLog(ctx context.Context, log *db.AuditLog) (*db.AuditLog, error)
	GetAuditLogs(ctx context.Context, instanceID uuid.UUID) ([]*db.AuditLog, error)
	GetAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*db.AuditLog, error)
	GetAuditLogsByAction(ctx context.Context, instanceID uuid.UUID, action string) ([]*db.AuditLog, error)
	GetRecentAuditLogs(ctx context.Context, since time.Time, limit, offset int32) ([]*db.AuditLog, error)

	// Attachment methods
	CreateAttachment(ctx context.Context, attachment *db.Attachment) (*db.Attachment, error)
	GetAttachments(ctx context.Context, instanceID uuid.UUID) ([]*db.Attachment, error)
	GetAttachmentsByField(ctx context.Context, instanceID uuid.UUID, fieldID string) ([]*db.Attachment, error)
	GetAttachment(ctx context.Context, attachmentID uuid.UUID) (*db.Attachment, error)
	DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error
	GetAttachmentsByUser(ctx context.Context, uploadedBy string, limit, offset int32) ([]*db.Attachment, error)

	// Comment methods
	CreateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error)
	GetComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error)
	GetPublicComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error)
	GetInternalComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error)
	UpdateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
	GetCommentsByUser(ctx context.Context, userID string, limit, offset int32) ([]*db.Comment, error)
}

// IWorkflowRepository defines workflow repository interface
type IWorkflowRepository interface {
	CreateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error)
	GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*db.Workflow, error)
	GetWorkflowByFormID(ctx context.Context, formID uuid.UUID) (*db.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *db.Workflow) (*db.Workflow, error)
	DeleteWorkflow(ctx context.Context, workflowID uuid.UUID) error
	ListWorkflows(ctx context.Context, limit, offset int32) ([]*db.Workflow, error)
	CountWorkflows(ctx context.Context) (int64, error)

	// SLA rule methods
	CreateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error)
	GetSLARule(ctx context.Context, ruleID uuid.UUID) (*db.SlaRule, error)
	GetSLARules(ctx context.Context) ([]*db.SlaRule, error)
	GetSLARulesByState(ctx context.Context, state string) ([]*db.SlaRule, error)
	GetActiveSLARules(ctx context.Context) ([]*db.SlaRule, error)
	UpdateSLARule(ctx context.Context, rule *db.SlaRule) (*db.SlaRule, error)
	DeleteSLARule(ctx context.Context, ruleID uuid.UUID) error
	ToggleSLARule(ctx context.Context, ruleID uuid.UUID) (*db.SlaRule, error)

	// Escalation methods
	CreateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error)
	GetEscalation(ctx context.Context, escalationID uuid.UUID) (*db.Escalation, error)
	GetEscalations(ctx context.Context) ([]*db.Escalation, error)
	GetEscalationsByFromState(ctx context.Context, fromState string) ([]*db.Escalation, error)
	GetAutoEscalations(ctx context.Context) ([]*db.Escalation, error)
	UpdateEscalation(ctx context.Context, escalation *db.Escalation) (*db.Escalation, error)
	DeleteEscalation(ctx context.Context, escalationID uuid.UUID) error
	ToggleAutoEscalation(ctx context.Context, escalationID uuid.UUID) (*db.Escalation, error)
}

// IAnalyticsRepository defines analytics and reporting repository interface
type IAnalyticsRepository interface {
	GetFormStatistics(ctx context.Context, formID uuid.UUID) (*db.GetFormStatisticsRow, error)
	GetUserWorkload(ctx context.Context) ([]*db.GetUserWorkloadRow, error)
	GetOverdueInstances(ctx context.Context) ([]*db.GetOverdueInstancesRow, error)
	GetInstancesRequiringEscalation(ctx context.Context) ([]*db.GetInstancesRequiringEscalationRow, error)
	GetRecentActivity(ctx context.Context, since time.Time, limit, offset int32) ([]*db.GetRecentActivityRow, error)
}
