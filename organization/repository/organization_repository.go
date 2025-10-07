package repository

import (
	"context"

	db "p9e.in/ugcl/organization/db/generated"
)

type OrganizationRepository struct {
	queries *db.Queries
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(queries *db.Queries) IOrganizationRepository {
	return &OrganizationRepository{
		queries: queries,
	}
}

func (r *OrganizationRepository) GetFullHierarchy(ctx context.Context) ([]*db.OrganizationFullHierarchy, error) {
	return r.queries.GetOrganizationFullHierarchy(ctx)
}
