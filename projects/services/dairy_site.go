package services

import (
	"context"

	db "p9e.in/ugcl/projects/db/generated"
	"p9e.in/ugcl/projects/repository"
)

type IDairySiteService interface {
	GetAllWithUser(ctx context.Context) ([]db.DairySitesWithUser, error)
	Create(ctx context.Context, params db.CreateDairySiteParams) (db.DairySite, error)
}

type DairySiteService struct {
	repo repository.IDairySiteRepository
}

func NewDairySiteService(repo repository.IDairySiteRepository) IDairySiteService {
	return &DairySiteService{repo: repo}
}

func (s *DairySiteService) GetAllWithUser(ctx context.Context) ([]db.DairySitesWithUser, error) {
	return s.repo.GetAllWithUser(ctx)
}

func (s *DairySiteService) Create(ctx context.Context, params db.CreateDairySiteParams) (db.DairySite, error) {
	return s.repo.Create(ctx, params)
}
