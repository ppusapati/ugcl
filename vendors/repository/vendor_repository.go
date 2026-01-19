// repository/vendor_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	db "p9e.in/ugcl/vendors/db/generated"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// IVendorRepository defines vendor data access methods
type IVendorRepository interface {
	Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error)
	GetByID(ctx context.Context, id uuid.UUID) (*db.Vendor, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*db.Vendor, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) (*db.Vendor, error)
	GetByPAN(ctx context.Context, pan string) (*db.Vendor, error)
	GetByGST(ctx context.Context, gst string) (*db.Vendor, error)
	Update(ctx context.Context, params db.UpdateVendorParams) (*db.Vendor, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params db.ListVendorsParams) ([]db.Vendor, int64, error)
	ListByCategory(ctx context.Context, category string) ([]db.Vendor, error)
	ListByContract(ctx context.Context, contractID []string) ([]db.Vendor, error)
	ListByPurchaseOrder(ctx context.Context, poID []string) ([]db.Vendor, error)
	ListBlacklisted(ctx context.Context) ([]db.Vendor, error)
}

type VendorRepository struct {
	queries *db.Queries
}

// NewVendorRepository creates a new vendor repository with fx
func NewVendorRepository(q *db.Queries) IVendorRepository {
	return &VendorRepository{
		queries: q,
	}
}

// Create creates a new vendor
func (r *VendorRepository) Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error) {
	vendor, err := r.queries.CreateVendor(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("vendor with pan already exists")
			}
		}
		return nil, fmt.Errorf("failed to create vendor: %w", err)
	}
	return &vendor, nil
}

// GetByID retrieves a vendor by ID
func (r *VendorRepository) GetByID(ctx context.Context, id uuid.UUID) (*db.Vendor, error) {
	vendor, err := r.queries.GetVendorByID(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}
	return &vendor, nil
}

// GetByUUID retrieves a vendor by UUID
func (r *VendorRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*db.Vendor, error) {
	vendor, err := r.queries.GetVendorByUUID(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("failed to get vendor by UUID: %w", err)
	}
	return &vendor, nil
}

// GetByPersonID retrieves a vendor by person ID
func (r *VendorRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) (*db.Vendor, error) {
	vendor, err := r.queries.GetVendorByPersonID(ctx, personID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found for person")
		}
		return nil, fmt.Errorf("failed to get vendor by person ID: %w", err)
	}
	return &vendor, nil
}

// GetByPAN retrieves a vendor by PAN
func (r *VendorRepository) GetByPAN(ctx context.Context, pan string) (*db.Vendor, error) {
	vendor, err := r.queries.GetVendorByPAN(ctx, &pan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("failed to get vendor by PAN: %w", err)
	}
	return &vendor, nil
}

// GetByGST retrieves a vendor by GST
func (r *VendorRepository) GetByGST(ctx context.Context, gst string) (*db.Vendor, error) {
	vendor, err := r.queries.GetVendorByGST(ctx, &gst)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("failed to get vendor by GST: %w", err)
	}
	return &vendor, nil
}

// Update updates a vendor
func (r *VendorRepository) Update(ctx context.Context, params db.UpdateVendorParams) (*db.Vendor, error) {
	vendor, err := r.queries.UpdateVendor(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("failed to update vendor: %w", err)
	}
	return &vendor, nil
}

// Delete soft deletes a vendor
func (r *VendorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteVendor(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("vendor not found")
		}
		return fmt.Errorf("failed to delete vendor: %w", err)
	}
	return nil
}

// List lists vendors with filters
func (r *VendorRepository) List(ctx context.Context, params db.ListVendorsParams) ([]db.Vendor, int64, error) {
	// Get count
	total, err := r.queries.CountVendors(ctx, db.CountVendorsParams{
		Column1: params.Column1, // company_name
		Column2: params.Column2, // vendor_category
		Column3: params.Column3, // status
		Column4: params.Column4, // min_rating
		Column5: params.Column5, // max_rating
		Column6: params.Column6, // is_blacklisted
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count vendors: %w", err)
	}

	// Get list
	vendors, err := r.queries.ListVendors(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list vendors: %w", err)
	}

	return vendors, total, nil
}

// ListByCategory lists vendors by category
func (r *VendorRepository) ListByCategory(ctx context.Context, category string) ([]db.Vendor, error) {
	vendors, err := r.queries.ListVendorsByCategory(ctx, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors by category: %w", err)
	}
	return vendors, nil
}

// ListByContract lists vendors by contract
func (r *VendorRepository) ListByContract(ctx context.Context, contractID []string) ([]db.Vendor, error) {
	vendors, err := r.queries.ListVendorsByContract(ctx, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors by contract: %w", err)
	}
	return vendors, nil
}

// ListByPurchaseOrder lists vendors by purchase order
func (r *VendorRepository) ListByPurchaseOrder(ctx context.Context, poID []string) ([]db.Vendor, error) {
	vendors, err := r.queries.ListVendorsByPurchaseOrder(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors by purchase order: %w", err)
	}
	return vendors, nil
}

// ListBlacklisted lists all blacklisted vendors
func (r *VendorRepository) ListBlacklisted(ctx context.Context) ([]db.Vendor, error) {
	vendors, err := r.queries.ListBlacklistedVendors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list blacklisted vendors: %w", err)
	}
	return vendors, nil
}
