// =============================================================================
// internal/repository/form_instance_repository.go
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/utils"

	"github.com/google/uuid"
)

type formInstanceRepository struct {
	queries *db.Queries
}

// NewFormInstanceRepository creates a new form instance repository
func NewFormInstanceRepository(queries *db.Queries) IFormInstanceRepository {
	return &formInstanceRepository{
		queries: queries,
	}
}

func (r *formInstanceRepository) CreateFormInstance(ctx context.Context, instance *db.FormInstance) (*db.FormInstance, error) {
	// Set timestamps and defaults
	if instance.ID == uuid.Nil {
		instance.ID = utils.GenerateUUID()
	}

	params := db.CreateFormInstanceParams{
		ID:           instance.ID,
		FormID:       instance.FormID,
		CurrentState: instance.CurrentState,
		FieldValues:  instance.FieldValues,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    instance.CreatedBy,
		AssignedTo:   instance.AssignedTo,
		Metadata:     instance.Metadata,
	}

	return r.queries.CreateFormInstance(ctx, params)
}

func (r *formInstanceRepository) GetFormInstance(ctx context.Context, instanceID uuid.UUID) (*db.FormInstance, error) {

	instance, err := r.queries.GetFormInstance(ctx, db.GetFormInstanceParams{ID: instanceID})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (r *formInstanceRepository) UpdateFormInstance(ctx context.Context, instance *db.FormInstance) (*db.FormInstance, error) {
	params := db.UpdateFormInstanceParams{
		ID:           instance.ID,
		CurrentState: instance.CurrentState,
		FieldValues:  instance.FieldValues,
		UpdatedAt:    time.Now(),
		AssignedTo:   instance.AssignedTo,
		Metadata:     instance.Metadata,
	}

	return r.queries.UpdateFormInstance(ctx, params)
}

func (r *formInstanceRepository) DeleteFormInstance(ctx context.Context, instanceID uuid.UUID) error {
	return r.queries.DeleteFormInstance(ctx, db.DeleteFormInstanceParams{ID: instanceID})
}

func (r *formInstanceRepository) ListFormInstances(ctx context.Context, formID uuid.UUID, limit, offset int32) ([]*db.FormInstance, error) {
	params := db.ListFormInstancesParams{
		FormID: formID,
		Limit:  limit,
		Offset: offset,
	}

	instances, err := r.queries.ListFormInstances(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.FormInstance, len(instances))
	copy(result, instances)
	return result, nil
}

func (r *formInstanceRepository) CountFormInstances(ctx context.Context, formID uuid.UUID) (int64, error) {
	return r.queries.CountFormInstances(ctx, db.CountFormInstancesParams{FormID: formID})
}

func (r *formInstanceRepository) GetFormInstancesByState(ctx context.Context, formID uuid.UUID, state string, limit, offset int32) ([]*db.FormInstance, error) {
	params := db.GetFormInstancesByStateParams{
		FormID:       formID,
		CurrentState: state,
		Limit:        limit,
		Offset:       offset,
	}

	instances, err := r.queries.GetFormInstancesByState(ctx, params)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *formInstanceRepository) GetFormInstancesByAssignee(ctx context.Context, assignedTo string, limit, offset int32) ([]*db.FormInstance, error) {
	params := db.GetFormInstancesByAssigneeParams{
		AssignedTo: &assignedTo,
		Limit:      limit,
		Offset:     offset,
	}

	instances, err := r.queries.GetFormInstancesByAssignee(ctx, params)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *formInstanceRepository) GetFormInstancesByCreator(ctx context.Context, createdBy string, limit, offset int32) ([]*db.FormInstance, error) {
	params := db.GetFormInstancesByCreatorParams{
		CreatedBy: createdBy,
		Limit:     limit,
		Offset:    offset,
	}

	instances, err := r.queries.GetFormInstancesByCreator(ctx, params)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *formInstanceRepository) UpdateFormInstanceState(ctx context.Context, instanceID uuid.UUID, state string, assignedTo *string) (*db.FormInstance, error) {
	params := db.UpdateFormInstanceStateParams{
		ID:           instanceID,
		CurrentState: state,
		AssignedTo:   assignedTo,
	}

	return r.queries.UpdateFormInstanceState(ctx, params)
}

// Audit log methods
func (r *formInstanceRepository) CreateAuditLog(ctx context.Context, log *db.AuditLog) (*db.AuditLog, error) {
	if log.ID == uuid.Nil {
		log.ID = utils.GenerateUUID()
	}

	params := db.CreateAuditLogParams{
		ID:         log.ID,
		InstanceID: log.InstanceID,
		UserID:     log.UserID,
		Action:     log.Action,
		FromState:  log.FromState,
		ToState:    log.ToState,
		Changes:    log.Changes,
		Timestamp:  time.Now(),
		IpAddress:  log.IpAddress,
		UserAgent:  log.UserAgent,
	}

	return r.queries.CreateAuditLog(ctx, params)
}

func (r *formInstanceRepository) GetAuditLogs(ctx context.Context, instanceID uuid.UUID) ([]*db.AuditLog, error) {
	logs, err := r.queries.GetAuditLogs(ctx, db.GetAuditLogsParams{InstanceID: instanceID})
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *formInstanceRepository) GetAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*db.AuditLog, error) {
	params := db.GetAuditLogsByUserParams{
		UserID: userID.String(),
		Limit:  limit,
		Offset: offset,
	}

	logs, err := r.queries.GetAuditLogsByUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *formInstanceRepository) GetAuditLogsByAction(ctx context.Context, instanceID uuid.UUID, action string) ([]*db.AuditLog, error) {
	params := db.GetAuditLogsByActionParams{
		InstanceID: instanceID,
		Action:     action,
	}

	logs, err := r.queries.GetAuditLogsByAction(ctx, params)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *formInstanceRepository) GetRecentAuditLogs(ctx context.Context, since time.Time, limit, offset int32) ([]*db.AuditLog, error) {
	params := db.GetRecentAuditLogsParams{
		Timestamp: since,
		Limit:     limit,
		Offset:    offset,
	}

	logs, err := r.queries.GetRecentAuditLogs(ctx, params)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// Attachment methods
func (r *formInstanceRepository) CreateAttachment(ctx context.Context, attachment *db.Attachment) (*db.Attachment, error) {
	if attachment.ID == uuid.Nil {
		attachment.ID = utils.GenerateUUID()
	}

	params := db.CreateAttachmentParams{
		ID:          attachment.ID,
		InstanceID:  attachment.InstanceID,
		FieldID:     attachment.FieldID,
		Filename:    attachment.Filename,
		MimeType:    attachment.MimeType,
		Size:        attachment.Size,
		StoragePath: attachment.StoragePath,
		UploadedAt:  time.Now(),
		UploadedBy:  attachment.UploadedBy,
	}

	return r.queries.CreateAttachment(ctx, params)
}

func (r *formInstanceRepository) GetAttachments(ctx context.Context, instanceID uuid.UUID) ([]*db.Attachment, error) {
	attachments, err := r.queries.GetAttachments(ctx, db.GetAttachmentsParams{InstanceID: instanceID})
	if err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *formInstanceRepository) GetAttachmentsByField(ctx context.Context, instanceID uuid.UUID, fieldID string) ([]*db.Attachment, error) {
	params := db.GetAttachmentsByFieldParams{
		InstanceID: instanceID,
		FieldID:    fieldID,
	}

	attachments, err := r.queries.GetAttachmentsByField(ctx, params)
	if err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *formInstanceRepository) GetAttachment(ctx context.Context, attachmentID uuid.UUID) (*db.Attachment, error) {
	attachment, err := r.queries.GetAttachment(ctx, db.GetAttachmentParams{ID: attachmentID})
	if err != nil {
		return nil, err
	}
	return attachment, nil
}

func (r *formInstanceRepository) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	return r.queries.DeleteAttachment(ctx, db.DeleteAttachmentParams{ID: attachmentID})
}

func (r *formInstanceRepository) GetAttachmentsByUser(ctx context.Context, uploadedBy string, limit, offset int32) ([]*db.Attachment, error) {
	params := db.GetAttachmentsByUserParams{
		UploadedBy: uploadedBy,
		Limit:      limit,
		Offset:     offset,
	}

	attachments, err := r.queries.GetAttachmentsByUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return attachments, nil
}

// Comment methods
func (r *formInstanceRepository) CreateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error) {
	if comment.ID == uuid.Nil {
		comment.ID = utils.GenerateUUID()
	}

	params := db.CreateCommentParams{
		ID:         comment.ID,
		InstanceID: comment.InstanceID,
		UserID:     comment.UserID,
		Text:       comment.Text,
		CreatedAt:  time.Now(),
		Internal:   comment.Internal,
	}

	return r.queries.CreateComment(ctx, params)
}

func (r *formInstanceRepository) GetComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error) {
	comments, err := r.queries.GetComments(ctx, db.GetCommentsParams{InstanceID: instanceID})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *formInstanceRepository) GetPublicComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error) {
	comments, err := r.queries.GetPublicComments(ctx, db.GetPublicCommentsParams{InstanceID: instanceID})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *formInstanceRepository) GetInternalComments(ctx context.Context, instanceID uuid.UUID) ([]*db.Comment, error) {
	comments, err := r.queries.GetInternalComments(ctx, db.GetInternalCommentsParams{InstanceID: instanceID})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *formInstanceRepository) UpdateComment(ctx context.Context, comment *db.Comment) (*db.Comment, error) {
	params := db.UpdateCommentParams{
		ID:   comment.ID,
		Text: comment.Text,
	}

	return r.queries.UpdateComment(ctx, params)
}

func (r *formInstanceRepository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return r.queries.DeleteComment(ctx, db.DeleteCommentParams{ID: commentID})
}

func (r *formInstanceRepository) GetCommentsByUser(ctx context.Context, userID string, limit, offset int32) ([]*db.Comment, error) {
	params := db.GetCommentsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	comments, err := r.queries.GetCommentsByUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return comments, nil
}
