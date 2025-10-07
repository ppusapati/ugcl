package handlers

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	approvalv2 "p9e.in/ugcl/formbuilder/api/v2/approval"
	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/services"
)

// Helper methods for approval handler

func (h *ApprovalHandler) convertProtoPriority(priority approvalv2.Priority) string {
	switch priority {
	case approvalv2.Priority_PRIORITY_LOW:
		return "LOW"
	case approvalv2.Priority_PRIORITY_MEDIUM:
		return "MEDIUM"
	case approvalv2.Priority_PRIORITY_HIGH:
		return "HIGH"
	case approvalv2.Priority_PRIORITY_CRITICAL:
		return "CRITICAL"
	default:
		return "MEDIUM"
	}
}

func (h *ApprovalHandler) convertProtoAction(action approvalv2.ActionType) string {
	switch action {
	case approvalv2.ActionType_ACTION_TYPE_APPROVE:
		return "APPROVE"
	case approvalv2.ActionType_ACTION_TYPE_REJECT:
		return "REJECT"
	case approvalv2.ActionType_ACTION_TYPE_DELEGATE:
		return "DELEGATE"
	case approvalv2.ActionType_ACTION_TYPE_REQUEST_INFO:
		return "REQUEST_INFO"
	case approvalv2.ActionType_ACTION_TYPE_WITHDRAW:
		return "WITHDRAW"
	case approvalv2.ActionType_ACTION_TYPE_REASSIGN:
		return "REASSIGN"
	default:
		return "APPROVE"
	}
}

func (h *ApprovalHandler) convertProtoStatus(status approvalv2.ApprovalStatus) string {
	switch status {
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_PENDING:
		return "PENDING"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_IN_PROGRESS:
		return "IN_PROGRESS"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_APPROVED:
		return "APPROVED"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_REJECTED:
		return "REJECTED"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_CANCELLED:
		return "CANCELLED"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_ESCALATED:
		return "ESCALATED"
	case approvalv2.ApprovalStatus_APPROVAL_STATUS_EXPIRED:
		return "EXPIRED"
	default:
		return "PENDING"
	}
}

func (h *ApprovalHandler) convertStringToProtoStatus(status string) approvalv2.ApprovalStatus {
	switch status {
	case "PENDING", "pending_approval":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_PENDING
	case "IN_PROGRESS", "in_review":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_IN_PROGRESS
	case "APPROVED", "approved":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_APPROVED
	case "REJECTED", "rejected":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_REJECTED
	case "CANCELLED", "cancelled":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_CANCELLED
	case "ESCALATED", "escalated":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_ESCALATED
	case "EXPIRED", "expired":
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_EXPIRED
	default:
		return approvalv2.ApprovalStatus_APPROVAL_STATUS_PENDING
	}
}

func (h *ApprovalHandler) convertStringToProtoAction(action string) approvalv2.ActionType {
	switch action {
	case "APPROVE":
		return approvalv2.ActionType_ACTION_TYPE_APPROVE
	case "REJECT":
		return approvalv2.ActionType_ACTION_TYPE_REJECT
	case "DELEGATE":
		return approvalv2.ActionType_ACTION_TYPE_DELEGATE
	case "REQUEST_INFO":
		return approvalv2.ActionType_ACTION_TYPE_REQUEST_INFO
	case "WITHDRAW":
		return approvalv2.ActionType_ACTION_TYPE_WITHDRAW
	case "REASSIGN":
		return approvalv2.ActionType_ACTION_TYPE_REASSIGN
	default:
		return approvalv2.ActionType_ACTION_TYPE_APPROVE
	}
}

func (h *ApprovalHandler) convertProtoAnyToMap(any *anypb.Any) map[string]interface{} {
	if any == nil {
		return make(map[string]interface{})
	}
	// TODO: Implement proper Any to map conversion
	return make(map[string]interface{})
}

func (h *ApprovalHandler) convertMapToProtoAny(data map[string]interface{}) *anypb.Any {
	if len(data) == 0 {
		return nil
	}
	// TODO: Implement proper map to Any conversion
	return nil
}

func (h *ApprovalHandler) parseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func (h *ApprovalHandler) parseUUIDPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func (h *ApprovalHandler) stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *ApprovalHandler) convertInstanceToProto(instance *db.FormInstance) *approvalv2.ApprovalInstance {
	if instance == nil {
		return nil
	}

	return &approvalv2.ApprovalInstance{
		Id:           instance.ID.String(),
		FormId:       instance.FormID.String(),
		EntityType:   "form", // TODO: Map from form metadata
		EntityId:     instance.ID.String(),
		RequestedBy:  instance.CreatedBy,
		AssignedTo:   func() string { if instance.AssignedTo != nil { return *instance.AssignedTo }; return "" }(),
		Status:       h.convertStringToProtoStatus(instance.CurrentState),
		Priority:     approvalv2.Priority_PRIORITY_MEDIUM, // TODO: Map from metadata
		CurrentStep:  instance.CurrentState,
		RequestData:  h.convertMapToProtoAny(nil), // TODO: Convert field_values
		Comments:     "",                          // TODO: Get from latest action
		CreatedAt:    timestamppb.New(instance.CreatedAt),
		UpdatedAt:    timestamppb.New(instance.UpdatedAt),
		Metadata:     h.convertMapToProtoAny(nil), // TODO: Convert metadata
	}
}

func (h *ApprovalHandler) convertActionToProto(action *db.ApprovalAction) *approvalv2.ApprovalAction {
	if action == nil {
		return nil
	}

	return &approvalv2.ApprovalAction{
		Id:             action.ID.String(),
		FormInstanceId: func() string { if action.FormInstanceID.Valid { return uuid.UUID(action.FormInstanceID.Bytes).String() }; return "" }(),
		ApproverId:     action.ApproverID,
		Action:         h.convertStringToProtoAction(action.Action),
		Comments:       func() string { if action.Comments != nil { return *action.Comments }; return "" }(),
		ActedAt:        timestamppb.New(action.ActedAt),
		IpAddress:      func() string { if action.IpAddress != nil { return action.IpAddress.String() }; return "" }(),
		UserAgent:      func() string { if action.UserAgent != nil { return *action.UserAgent }; return "" }(),
		DelegatedFrom:  func() string { if action.DelegatedFrom != nil { return *action.DelegatedFrom }; return "" }(),
		AttachmentUrls: []string{}, // TODO: Unmarshal from JSON
		Metadata:       h.convertMapToProtoAny(nil), // TODO: Convert metadata
		CreatedAt:      timestamppb.New(action.CreatedAt),
	}
}

func (h *ApprovalHandler) convertDelegateToProto(delegate *db.ApprovalDelegate) *approvalv2.ApprovalDelegate {
	if delegate == nil {
		return nil
	}

	return &approvalv2.ApprovalDelegate{
		Id:          delegate.ID.String(),
		DelegatorId: delegate.DelegatorID,
		DelegateId:  delegate.DelegateID,
		EntityTypes: []string{}, // TODO: Unmarshal from JSON
		StartDate:   timestamppb.New(delegate.StartDate),
		EndDate:     func() *timestamppb.Timestamp { if delegate.EndDate.Valid { return timestamppb.New(delegate.EndDate.Time) }; return nil }(),
		IsActive:    func() bool { if delegate.IsActive != nil { return *delegate.IsActive }; return false }(),
		Reason:      func() string { if delegate.Reason != nil { return *delegate.Reason }; return "" }(),
		CreatedAt:   timestamppb.New(delegate.CreatedAt),
		UpdatedAt:   timestamppb.New(delegate.UpdatedAt),
		Metadata:    h.convertMapToProtoAny(nil), // TODO: Convert metadata
	}
}

func (h *ApprovalHandler) convertProtoToCreateDelegateRequest(proto *approvalv2.ApprovalDelegate) *services.CreateDelegateRequest {
	if proto == nil {
		return nil
	}

	return &services.CreateDelegateRequest{
		DelegatorID: proto.DelegatorId,
		DelegateID:  proto.DelegateId,
		EntityTypes: proto.EntityTypes,
		StartDate:   func() time.Time { if proto.StartDate != nil { return proto.StartDate.AsTime() }; return time.Now() }(),
		EndDate:     func() *time.Time { if proto.EndDate != nil { t := proto.EndDate.AsTime(); return &t }; return nil }(),
		Reason:      proto.Reason,
		Metadata:    h.convertProtoAnyToMap(proto.Metadata),
	}
}