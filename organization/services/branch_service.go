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

type BranchService struct {
	branchRepo repository.IBranchRepository
}

func NewBranchService(branchRepo repository.IBranchRepository) IBranchService {
	return &BranchService{branchRepo: branchRepo}
}

func (s *BranchService) CreateBranch(ctx context.Context, req *pb.CreateBranchRequest) (*pb.Branch, error) {
	if err := validateCreateBranchRequest(req); err != nil {
		return nil, err
	}

	tenantID, _ := uuid.Parse(req.TenantId)
	existing, _ := s.branchRepo.GetByCode(ctx, tenantID, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("branch with code '%s' already exists", req.Code)
	}

	params, err := mappers.CreateBranchProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	dbBranch, err := s.branchRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return mappers.BranchDBToProto(dbBranch), nil
}

func (s *BranchService) UpdateBranch(ctx context.Context, req *pb.UpdateBranchRequest) (*pb.Branch, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("id is required")
	}

	params, err := mappers.UpdateBranchProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	dbBranch, err := s.branchRepo.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}

	return mappers.BranchDBToProto(dbBranch), nil
}

func (s *BranchService) GetBranch(ctx context.Context, req *pb.GetBranchRequest) (*pb.Branch, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid branch ID: %w", err)
	}

	dbBranch, err := s.branchRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	return mappers.BranchDBToProto(dbBranch), nil
}

func (s *BranchService) GetBranchByCode(ctx context.Context, req *pb.GetBranchByCodeRequest) (*pb.Branch, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	dbBranch, err := s.branchRepo.GetByCode(ctx, tenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	return mappers.BranchDBToProto(dbBranch), nil
}

func (s *BranchService) ListBranches(ctx context.Context, req *pb.ListBranchesRequest) (*pb.ListBranchesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	params := db.ListBranchesParams{
		TenantID: tenantID,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbBranches, err := s.branchRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	countParams := db.CountBranchesParams{TenantID: tenantID}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		countParams.IsActive = &isActive
	}
	totalCount, err := s.branchRepo.Count(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count branches: %w", err)
	}

	branches := make([]*pb.Branch, len(dbBranches))
	for i, dbBranch := range dbBranches {
		branches[i] = mappers.BranchDBToProto(dbBranch)
	}

	return &pb.ListBranchesResponse{
		Branches:   branches,
		TotalCount: totalCount,
	}, nil
}

func (s *BranchService) ListBranchesByDivision(ctx context.Context, req *pb.ListBranchesByDivisionRequest) (*pb.ListBranchesByDivisionResponse, error) {
	divisionID, err := uuid.Parse(req.DivisionId)
	if err != nil {
		return nil, fmt.Errorf("invalid division ID: %w", err)
	}

	params := db.ListBranchesByDivisionParams{
		DivisionID: uuid.NullUUID{UUID: divisionID, Valid: true},
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbBranches, err := s.branchRepo.ListByDivision(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := make([]*pb.Branch, len(dbBranches))
	for i, dbBranch := range dbBranches {
		branches[i] = mappers.BranchDBToProto(dbBranch)
	}

	return &pb.ListBranchesByDivisionResponse{Branches: branches}, nil
}

func (s *BranchService) ListBranchesByCity(ctx context.Context, req *pb.ListBranchesByCityRequest) (*pb.ListBranchesResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.ListBranchesByCityParams{
		TenantID: tenantID,
		City:     &req.City,
	}

	dbBranches, err := s.branchRepo.ListByCity(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := make([]*pb.Branch, len(dbBranches))
	for i, dbBranch := range dbBranches {
		branches[i] = mappers.BranchDBToProto(dbBranch)
	}

	return &pb.ListBranchesResponse{
		Branches:   branches,
		TotalCount: int64(len(branches)),
	}, nil
}

func (s *BranchService) ListBranchesByState(ctx context.Context, req *pb.ListBranchesByStateRequest) (*pb.ListBranchesResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.ListBranchesByStateParams{
		TenantID: tenantID,
		State:    &req.State,
	}

	dbBranches, err := s.branchRepo.ListByState(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := make([]*pb.Branch, len(dbBranches))
	for i, dbBranch := range dbBranches {
		branches[i] = mappers.BranchDBToProto(dbBranch)
	}

	return &pb.ListBranchesResponse{
		Branches:   branches,
		TotalCount: int64(len(branches)),
	}, nil
}

func (s *BranchService) ListBranchSummaries(ctx context.Context, req *pb.ListBranchSummariesRequest) (*pb.ListBranchSummariesResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	params := db.GetBranchSummaryParams{TenantID: tenantID}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbSummaries, err := s.branchRepo.GetSummary(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get summaries: %w", err)
	}

	summaries := make([]*pb.BranchSummary, len(dbSummaries))
	for i, dbSummary := range dbSummaries {
		summaries[i] = mappers.BranchSummaryDBToProto(dbSummary)
	}

	return &pb.ListBranchSummariesResponse{Summaries: summaries}, nil
}

func (s *BranchService) DeleteBranch(ctx context.Context, req *pb.DeleteBranchRequest) error {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return fmt.Errorf("invalid branch ID: %w", err)
	}

	branch, err := s.branchRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	if err := s.branchRepo.Delete(ctx, id, branch.TenantID, &req.DeletedBy); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

func validateCreateBranchRequest(req *pb.CreateBranchRequest) error {
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
