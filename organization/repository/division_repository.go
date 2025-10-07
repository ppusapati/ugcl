package repository

import (
	"context"

	db "p9e.in/ugcl/organization/db/generated"
)

type DivisionRepository struct {
	queries *db.Queries
}

func NewDivisionRepository(queries *db.Queries) IDivisionRepository {
	return &DivisionRepository{queries: queries}
}

func (r *DivisionRepository) Create(ctx context.Context, arg db.CreateDivisionParams) (*db.Division, error) {
	return r.queries.CreateDivision(ctx, arg)
}

func (r *DivisionRepository) Update(ctx context.Context, arg db.UpdateDivisionParams) (*db.Division, error) {
	return r.queries.UpdateDivision(ctx, arg)
}

func (r *DivisionRepository) GetByID(ctx context.Context, arg db.GetDivisionParams) (*db.Division, error) {
	return r.queries.GetDivision(ctx, arg)
}

func (r *DivisionRepository) GetByCode(ctx context.Context, arg db.GetDivisionByCodeParams) (*db.Division, error) {
	return r.queries.GetDivisionByCode(ctx, arg)
}

func (r *DivisionRepository) List(ctx context.Context, arg db.ListDivisionsParams) ([]*db.Division, error) {
	return r.queries.ListDivisions(ctx, arg)
}

func (r *DivisionRepository) Count(ctx context.Context, arg db.CountDivisionsParams) (int64, error) {
	return r.queries.CountDivisions(ctx, arg)
}

func (r *DivisionRepository) GetSummary(ctx context.Context, arg db.GetDivisionSummaryParams) ([]*db.DivisionSummary, error) {
	return r.queries.GetDivisionSummary(ctx, arg)
}

func (r *DivisionRepository) Delete(ctx context.Context, arg db.DeleteDivisionParams) error {
	return r.queries.DeleteDivision(ctx, arg)
}

func (r *DivisionRepository) HardDelete(ctx context.Context, arg db.HardDeleteDivisionParams) error {
	return r.queries.HardDeleteDivision(ctx, arg)
}
