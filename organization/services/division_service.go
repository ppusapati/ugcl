package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	db "p9e.in/ugcl/organization/db/generated"
	"p9e.in/ugcl/organization/mappers"
	"p9e.in/ugcl/organization/repository"
)

type DivisionService struct {
	divisionRepo repository.IDivisionRepository
	branchRepo   repository.IBranchRepository
}

// NewDivisionService creates a new division service
func NewDivisionService(
	divisionRepo repository.IDivisionRepository,
	branchRepo repository.IBranchRepository,
) IDivisionService {
	return &DivisionService{
		divisionRepo: divisionRepo,
		branchRepo:   branchRepo,
	}
}

func (s *DivisionService) CreateDivision(ctx context.Context, req *pb.CreateDivisionRequest) (*pb.Division, error) {
	// Validate request
	if err := s.validateCreateDivisionRequest(req); err != nil {
		return nil, err
	}

	// Check if code already exists
	tenantID, _ := uuid.Parse(req.TenantId)
	existing, _ := s.divisionRepo.GetByCode(ctx, tenantID, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("division with code '%s' already exists", req.Code)
	}

	// Convert proto to DB params
	params, err := mappers.CreateDivisionProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Create division
	dbDivision, err := s.divisionRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create division: %w", err)
	}

	// Convert DB to proto
	return mappers.DivisionDBToProto(dbDivision), nil
}

func (s *DivisionService) UpdateDivision(ctx context.Context, req *pb.UpdateDivisionRequest) (*pb.Division, error) {
	// Validate request
	if err := s.validateUpdateDivisionRequest(req); err != nil {
		return nil, err
	}

	// Check if division exists
	id, _ := uuid.Parse(req.Id)
	// Note: We'd need tenant_id from context or request
	// For now, using zero UUID as placeholder - should be fixed with proper tenant context
	existing, err := s.divisionRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return nil, fmt.Errorf("division not found: %w", err)
	}

	// If code is being changed, check for duplicates
	if req.Code != nil && req.Code.Value != existing.Code {
		existingByCode, _ := s.divisionRepo.GetByCode(ctx, existing.TenantID, req.Code.Value)
		if existingByCode != nil {
			return nil, fmt.Errorf("division with code '%s' already exists", req.Code.Value)
		}
	}

	// Convert proto to DB params
	params, err := mappers.UpdateDivisionProtoToDB(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Update division
	dbDivision, err := s.divisionRepo.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update division: %w", err)
	}

	// Convert DB to proto
	return mappers.DivisionDBToProto(dbDivision), nil
}

func (s *DivisionService) GetDivision(ctx context.Context, req *pb.GetDivisionRequest) (*pb.Division, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid division ID: %w", err)
	}

	// Note: Should get tenant_id from context
	dbDivision, err := s.divisionRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return nil, fmt.Errorf("division not found: %w", err)
	}

	return mappers.DivisionDBToProto(dbDivision), nil
}

func (s *DivisionService) GetDivisionByCode(ctx context.Context, req *pb.GetDivisionByCodeRequest) (*pb.Division, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	dbDivision, err := s.divisionRepo.GetByCode(ctx, tenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("division not found: %w", err)
	}

	return mappers.DivisionDBToProto(dbDivision), nil
}

func (s *DivisionService) ListDivisions(ctx context.Context, req *pb.ListDivisionsRequest) (*pb.ListDivisionsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Build list params
	params := db.ListDivisionsParams{
		TenantID: tenantID,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	// Get divisions
	dbDivisions, err := s.divisionRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list divisions: %w", err)
	}

	// Get total count
	countParams := db.CountDivisionsParams{
		TenantID: tenantID,
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		countParams.IsActive = &isActive
	}
	totalCount, err := s.divisionRepo.Count(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count divisions: %w", err)
	}

	// Convert to proto
	divisions := make([]*pb.Division, len(dbDivisions))
	for i, dbDiv := range dbDivisions {
		divisions[i] = mappers.DivisionDBToProto(dbDiv)
	}

	return &pb.ListDivisionsResponse{
		Divisions:  divisions,
		TotalCount: totalCount,
	}, nil
}

func (s *DivisionService) ListDivisionSummaries(ctx context.Context, req *pb.ListDivisionSummariesRequest) (*pb.ListDivisionSummariesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	params := db.GetDivisionSummaryParams{
		TenantID: tenantID,
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}

	dbSummaries, err := s.divisionRepo.GetSummary(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get division summaries: %w", err)
	}

	// Convert to proto
	summaries := make([]*pb.DivisionSummary, len(dbSummaries))
	for i, dbSummary := range dbSummaries {
		summaries[i] = mappers.DivisionSummaryDBToProto(dbSummary)
	}

	return &pb.ListDivisionSummariesResponse{
		Summaries: summaries,
	}, nil
}

func (s *DivisionService) DeleteDivision(ctx context.Context, req *pb.DeleteDivisionRequest) error {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return fmt.Errorf("invalid division ID: %w", err)
	}

	// Check if division has active branches
	// Note: Should get tenant_id from context
	division, err := s.divisionRepo.GetByID(ctx, id, uuid.UUID{})
	if err != nil {
		return fmt.Errorf("division not found: %w", err)
	}

	branchCount, err := s.branchRepo.CountByDivision(ctx, db.CountBranchesByDivisionParams{
		DivisionID: uuid.NullUUID{UUID: id, Valid: true},
		IsActive:   wrapperspb.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to check branches: %w", err)
	}

	if branchCount > 0 {
		return fmt.Errorf("cannot delete division with %d active branches", branchCount)
	}

	// Soft delete
	if err := s.divisionRepo.Delete(ctx, id, division.TenantID, &req.DeletedBy); err != nil {
		return fmt.Errorf("failed to delete division: %w", err)
	}

	return nil
}

// Validation helpers
func (s *DivisionService) validateCreateDivisionRequest(req *pb.CreateDivisionRequest) error {
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

func (s *DivisionService) validateUpdateDivisionRequest(req *pb.UpdateDivisionRequest) error {
	if req.Id == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}
