package repository

import (
	"context"

	db "p9e.in/ugcl/organization/db/generated"
)

type BranchRepository struct {
	queries *db.Queries
}

func NewBranchRepository(queries *db.Queries) IBranchRepository {
	return &BranchRepository{queries: queries}
}

func (r *BranchRepository) Create(ctx context.Context, arg db.CreateBranchParams) (*db.Branch, error) {
	return r.queries.CreateBranch(ctx, arg)
}

func (r *BranchRepository) Update(ctx context.Context, arg db.UpdateBranchParams) (*db.Branch, error) {
	return r.queries.UpdateBranch(ctx, arg)
}

func (r *BranchRepository) GetByID(ctx context.Context, arg db.GetBranchParams) (*db.Branch, error) {
	return r.queries.GetBranch(ctx, arg)
}

func (r *BranchRepository) GetByCode(ctx context.Context, arg db.GetBranchByCodeParams) (*db.Branch, error) {
	return r.queries.GetBranchByCode(ctx, arg)
}

func (r *BranchRepository) List(ctx context.Context, arg db.ListBranchesParams) ([]*db.Branch, error) {
	return r.queries.ListBranches(ctx, arg)
}

func (r *BranchRepository) ListByDivision(ctx context.Context, arg db.ListBranchesByDivisionParams) ([]*db.Branch, error) {
	return r.queries.ListBranchesByDivision(ctx, arg)
}

func (r *BranchRepository) ListByCity(ctx context.Context, arg db.ListBranchesByCityParams) ([]*db.Branch, error) {
	return r.queries.ListBranchesByCity(ctx, arg)
}

func (r *BranchRepository) ListByState(ctx context.Context, arg db.ListBranchesByStateParams) ([]*db.Branch, error) {
	return r.queries.ListBranchesByState(ctx, arg)
}

func (r *BranchRepository) Count(ctx context.Context, arg db.CountBranchesParams) (int64, error) {
	return r.queries.CountBranches(ctx, arg)
}

func (r *BranchRepository) CountByDivision(ctx context.Context, arg db.CountBranchesByDivisionParams) (int64, error) {
	return r.queries.CountBranchesByDivision(ctx, arg)
}

func (r *BranchRepository) GetSummary(ctx context.Context, arg db.GetBranchSummaryParams) ([]*db.BranchSummary, error) {
	return r.queries.GetBranchSummary(ctx, arg)
}

func (r *BranchRepository) GetSummaryByDivision(ctx context.Context, arg db.GetBranchSummaryByDivisionParams) ([]*db.BranchSummary, error) {
	return r.queries.GetBranchSummaryByDivision(ctx, arg)
}

func (r *BranchRepository) Delete(ctx context.Context, arg db.DeleteBranchParams) error {
	return r.queries.DeleteBranch(ctx, arg)
}

func (r *BranchRepository) HardDelete(ctx context.Context, arg db.HardDeleteBranchParams) error {
	return r.queries.HardDeleteBranch(ctx, arg)
}
