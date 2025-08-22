// repository/postgres/contractor.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	db "p9e.in/ugcl/vendors/db/generated"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ContractorRepository defines contractor data access methods
type IContractorRepository interface {
	Create(ctx context.Context, params db.CreateContractorParams) (*db.Contractor, error)
	GetByID(ctx context.Context, id uuid.UUID) (*db.Contractor, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) (*db.Contractor, error)
	GetByPAN(ctx context.Context, pan string) (*db.Contractor, error)
	GetByGST(ctx context.Context, gst string) (*db.Contractor, error)
	Update(ctx context.Context, params db.UpdateContractorParams) (*db.Contractor, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params db.ListContractorsParams) ([]db.Contractor, int64, error)
	ListByProject(ctx context.Context, projectID []string) ([]db.Contractor, error)
	ListBySite(ctx context.Context, siteID []string) ([]db.Contractor, error)
}

type ContractorRepository struct {
	queries *db.Queries
}

// NewContractorRepository creates a new contractor repository with fx
func NewContractorRepository(q *db.Queries) IContractorRepository {
	return &ContractorRepository{
		queries: q,
	}
}

// Create creates a new contractor
func (r *ContractorRepository) Create(ctx context.Context, params db.CreateContractorParams) (*db.Contractor, error) {
	contractor, err := r.queries.CreateContractor(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("contractor with pan already exists")
			}
		}
		return nil, fmt.Errorf("failed to create contractor: %w", err)
	}
	return &contractor, nil
}

// GetByID retrieves a contractor by ID
func (r *ContractorRepository) GetByID(ctx context.Context, id uuid.UUID) (*db.Contractor, error) {
	contractor, err := r.queries.GetContractorByID(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contractor not found")
		}
		return nil, fmt.Errorf("failed to get contractor: %w", err)
	}
	return &contractor, nil
}

// GetByPersonID retrieves a contractor by person ID
func (r *ContractorRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) (*db.Contractor, error) {
	contractor, err := r.queries.GetContractorByPersonID(ctx, personID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contractor not found for person")
		}
		return nil, fmt.Errorf("failed to get contractor by person ID: %w", err)
	}
	return &contractor, nil
}

// GetByPAN retrieves a contractor by PAN
func (r *ContractorRepository) GetByPAN(ctx context.Context, pan string) (*db.Contractor, error) {
	contractor, err := r.queries.GetContractorByPAN(ctx, &pan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contractor not found")
		}
		return nil, fmt.Errorf("failed to get contractor by PAN: %w", err)
	}
	return &contractor, nil
}

// GetByGST retrieves a contractor by GST
func (r *ContractorRepository) GetByGST(ctx context.Context, gst string) (*db.Contractor, error) {
	contractor, err := r.queries.GetContractorByGST(ctx, &gst)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contractor not found")
		}
		return nil, fmt.Errorf("failed to get contractor by GST: %w", err)
	}
	return &contractor, nil
}

// Update updates a contractor
func (r *ContractorRepository) Update(ctx context.Context, params db.UpdateContractorParams) (*db.Contractor, error) {
	contractor, err := r.queries.UpdateContractor(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contractor not found")
		}
		return nil, fmt.Errorf("failed to update contractor: %w", err)
	}
	return &contractor, nil
}

// Delete soft deletes a contractor
func (r *ContractorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteContractor(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("contractor not found")
		}
		return fmt.Errorf("failed to delete contractor: %w", err)
	}
	return nil
}

// List lists contractors with filters
func (r *ContractorRepository) List(ctx context.Context, params db.ListContractorsParams) ([]db.Contractor, int64, error) {
	// Get count
	countParams := db.ListContractorsParams{
		Column1: params.Column1, // company_name
		Column2: params.Column2, // category
		Column3: params.Column3, // status
		Column4: params.Column4, // associated_project
		Column5: params.Column5, // working_site
	}

	total, err := r.queries.CountContractors(ctx, countParams.Column1, countParams.Column2, countParams.Column3, countParams.Column4, countParams.Column5)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count contractors: %w", err)
	}

	// Get list
	contractors, err := r.queries.ListContractors(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list contractors: %w", err)
	}

	return contractors, total, nil
}

// ListByProject lists contractors by project
func (r *ContractorRepository) ListByProject(ctx context.Context, projectID []string) ([]db.Contractor, error) {
	contractors, err := r.queries.ListContractorsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contractors by project: %w", err)
	}
	return contractors, nil
}

// ListBySite lists contractors by site
func (r *ContractorRepository) ListBySite(ctx context.Context, siteID []string) ([]db.Contractor, error) {
	contractors, err := r.queries.ListContractorsBySite(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contractors by site: %w", err)
	}
	return contractors, nil
}
