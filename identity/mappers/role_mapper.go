package mappers

import (
	"encoding/json"

	"p9e.in/ugcl/packages/converters"

	pb "p9e.in/ugcl/identity/api/v2/role"
	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/models"
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
	return &sqlc.CreateRoleParams{
		ID:          role.ID,
		Name:        role.Name,
		ParentID:    &role.ParentID,
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

	return &models.Role{
		ID:          role.ID,
		Name:        role.Name,
		ParentID:    *role.ParentID,
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
