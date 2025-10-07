package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	db "p9e.in/ugcl/organization/db/generated"
	"p9e.in/ugcl/organization/mappers"
	"p9e.in/ugcl/organization/repository"
)

type DepartmentService struct {
	deptRepo repository.IDepartmentRepository
}

func NewDepartmentService(deptRepo repository.IDepartmentRepository) IDepartmentService {
	return &DepartmentService{deptRepo: deptRepo}
}

func (s *DepartmentService) CreateDepartment(ctx context.Context, req *pb.CreateDepartmentRequest) (*pb.Department, error) {
	if err := validateCreateDepartmentRequest(req); err != nil {
		return nil, err
	}

	tenantID, _ := uuid.Parse(req.TenantId)
	existing, _ := s.deptRepo.GetByCode(ctx, tenantID, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("department with code '%s' already exists", req.Code)
	}

	params, err := mappers.CreateDepartmentProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	dbDept, err := s.deptRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return mappers.DepartmentDBToProto(dbDept), nil
}

func (s *DepartmentService) UpdateDepartment(ctx context.Context, req *pb.UpdateDepartmentRequest) (*pb.Department, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("id is required")
	}

	params, err := mappers.UpdateDepartmentProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	dbDept, err := s.deptRepo.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return mappers.DepartmentDBToProto(dbDept), nil
}

func (s *DepartmentService) GetDepartment(ctx context.Context, req *pb.GetDepartmentRequest) (*pb.Department, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID: %w", err)
	}

	dbDept, err := s.deptRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	return mappers.DepartmentDBToProto(dbDept), nil
}

func (s *DepartmentService) GetDepartmentByCode(ctx context.Context, req *pb.GetDepartmentByCodeRequest) (*pb.Department, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	dbDept, err := s.deptRepo.GetByCode(ctx, tenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	return mappers.DepartmentDBToProto(dbDept), nil
}

func (s *DepartmentService) ListDepartments(ctx context.Context, req *pb.ListDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.ListDepartmentsParams{
		TenantID: tenantID,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbDepts, err := s.deptRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	countParams := db.CountDepartmentsParams{TenantID: tenantID}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		countParams.IsActive = &isActive
	}
	totalCount, err := s.deptRepo.Count(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count departments: %w", err)
	}

	departments := make([]*pb.Department, len(dbDepts))
	for i, dbDept := range dbDepts {
		departments[i] = mappers.DepartmentDBToProto(dbDept)
	}

	return &pb.ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  totalCount,
	}, nil
}

func (s *DepartmentService) ListDepartmentsByDivision(ctx context.Context, req *pb.ListDepartmentsByDivisionRequest) (*pb.ListDepartmentsResponse, error) {
	divisionID, _ := uuid.Parse(req.DivisionId)
	params := db.ListDepartmentsByDivisionParams{
		DivisionID: uuid.NullUUID{UUID: divisionID, Valid: true},
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbDepts, err := s.deptRepo.ListByDivision(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	departments := make([]*pb.Department, len(dbDepts))
	for i, dbDept := range dbDepts {
		departments[i] = mappers.DepartmentDBToProto(dbDept)
	}

	return &pb.ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  int64(len(departments)),
	}, nil
}

func (s *DepartmentService) ListBusinessLevelDepartments(ctx context.Context, req *pb.ListBusinessLevelDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.ListBusinessLevelDepartmentsParams{TenantID: tenantID}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbDepts, err := s.deptRepo.ListBusinessLevel(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	departments := make([]*pb.Department, len(dbDepts))
	for i, dbDept := range dbDepts {
		departments[i] = mappers.DepartmentDBToProto(dbDept)
	}

	return &pb.ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  int64(len(departments)),
	}, nil
}

func (s *DepartmentService) ListSubDepartments(ctx context.Context, req *pb.ListSubDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	parentID, _ := uuid.Parse(req.ParentDepartmentId)
	params := db.ListSubDepartmentsParams{
		ParentDepartmentID: uuid.NullUUID{UUID: parentID, Valid: true},
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbDepts, err := s.deptRepo.ListSubDepartments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	departments := make([]*pb.Department, len(dbDepts))
	for i, dbDept := range dbDepts {
		departments[i] = mappers.DepartmentDBToProto(dbDept)
	}

	return &pb.ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  int64(len(departments)),
	}, nil
}

func (s *DepartmentService) GetDepartmentHierarchy(ctx context.Context, req *pb.GetDepartmentHierarchyRequest) (*pb.GetDepartmentHierarchyResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.GetDepartmentHierarchyParams{TenantID: tenantID}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbHierarchy, err := s.deptRepo.GetHierarchy(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy: %w", err)
	}

	departments := make([]*pb.DepartmentHierarchy, len(dbHierarchy))
	for i, dbHier := range dbHierarchy {
		departments[i] = mappers.DepartmentHierarchyDBToProto(dbHier)
	}

	return &pb.GetDepartmentHierarchyResponse{Departments: departments}, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, req *pb.DeleteDepartmentRequest) error {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return fmt.Errorf("invalid department ID: %w", err)
	}

	dept, err := s.deptRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return fmt.Errorf("department not found: %w", err)
	}

	// Check for sub-departments
	subDepts, _ := s.deptRepo.ListSubDepartments(ctx, db.ListSubDepartmentsParams{
		ParentDepartmentID: uuid.NullUUID{UUID: id, Valid: true},
	})
	if len(subDepts) > 0 {
		return fmt.Errorf("cannot delete department with %d sub-departments", len(subDepts))
	}

	if err := s.deptRepo.Delete(ctx, id, dept.TenantID, &req.DeletedBy); err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	return nil
}

func validateCreateDepartmentRequest(req *pb.CreateDepartmentRequest) error {
	if req.TenantId == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.Code == "" {
		return fmt.Errorf("code is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
