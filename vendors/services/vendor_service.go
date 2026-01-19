package services

import (
	"context"
	"fmt"

	userpb "p9e.in/ugcl/identity/user/api/v2/user"
	usermapper "p9e.in/ugcl/identity/user/mappers"
	userservices "p9e.in/ugcl/identity/user/services"
	pb "p9e.in/ugcl/vendors/api/v2/vendor"
	db "p9e.in/ugcl/vendors/db/generated"
	"p9e.in/ugcl/vendors/repository"

	"github.com/google/uuid"
)

// IVendorService defines the business logic for vendor operations
type IVendorService interface {
	Create(ctx context.Context, vendor *db.Vendor, user *userpb.User) (*db.Vendor, error)
	Update(ctx context.Context, vendor *db.Vendor) (*db.Vendor, error)
	GetByID(ctx context.Context, id string) (*db.Vendor, error)
	GetByPersonID(ctx context.Context, personID string) (*db.Vendor, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize, pageOffset int32, filter *pb.VendorFilter, sort []string) ([]db.Vendor, int64, error)
	ListByCategory(ctx context.Context, category string) ([]db.Vendor, error)
}

type VendorService struct {
	repo    repository.IVendorRepository
	userSvc userservices.IUserService
}

// NewVendorService creates a new vendor service
func NewVendorService(repo repository.IVendorRepository, userSvc userservices.IUserService) IVendorService {
	return &VendorService{
		repo:    repo,
		userSvc: userSvc,
	}
}

// Create implements IVendorService.
func (s *VendorService) Create(ctx context.Context, vendor *db.Vendor, user *userpb.User) (*db.Vendor, error) {
	// Step 1: Create user first
	userResp, err := s.userSvc.RegisterUser(ctx, usermapper.UserProtoToModel(user))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	var userID string
	// Extract user ID from response
	if userResp.Uuid != uuid.Nil {
		userID = userResp.Uuid.String()
	} else {
		return nil, fmt.Errorf("user created but has no valid UUID")
	}

	vendor.PersonID = userID

	return s.repo.Create(ctx, db.CreateVendorParams{
		CompanyName:      vendor.CompanyName,
		CompanyType:      vendor.CompanyType,
		Gst:              vendor.Gst,
		Pan:              vendor.Pan,
		VendorCategory:   vendor.VendorCategory,
		PersonID:         vendor.PersonID,
		PaymentTerms:     vendor.PaymentTerms,
		CreditPeriodDays: vendor.CreditPeriodDays,
		BankName:         vendor.BankName,
		AccountNumber:    vendor.AccountNumber,
		Ifsc:             vendor.Ifsc,
		Rating:           vendor.Rating,
		IsBlacklisted:    vendor.IsBlacklisted,
		Contracts:        vendor.Contracts,
		PurchaseOrders:   vendor.PurchaseOrders,
		Status:           vendor.Status,
		Metadata:         vendor.Metadata,
	})
}

// Update implements IVendorService.
func (s *VendorService) Update(ctx context.Context, vendor *db.Vendor) (*db.Vendor, error) {
	params := db.UpdateVendorParams{
		ID:               vendor.ID,
		CompanyName:      &vendor.CompanyName,
		CompanyType:      vendor.CompanyType,
		Gst:              vendor.Gst,
		Pan:              vendor.Pan,
		VendorCategory:   vendor.VendorCategory,
		PersonID:         &vendor.PersonID,
		PaymentTerms:     vendor.PaymentTerms,
		CreditPeriodDays: vendor.CreditPeriodDays,
		BankName:         vendor.BankName,
		AccountNumber:    vendor.AccountNumber,
		Ifsc:             vendor.Ifsc,
		Rating:           vendor.Rating,
		IsBlacklisted:    vendor.IsBlacklisted,
		Contracts:        vendor.Contracts,
		PurchaseOrders:   vendor.PurchaseOrders,
		Status:           vendor.Status,
		Metadata:         vendor.Metadata,
	}

	return s.repo.Update(ctx, params)
}

// GetByID implements IVendorService.
func (s *VendorService) GetByID(ctx context.Context, id string) (*db.Vendor, error) {
	vendorID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.GetByID(ctx, vendorID)
}

// GetByPersonID implements IVendorService.
func (s *VendorService) GetByPersonID(ctx context.Context, personID string) (*db.Vendor, error) {
	pid, err := uuid.Parse(personID)
	if err != nil {
		return nil, fmt.Errorf("invalid person ID: %w", err)
	}
	return s.repo.GetByPersonID(ctx, pid)
}

// Delete implements IVendorService.
func (s *VendorService) Delete(ctx context.Context, id string) error {
	vendorID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.Delete(ctx, vendorID)
}

// List implements IVendorService.
func (s *VendorService) List(ctx context.Context, pageSize, pageOffset int32, filter *pb.VendorFilter, sort []string) ([]db.Vendor, int64, error) {
	params := db.ListVendorsParams{
		Limit:  pageSize,
		Offset: pageOffset,
	}

	if filter != nil {
		if filter.CompanyName != "" {
			params.Column1 = &filter.CompanyName
		}
		if filter.VendorCategory != "" {
			params.Column2 = &filter.VendorCategory
		}
		if filter.Status != "" {
			params.Column3 = &filter.Status
		}
		if filter.MinRating > 0 {
			params.Column4 = &filter.MinRating
		}
		if filter.MaxRating > 0 {
			params.Column5 = &filter.MaxRating
		}
		params.Column6 = &filter.IsBlacklisted
	}

	return s.repo.List(ctx, params)
}

// ListByCategory implements IVendorService.
func (s *VendorService) ListByCategory(ctx context.Context, category string) ([]db.Vendor, error) {
	return s.repo.ListByCategory(ctx, category)
}
