package repository

import (
	"context"

	db "p9e.in/ugcl/projects/db/generated"
)

// DairySiteRepository defines the public interface
type IDairySiteRepository interface {
	GetAll(ctx context.Context) ([]db.DairySite, error)
	GetByID(ctx context.Context, id string) (db.DairySite, error)
	GetByEmployeeID(ctx context.Context, employeeID string) ([]db.DairySite, error)
	Create(ctx context.Context, arg db.CreateDairySiteParams) (db.DairySite, error)
	Update(ctx context.Context, arg db.UpdateDairySiteParams) (db.DairySite, error)
	SoftDelete(ctx context.Context, id string) error

	// With user join
	GetAllWithUser(ctx context.Context) ([]db.DairySitesWithUser, error)
	GetWithUserByID(ctx context.Context, id string) (db.DairySitesWithUser, error)
}

type DairySiteRepository struct {
	queries *db.Queries
}

// NewRepository creates a new DairySiteRepository
func NewDairySiteRepository(q *db.Queries) IDairySiteRepository {
	return &DairySiteRepository{
		queries: q,
	}
}

func (r *DairySiteRepository) GetAll(ctx context.Context) ([]db.DairySite, error) {
	return r.queries.GetAllDairySites(ctx)
}

func (r *DairySiteRepository) GetByID(ctx context.Context, id string) (db.DairySite, error) {
	return r.queries.GetDairySiteByID(ctx, db.GetDairySiteByIDParams{ID: id})
}

func (r *DairySiteRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]db.DairySite, error) {
	return r.queries.GetDairySitesByEmployeeID(ctx, db.GetDairySitesByEmployeeIDParams{EmployeeID: employeeID})
}

func (r *DairySiteRepository) Create(ctx context.Context, arg db.CreateDairySiteParams) (db.DairySite, error) {
	return r.queries.CreateDairySite(ctx, arg)
}

func (r *DairySiteRepository) Update(ctx context.Context, arg db.UpdateDairySiteParams) (db.DairySite, error) {
	return r.queries.UpdateDairySite(ctx, arg)
}

func (r *DairySiteRepository) SoftDelete(ctx context.Context, id string) error {
	return r.queries.SoftDeleteDairySite(ctx, db.SoftDeleteDairySiteParams{ID: id})
}

// With User Join
func (r *DairySiteRepository) GetAllWithUser(ctx context.Context) ([]db.DairySitesWithUser, error) {
	return r.queries.GetAllDairySitesWithUser(ctx)
}

func (r *DairySiteRepository) GetWithUserByID(ctx context.Context, id string) (db.DairySitesWithUser, error) {
	return r.queries.GetDairySiteWithUserByID(ctx, db.GetDairySiteWithUserByIDParams{ID: id})
}
