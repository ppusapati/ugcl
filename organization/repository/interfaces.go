package repository

import (
	"context"

	db "p9e.in/ugcl/organization/db/generated"
)

// IDivisionRepository defines the interface for division data operations
type IDivisionRepository interface {
	Create(ctx context.Context, arg db.CreateDivisionParams) (*db.Division, error)
	Update(ctx context.Context, arg db.UpdateDivisionParams) (*db.Division, error)
	GetByID(ctx context.Context, arg db.GetDivisionParams) (*db.Division, error)
	GetByCode(ctx context.Context, arg db.GetDivisionByCodeParams) (*db.Division, error)
	List(ctx context.Context, arg db.ListDivisionsParams) ([]*db.Division, error)
	Count(ctx context.Context, arg db.CountDivisionsParams) (int64, error)
	GetSummary(ctx context.Context, arg db.GetDivisionSummaryParams) ([]*db.DivisionSummary, error)
	Delete(ctx context.Context, arg db.DeleteDivisionParams) error
	HardDelete(ctx context.Context, arg db.HardDeleteDivisionParams) error
}

// IBranchRepository defines the interface for branch data operations
type IBranchRepository interface {
	Create(ctx context.Context, arg db.CreateBranchParams) (*db.Branch, error)
	Update(ctx context.Context, arg db.UpdateBranchParams) (*db.Branch, error)
	GetByID(ctx context.Context, arg db.GetBranchParams) (*db.Branch, error)
	GetByCode(ctx context.Context, arg db.GetBranchByCodeParams) (*db.Branch, error)
	List(ctx context.Context, arg db.ListBranchesParams) ([]*db.Branch, error)
	ListByDivision(ctx context.Context, arg db.ListBranchesByDivisionParams) ([]*db.Branch, error)
	ListByCity(ctx context.Context, arg db.ListBranchesByCityParams) ([]*db.Branch, error)
	ListByState(ctx context.Context, arg db.ListBranchesByStateParams) ([]*db.Branch, error)
	Count(ctx context.Context, arg db.CountBranchesParams) (int64, error)
	CountByDivision(ctx context.Context, arg db.CountBranchesByDivisionParams) (int64, error)
	GetSummary(ctx context.Context, arg db.GetBranchSummaryParams) ([]*db.BranchSummary, error)
	GetSummaryByDivision(ctx context.Context, arg db.GetBranchSummaryByDivisionParams) ([]*db.BranchSummary, error)
	Delete(ctx context.Context, arg db.DeleteBranchParams) error
	HardDelete(ctx context.Context, arg db.HardDeleteBranchParams) error
}

// IDepartmentRepository defines the interface for department data operations
type IDepartmentRepository interface {
	Create(ctx context.Context, arg db.CreateDepartmentParams) (*db.Department, error)
	Update(ctx context.Context, arg db.UpdateDepartmentParams) (*db.Department, error)
	GetByID(ctx context.Context, arg db.GetDepartmentParams) (*db.Department, error)
	GetByCode(ctx context.Context, arg db.GetDepartmentByCodeParams) (*db.Department, error)
	List(ctx context.Context, arg db.ListDepartmentsParams) ([]*db.Department, error)
	ListByDivision(ctx context.Context, arg db.ListDepartmentsByDivisionParams) ([]*db.Department, error)
	ListBusinessLevel(ctx context.Context, arg db.ListBusinessLevelDepartmentsParams) ([]*db.Department, error)
	ListSubDepartments(ctx context.Context, arg db.ListSubDepartmentsParams) ([]*db.Department, error)
	GetHierarchy(ctx context.Context, arg db.GetDepartmentHierarchyParams) ([]*db.DepartmentHierarchy, error)
	Count(ctx context.Context, arg db.CountDepartmentsParams) (int64, error)
	CountByDivision(ctx context.Context, arg db.CountDepartmentsByDivisionParams) (int64, error)
	Delete(ctx context.Context, arg db.DeleteDepartmentParams) error
	HardDelete(ctx context.Context, arg db.HardDeleteDepartmentParams) error
}

// IOrganizationRepository provides access to full organizational hierarchy
type IOrganizationRepository interface {
	GetFullHierarchy(ctx context.Context) ([]*db.OrganizationFullHierarchy, error)
}
