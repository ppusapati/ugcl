package services

import (
	"context"

	pb "p9e.in/ugcl/organization/api/v1/organization"
)

// IDivisionService defines the interface for division business operations
type IDivisionService interface {
	CreateDivision(ctx context.Context, req *pb.CreateDivisionRequest) (*pb.Division, error)
	UpdateDivision(ctx context.Context, req *pb.UpdateDivisionRequest) (*pb.Division, error)
	GetDivision(ctx context.Context, req *pb.GetDivisionRequest) (*pb.Division, error)
	GetDivisionByCode(ctx context.Context, req *pb.GetDivisionByCodeRequest) (*pb.Division, error)
	ListDivisions(ctx context.Context, req *pb.ListDivisionsRequest) (*pb.ListDivisionsResponse, error)
	ListDivisionSummaries(ctx context.Context, req *pb.ListDivisionSummariesRequest) (*pb.ListDivisionSummariesResponse, error)
	DeleteDivision(ctx context.Context, req *pb.DeleteDivisionRequest) error
}

// IBranchService defines the interface for branch business operations
type IBranchService interface {
	CreateBranch(ctx context.Context, req *pb.CreateBranchRequest) (*pb.Branch, error)
	UpdateBranch(ctx context.Context, req *pb.UpdateBranchRequest) (*pb.Branch, error)
	GetBranch(ctx context.Context, req *pb.GetBranchRequest) (*pb.Branch, error)
	GetBranchByCode(ctx context.Context, req *pb.GetBranchByCodeRequest) (*pb.Branch, error)
	ListBranches(ctx context.Context, req *pb.ListBranchesRequest) (*pb.ListBranchesResponse, error)
	ListBranchesByDivision(ctx context.Context, req *pb.ListBranchesByDivisionRequest) (*pb.ListBranchesByDivisionResponse, error)
	ListBranchesByCity(ctx context.Context, req *pb.ListBranchesByCityRequest) (*pb.ListBranchesResponse, error)
	ListBranchesByState(ctx context.Context, req *pb.ListBranchesByStateRequest) (*pb.ListBranchesResponse, error)
	ListBranchSummaries(ctx context.Context, req *pb.ListBranchSummariesRequest) (*pb.ListBranchSummariesResponse, error)
	DeleteBranch(ctx context.Context, req *pb.DeleteBranchRequest) error
}

// IDepartmentService defines the interface for department business operations
type IDepartmentService interface {
	CreateDepartment(ctx context.Context, req *pb.CreateDepartmentRequest) (*pb.Department, error)
	UpdateDepartment(ctx context.Context, req *pb.UpdateDepartmentRequest) (*pb.Department, error)
	GetDepartment(ctx context.Context, req *pb.GetDepartmentRequest) (*pb.Department, error)
	GetDepartmentByCode(ctx context.Context, req *pb.GetDepartmentByCodeRequest) (*pb.Department, error)
	ListDepartments(ctx context.Context, req *pb.ListDepartmentsRequest) (*pb.ListDepartmentsResponse, error)
	ListDepartmentsByDivision(ctx context.Context, req *pb.ListDepartmentsByDivisionRequest) (*pb.ListDepartmentsResponse, error)
	ListBusinessLevelDepartments(ctx context.Context, req *pb.ListBusinessLevelDepartmentsRequest) (*pb.ListDepartmentsResponse, error)
	ListSubDepartments(ctx context.Context, req *pb.ListSubDepartmentsRequest) (*pb.ListDepartmentsResponse, error)
	GetDepartmentHierarchy(ctx context.Context, req *pb.GetDepartmentHierarchyRequest) (*pb.GetDepartmentHierarchyResponse, error)
	DeleteDepartment(ctx context.Context, req *pb.DeleteDepartmentRequest) error
}
