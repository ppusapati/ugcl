package services

import (
	"context"
	"fmt"

	userpb "p9e.in/ugcl/identity/api/v2/user"
	usermapper "p9e.in/ugcl/identity/mappers"
	userservices "p9e.in/ugcl/identity/services"
	pb "p9e.in/ugcl/vendors/api/v2/contractor"
	db "p9e.in/ugcl/vendors/db/generated"
	"p9e.in/ugcl/vendors/repository"

	"github.com/google/uuid"
)

// ContractorService defines the business logic for contractor operations
type IContractorService interface {
	Create(ctx context.Context, contractor *db.Contractor, user *userpb.User) (*db.Contractor, error)
	Update(ctx context.Context, contractor *db.Contractor) (*db.Contractor, error)
	GetByID(ctx context.Context, id string) (*db.Contractor, error)
	GetByPersonID(ctx context.Context, personID string) (*db.Contractor, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize, pageOffset int32, filter *pb.ContractorFilter, sort []string) ([]db.Contractor, int64, error)
}

type ContractorService struct {
	repo    repository.IContractorRepository
	userSvc userservices.IUserService
}

// NewContractorService creates a new contractor service
func NewContractorService(repo repository.IContractorRepository, userSvc userservices.IUserService) IContractorService {
	return &ContractorService{
		repo:    repo,
		userSvc: userSvc,
	}
}

// Create implements IContractorService.
func (c *ContractorService) Create(ctx context.Context, contractor *db.Contractor, user *userpb.User) (*db.Contractor, error) {
	// Step 1: Create user first

	userResp, err := c.userSvc.RegisterUser(ctx, usermapper.UserProtoToModel(user))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	var userID string
	// Extract user ID from response
	if userResp.Uuid != "" {
		userID = userResp.Uuid
	} else {
		return nil, fmt.Errorf("user created but has no valid UUID")
	}

	contractor.PersonID = userID
	// con, _ := mappers.ProtoToDBContractor(contractor)
	return c.repo.Create(ctx, db.CreateContractorParams{
		// Id:                contractor.ID,
		CompanyName:       contractor.CompanyName,
		CompanyType:       contractor.CompanyType,
		Gst:               contractor.Gst,
		Pan:               contractor.Pan,
		Category:          contractor.Category,
		PersonID:          contractor.PersonID,
		AssociatedProject: contractor.AssociatedProject,
		WorkingSite:       contractor.WorkingSite,
		ContractStartDate: contractor.ContractStartDate,
		ContractEndDate:   contractor.ContractEndDate,
		Status:            contractor.Status,
		Metadata:          contractor.Metadata,
	})
}

// Other methods simply delegate to repository
func (s *ContractorService) Update(ctx context.Context, contractor *db.Contractor) (*db.Contractor, error) {
	params := db.UpdateContractorParams{
		ID:                contractor.ID,
		CompanyName:       &contractor.CompanyName,
		CompanyType:       contractor.CompanyType,
		Gst:               contractor.Gst,
		Pan:               contractor.Pan,
		Category:          contractor.Category,
		PersonID:          &contractor.PersonID,
		AssociatedProject: contractor.AssociatedProject,
		WorkingSite:       contractor.WorkingSite,
		ContractStartDate: contractor.ContractStartDate,
		ContractEndDate:   contractor.ContractEndDate,
		Status:            contractor.Status,
		Metadata:          contractor.Metadata,
	}

	return s.repo.Update(ctx, params)
}

func (s *ContractorService) GetByID(ctx context.Context, id string) (*db.Contractor, error) {
	contractorID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.GetByID(ctx, contractorID)
}

func (s *ContractorService) GetByPersonID(ctx context.Context, personID string) (*db.Contractor, error) {
	pid, err := uuid.Parse(personID)
	if err != nil {
		return nil, fmt.Errorf("invalid person ID: %w", err)
	}
	return s.repo.GetByPersonID(ctx, pid)
}

func (s *ContractorService) Delete(ctx context.Context, id string) error {
	contractorID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.Delete(ctx, contractorID)
}

func (s *ContractorService) List(ctx context.Context, pageSize, pageOffset int32, filter *pb.ContractorFilter, sort []string) ([]db.Contractor, int64, error) {
	params := db.ListContractorsParams{
		Limit:  pageSize,
		Offset: pageOffset,
	}

	if filter != nil {
		if filter.CompanyName != "" {
			params.Column1 = filter.CompanyName
		}
		if filter.Department != "" {
			params.Column2 = filter.Department
		}
		if filter.Status != "" {
			params.Column3 = filter.Status
		}
	}

	return s.repo.List(ctx, params)
}
