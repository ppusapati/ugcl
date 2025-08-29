// =============================================================================
// internal/helpers/mappers/instance_mapper.go
// =============================================================================
package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/formbuilder/api/v2/form_instance"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// ProtoToFormInstance converts protobuf FormInstance to SQLC FormInstance struct
func ProtoToFormInstance(proto *pb.FormInstance) (*db.FormInstance, error) {
	if proto == nil {
		return nil, fmt.Errorf("invalid form instance")
	}

	// Convert field values
	fieldValues, err := ProtoAnyToMapStringInterface(proto.FieldValues)
	if err != nil {
		return nil, fmt.Errorf("failed to convert field values: %w", err)
	}

	fieldValuesJSON, err := json.Marshal(fieldValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field values: %w", err)
	}

	metadataJSON, err := json.Marshal(proto.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Parse UUIDs
	id, err := uuid.Parse(proto.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid instance ID: %w", err)
	}

	formID, err := uuid.Parse(proto.FormId)
	if err != nil {
		return nil, fmt.Errorf("invalid form ID: %w", err)
	}

	return &db.FormInstance{
		ID:           id,
		FormID:       formID,
		CurrentState: proto.CurrentState,
		FieldValues:  fieldValuesJSON,
		CreatedBy:    proto.CreatedBy,
		AssignedTo:   stringPtr(proto.AssignedTo),
		Metadata:     metadataJSON,
	}, nil
}

// FormInstanceToProto converts SQLC FormInstance struct to protobuf FormInstance
func FormInstanceToProto(instance *db.FormInstance) (*pb.FormInstance, error) {
	if instance == nil {
		return nil, fmt.Errorf("invalid form instance model")
	}

	// Convert field values
	var fieldValues map[string]interface{}
	if err := json.Unmarshal(instance.FieldValues, &fieldValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field values: %w", err)
	}

	protoFieldValues, err := MapStringInterfaceToProtoAny(fieldValues)
	if err != nil {
		return nil, fmt.Errorf("failed to convert field values: %w", err)
	}

	var metadata map[string]string
	if err := json.Unmarshal(instance.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &pb.FormInstance{
		Id:           instance.ID.String(),
		FormId:       instance.FormID.String(),
		CurrentState: instance.CurrentState,
		FieldValues:  protoFieldValues,
		CreatedAt:    timestamppb.New(instance.CreatedAt),
		UpdatedAt:    timestamppb.New(instance.UpdatedAt),
		CreatedBy:    instance.CreatedBy,
		AssignedTo:   stringValue(instance.AssignedTo),
		Metadata:     metadata,
	}, nil
}

// AuditLogToProto converts SQLC AuditLog struct to protobuf AuditLog
func AuditLogToProto(log *db.AuditLog) (*pb.AuditLog, error) {
	if log == nil {
		return nil, fmt.Errorf("invalid audit log model")
	}

	// Convert changes
	var changes map[string]interface{}
	if err := json.Unmarshal(log.Changes, &changes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
	}

	protoChanges, err := MapStringInterfaceToProtoAny(changes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert changes: %w", err)
	}

	var ipAddress *string
	if log.IpAddress != nil {
		ip := log.IpAddress.String()
		ipAddress = &ip
	}

	return &pb.AuditLog{
		Id:         log.ID.String(),
		InstanceId: log.InstanceID.String(),
		UserId:     log.UserID,
		Action:     log.Action,
		FromState:  stringValue(log.FromState),
		ToState:    stringValue(log.ToState),
		Changes:    protoChanges,
		Timestamp:  timestamppb.New(log.Timestamp),
		IpAddress:  *ipAddress,
		UserAgent:  stringValue(log.UserAgent),
	}, nil
}

// AttachmentToProto converts SQLC Attachment struct to protobuf Attachment
func AttachmentToProto(attachment *db.Attachment) (*pb.Attachment, error) {
	if attachment == nil {
		return nil, fmt.Errorf("invalid attachment model")
	}

	return &pb.Attachment{
		Id:          attachment.ID.String(),
		InstanceId:  attachment.InstanceID.String(),
		FieldId:     attachment.FieldID,
		Filename:    attachment.Filename,
		MimeType:    stringValue(attachment.MimeType),
		Size:        int64Value(attachment.Size),
		StoragePath: attachment.StoragePath,
		UploadedAt:  timestamppb.New(attachment.UploadedAt),
		UploadedBy:  attachment.UploadedBy,
	}, nil
}

// CommentToProto converts SQLC Comment struct to protobuf Comment
func CommentToProto(comment *db.Comment) (*pb.Comment, error) {
	if comment == nil {
		return nil, fmt.Errorf("invalid comment model")
	}

	return &pb.Comment{
		Id:         comment.ID.String(),
		InstanceId: comment.InstanceID.String(),
		UserId:     comment.UserID,
		Text:       comment.Text,
		CreatedAt:  timestamppb.New(comment.CreatedAt),
		Internal:   boolValue(comment.Internal),
	}, nil
}

func int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
