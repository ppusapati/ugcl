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

// IApprovalService defines approval workflow service interface
type IApprovalService interface {
	// Approval process management
	StartApprovalProcess(ctx context.Context, req *StartApprovalRequest) (*db.FormInstance, error)
	ProcessApprovalAction(ctx context.Context, req *ProcessApprovalActionRequest) (*db.ApprovalAction, *db.FormInstance, error)

	// Approval actions
	CreateApprovalAction(ctx context.Context, action *CreateApprovalActionRequest) (*db.ApprovalAction, error)
	GetApprovalActionsByInstance(ctx context.Context, instanceID uuid.UUID) ([]*db.ApprovalAction, error)
	GetApprovalActionsByApprover(ctx context.Context, approverID string, limit, offset int32) ([]*db.ApprovalAction, error)

	// Delegation management
	CreateDelegate(ctx context.Context, delegate *CreateDelegateRequest) (*db.ApprovalDelegate, error)
	UpdateDelegate(ctx context.Context, delegateID uuid.UUID, req *UpdateDelegateRequest) (*db.ApprovalDelegate, error)
	DeleteDelegate(ctx context.Context, delegateID uuid.UUID) error
	GetDelegatesByDelegator(ctx context.Context, delegatorID string) ([]*db.ApprovalDelegate, error)
	GetDelegatesByDelegate(ctx context.Context, delegateID string) ([]*db.ApprovalDelegate, error)
	GetActiveDelegation(ctx context.Context, delegatorID, entityType string) (*db.ApprovalDelegate, error)

	// Pending approvals and workload
	GetPendingApprovalsForUser(ctx context.Context, userID string, limit, offset int32) ([]*PendingApprovalWithForm, error)
	GetApprovalHistory(ctx context.Context, instanceID uuid.UUID) ([]*ApprovalHistoryEntry, error)

	// Metrics and reporting
	GetApprovalMetrics(ctx context.Context, req *ApprovalMetricsRequest) (*ApprovalMetrics, error)
	CreateApprovalReport(ctx context.Context, report *CreateApprovalReportRequest) (*db.ApprovalReport, error)
	GetApprovalReportsByWorkflow(ctx context.Context, workflowID uuid.UUID, startDate, endDate *time.Time) ([]*db.ApprovalReport, error)
	GetApprovalReportsByEntityType(ctx context.Context, entityType string, startDate, endDate *time.Time) ([]*db.ApprovalReport, error)

	// Escalation management
	GetEscalatedInstances(ctx context.Context, limit, offset int32) ([]*FormInstanceWithTitle, error)
	ProcessEscalations(ctx context.Context) (int32, []string, error)
}
