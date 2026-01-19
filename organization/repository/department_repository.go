package repository

import (
	"context"

	db "p9e.in/ugcl/organization/db/generated"
)

type DepartmentRepository struct {
	queries *db.Queries
}

func NewDepartmentRepository(queries *db.Queries) IDepartmentRepository {
	return &DepartmentRepository{queries: queries}
}

func (r *DepartmentRepository) Create(ctx context.Context, arg db.CreateDepartmentParams) (*db.Department, error) {
	return r.queries.CreateDepartment(ctx, arg)
}

func (r *DepartmentRepository) Update(ctx context.Context, arg db.UpdateDepartmentParams) (*db.Department, error) {
	return r.queries.UpdateDepartment(ctx, arg)
}

func (r *DepartmentRepository) GetByID(ctx context.Context, arg db.GetDepartmentParams) (*db.Department, error) {
	return r.queries.GetDepartment(ctx, arg)
}

func (r *DepartmentRepository) GetByCode(ctx context.Context, arg db.GetDepartmentByCodeParams) (*db.Department, error) {
	return r.queries.GetDepartmentByCode(ctx, arg)
}

func (r *DepartmentRepository) List(ctx context.Context, arg db.ListDepartmentsParams) ([]*db.Department, error) {
	return r.queries.ListDepartments(ctx, arg)
}

func (r *DepartmentRepository) ListByDivision(ctx context.Context, arg db.ListDepartmentsByDivisionParams) ([]*db.Department, error) {
	return r.queries.ListDepartmentsByDivision(ctx, arg)
}

func (r *DepartmentRepository) ListBusinessLevel(ctx context.Context, arg db.ListBusinessLevelDepartmentsParams) ([]*db.Department, error) {
	return r.queries.ListBusinessLevelDepartments(ctx, arg)
}

func (r *DepartmentRepository) ListSubDepartments(ctx context.Context, arg db.ListSubDepartmentsParams) ([]*db.Department, error) {
	return r.queries.ListSubDepartments(ctx, arg)
}

func (r *DepartmentRepository) GetHierarchy(ctx context.Context, arg db.GetDepartmentHierarchyParams) ([]*db.DepartmentHierarchy, error) {
	return r.queries.GetDepartmentHierarchy(ctx, arg)
}

func (r *DepartmentRepository) Count(ctx context.Context, arg db.CountDepartmentsParams) (int64, error) {
	return r.queries.CountDepartments(ctx, arg)
}

func (r *DepartmentRepository) CountByDivision(ctx context.Context, arg db.CountDepartmentsByDivisionParams) (int64, error) {
	return r.queries.CountDepartmentsByDivision(ctx, arg)
}

func (r *DepartmentRepository) Delete(ctx context.Context, arg db.DeleteDepartmentParams) error {
	return r.queries.DeleteDepartment(ctx, arg)
}

func (r *DepartmentRepository) HardDelete(ctx context.Context, arg db.HardDeleteDepartmentParams) error {
	return r.queries.HardDeleteDepartment(ctx, arg)
}
