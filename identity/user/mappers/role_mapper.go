package mappers

import (
	"encoding/json"
	"fmt"

	"p9e.in/ugcl/packages/converters"

	pb "p9e.in/ugcl/identity/user/api/v2/role"
	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/models"
)

func RoleProtoToModel(pbRole *pb.Role) *models.Role {
	if pbRole == nil {
		return nil
	}

	return &models.Role{
		ID:          pbRole.Id,
		Name:        pbRole.Name,
		ParentID:    pbRole.ParentId,
		Metadata:    pbRole.Metadata.AsMap(),
		IsPreserved: pbRole.IsPreserved,
	}
}

func RoleModelToProto(role *models.Role) *pb.Role {
	if role == nil {
		return nil
	}

	pbRole := &pb.Role{
		Id:          role.ID,
		Name:        role.Name,
		IsPreserved: role.IsPreserved,
	}

	if role.ParentID != "" {
		pbRole.ParentId = role.ParentID
	}

	if role.Metadata != nil {
		pbRole.Metadata = converters.MapToStructPB(role.Metadata)
	}

	return pbRole
}

func RoleModelToSQLC(role *models.Role) (*sqlc.CreateRoleParams, error) {
	if role == nil {
		return nil, nil
	}

	var metadata []byte
	var err error
	if role.Metadata != nil {
		metadata, err = json.Marshal(role.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// Convert string ID to int64 - for now use hash or parse
	var id int64
	if role.ID != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(role.ID) {
			id = id*31 + int64(b)
		}
	}

	var parentID *int64
	if role.ParentID != "" {
		var pID int64
		for _, b := range []byte(role.ParentID) {
			pID = pID*31 + int64(b)
		}
		parentID = &pID
	}

	return &sqlc.CreateRoleParams{
		ID:          id,
		Name:        role.Name,
		ParentID:    parentID,
		IsPreserved: role.IsPreserved,
		Metadata:    metadata,
	}, nil
}

func RoleSQLCToModel(role sqlc.Role) (*models.Role, error) {
	var metadata map[string]interface{}
	if len(role.Metadata) > 0 {
		if err := json.Unmarshal(role.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	// Convert int64 ID back to string using UUID or use the UUID field
	var id string
	if role.Uuid.String() != "" {
		id = role.Uuid.String()
	} else {
		// Fallback to converting int64 to string
		id = fmt.Sprintf("%d", role.ID)
	}

	var parentID string
	if role.ParentID != nil {
		parentID = fmt.Sprintf("%d", *role.ParentID)
	}

	return &models.Role{
		ID:          id,
		Name:        role.Name,
		ParentID:    parentID,
		Metadata:    metadata,
		IsPreserved: role.IsPreserved,
	}, nil
}

func RolesToProtoList(roles []*models.Role) ([]*pb.Role, error) {
	result := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		pbRole := RoleModelToProto(role)
		if pbRole == nil {
			continue
		}
		result = append(result, pbRole)
	}
	return result, nil
}
