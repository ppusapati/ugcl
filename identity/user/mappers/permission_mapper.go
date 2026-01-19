package mappers

import (
	pb "p9e.in/ugcl/identity/user/api/v2/permission"
	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/models"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func PermissionProtoToModel(pb *pb.Permission) *models.Permission {
	return &models.Permission{
		Namespace: pb.Namespace,
		Resource:  pb.Resource,
		Action:    pb.Action,
		Subject:   pb.Subject,
		TenantID:  pb.TenantId,
		Effect:    models.Effect(pb.Effect),
	}
}

func PermissionModelToProto(m *models.Permission) *pb.Permission {
	if m == nil {
		return nil
	}

	perm := &pb.Permission{
		Namespace: m.Namespace,
		Resource:  m.Resource,
		Action:    m.Action,
		Subject:   m.Subject,
		TenantId:  m.TenantID,
		Effect:    pb.Effect(m.Effect),
		DefName:   m.DefName,
	}

	// Add organizational scope if present
	if m.DivisionID != nil && *m.DivisionID != "" {
		perm.DivisionId = wrapperspb.String(*m.DivisionID)
	}
	if m.BranchID != nil && *m.BranchID != "" {
		perm.BranchId = wrapperspb.String(*m.BranchID)
	}
	if m.DepartmentID != nil && *m.DepartmentID != "" {
		perm.DepartmentId = wrapperspb.String(*m.DepartmentID)
	}

	// Add resource instance if present
	if m.ResourceID != nil && *m.ResourceID != "" {
		perm.ResourceId = wrapperspb.String(*m.ResourceID)
	}

	// Add temporal constraints if present
	if m.ValidFrom != nil {
		perm.ValidFrom = timestamppb.New(*m.ValidFrom)
	}
	if m.ValidUntil != nil {
		perm.ValidUntil = timestamppb.New(*m.ValidUntil)
	}

	// Add inheritance flag
	if m.AllowInheritance != nil {
		perm.AllowInheritance = *m.AllowInheritance
	}

	return perm
}

func PermissionModelToSQLC(m *models.Permission) sqlc.Permission {
	return sqlc.Permission{
		Namespace: m.Namespace,
		Resource:  m.Resource,
		Action:    m.Action,
		Subject:   m.Subject,
		Effect:    ModelToDBEffectInterface(m.Effect),
		// TenantID:  sql.NullString{String: m.TenantID, Valid: m.TenantID != ""},
	}
}

func PermissionModelToGrantSQLC(m *models.Permission) sqlc.GrantPermissionParams {
	return sqlc.GrantPermissionParams{
		Namespace: m.Namespace,
		Resource:  m.Resource,
		Action:    m.Action,
		Subject:   m.Subject,
		Effect:    ModelToDBEffectInterface(m.Effect),
		// TenantID:  sql.NullString{String: m.TenantID, Valid: m.TenantID != ""},
	}
}

func PermissionSQLCToModel(row sqlc.Permission) *models.Permission {
	return &models.Permission{
		Namespace: row.Namespace,
		Resource:  row.Resource,
		Action:    row.Action,
		Subject:   row.Subject,
		Effect:    DBInterfaceToModelEffect(row.Effect),
	}
}

func PermissionsToProtoList(permissions []*models.Permission) []*pb.Permission {
	result := make([]*pb.Permission, 0, len(permissions))
	for _, p := range permissions {
		result = append(result, PermissionModelToProto(p))
	}
	return result
}

func PermissionSQLCRowsToModels(rows []sqlc.Permission) []*models.Permission {
	result := make([]*models.Permission, 0, len(rows))
	for _, row := range rows {
		result = append(result, PermissionSQLCToModel(row))
	}
	return result
}

func MaterializeDefToPermission(def *models.PermissionDef, subject string, resource string, tenantID *string) *models.Permission {
	return &models.Permission{
		Namespace: def.Namespace,
		Resource:  resource, // provided dynamically
		Action:    def.Action,
		Subject:   subject, // role ID or user ID
		Effect:    1,       // assume 1 = allow; change if needed
		TenantID:  *tenantID,
	}
}

// db type ↔ proto enum helpers
func dbToProtoEffect(dbVal string) pb.Effect {
	switch dbVal {
	case "grant":
		return pb.Effect_GRANT
	case "forbidden":
		return pb.Effect_FORBIDDEN
	default:
		return pb.Effect_UNKNOWN
	}
}

func protoToDBEffect(e pb.Effect) string {
	switch e {
	case pb.Effect_GRANT:
		return "grant"
	case pb.Effect_FORBIDDEN:
		return "forbidden"
	default:
		return "unknown"
	}
}

func DBInterfaceToModelEffect(e interface{}) models.Effect {
	if e == nil {
		return models.EffectUnknown
	}
	switch str := e.(type) {
	case string:
		switch str {
		case "grant":
			return models.EffectGrant
		case "forbidden":
			return models.EffectForbidden
		default:
			return models.EffectUnknown
		}
	default:
		return models.EffectUnknown
	}
}

func ModelToDBEffectInterface(e models.Effect) interface{} {
	switch e {
	case models.EffectGrant:
		return "grant"
	case models.EffectForbidden:
		return "forbidden"
	default:
		return "unknown"
	}
}
