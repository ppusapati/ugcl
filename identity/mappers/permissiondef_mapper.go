package mappers

import (
	"database/sql"
	"time"

	pb "p9e.in/ugcl/identity/api/v2/permissiondef"
	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/models"
)

/*─────────────────────────────────────────────────────────────
  Helpers
─────────────────────────────────────────────────────────────*/

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

/*─────────────────────────────────────────────────────────────
  Proto  ⇄  Model
─────────────────────────────────────────────────────────────*/

// proto → model
func PermissionDefProtoToModel(src *pb.PermissionDef) *models.PermissionDef {
	if src == nil {
		return nil
	}
	return &models.PermissionDef{
		Name:        src.Name,
		Namespace:   src.Namespace,
		Resource:    src.Resource,
		Action:      src.Action,
		Scope:       src.Scope,
		Version:     src.Version,
		TenantID:    strPtr(src.TenantId),
		Description: strPtr(src.Description),
		Side:        models.PermissionSide(src.Side),
		// CreatedAt / UpdatedAt are DB-generated → zero values here
	}
}

// model → proto
func PermissionDefModelToProto(src *models.PermissionDef) *pb.PermissionDef {
	if src == nil {
		return nil
	}
	return &pb.PermissionDef{
		Name:        src.Name,
		Namespace:   src.Namespace,
		Resource:    src.Resource,
		Action:      src.Action,
		Scope:       src.Scope,
		Version:     src.Version,
		TenantId:    deref(src.TenantID),
		Description: deref(src.Description),
		Side:        pb.PermissionSide(src.Side),
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

/*─────────────────────────────────────────────────────────────
  Model  ⇄  sqlc params / rows
─────────────────────────────────────────────────────────────*/

// model → sqlc INSERT params
func PermissionDefModelToSQLC(src *models.PermissionDef) sqlc.CreatePermissionDefParams {
	side := int32(src.Side)
	return sqlc.CreatePermissionDefParams{
		Name:        src.Name,
		Description: src.Description,
		Side:        &side,
		Namespace:   src.Namespace,
		Resource:    src.Resource,
		Action:      src.Action,
		Scope:       src.Scope,
		Version:     src.Version,
	}
}

// model → sqlc UPDATE params
func PermissionDefModelToSQLCUpdate(src *models.PermissionDef) sqlc.UpdatePermissionDefParams {
	side := int32(src.Side)
	return sqlc.UpdatePermissionDefParams{
		Name:      src.Name,
		Side:      &side,
		Namespace: src.Namespace,
		Resource:  src.Resource,
		Action:    src.Action,
		Scope:     src.Scope,
	}
}

// sqlc row → model
func PermissionDefSQLCToModel(row sqlc.PermissionDef) *models.PermissionDef {
	return &models.PermissionDef{
		Name:        row.Name,
		Namespace:   row.Namespace,
		Resource:    row.Resource,
		Action:      row.Action,
		Scope:       row.Scope,
		Version:     row.Version,
		Description: row.Description,
		TenantID:    nil, // single-tenant for now (no column)
		Side:        models.PermissionSide(*row.Side),
		CreatedAt:   time.Time(row.CreatedAt.Time),
		UpdatedAt:   time.Time(row.UpdatedAt.Time),
	}
}

func toPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

/*─────────────────────────────────────────────────────────────
  Bulk helpers
─────────────────────────────────────────────────────────────*/

func PermissionDefsToProtoList(list []*models.PermissionDef) []*pb.PermissionDef {
	out := make([]*pb.PermissionDef, 0, len(list))
	for _, m := range list {
		out = append(out, PermissionDefModelToProto(m))
	}
	return out
}

func PermissionDefsSQLCRowsToModels(rows []sqlc.PermissionDef) []*models.PermissionDef {
	out := make([]*models.PermissionDef, 0, len(rows))
	for _, r := range rows {
		out = append(out, PermissionDefSQLCToModel(r))
	}
	return out
}
