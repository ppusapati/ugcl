package services

import (
	"p9e.in/ugcl/masters/pipeline/models"
	"p9e.in/ugcl/masters/pipeline/repository"

	"go.uber.org/fx"
)

type PipelineService interface {
	GetPipelineSegments(filter repository.SegmentFilter) ([]models.PipelineSegment, int64, error)
	CreatePipelineSegment(segment *models.PipelineSegment) error
	GetPipelineSegmentByID(id string) (*models.PipelineSegment, error)
	UpdatePipelineSegment(segment *models.PipelineSegment) error
	DeletePipelineSegment(id string) error
	GetPipelineView(siteName string, zoneName string, label string) ([]models.PipelineSegmentView, error)
}

type pipelineService struct {
	pipelineRepo repository.PipelineRepository
}

// GetPipelineView implements PipelineService.
func (s *pipelineService) GetPipelineView(siteName string, zoneName string, label string) ([]models.PipelineSegmentView, error) {
	return s.pipelineRepo.GetPipelineView(siteName, zoneName, label)
}

func NewPipelineService(pipelineRepo repository.PipelineRepository) PipelineService {
	return &pipelineService{
		pipelineRepo: pipelineRepo,
	}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var PipelineServiceModule = fx.Provide(NewPipelineService)

func (s *pipelineService) GetPipelineSegments(filter repository.SegmentFilter) ([]models.PipelineSegment, int64, error) {
	return s.pipelineRepo.GetSegments(filter)
}

func (s *pipelineService) CreatePipelineSegment(segment *models.PipelineSegment) error {
	return s.pipelineRepo.Create(segment)
}

func (s *pipelineService) GetPipelineSegmentByID(id string) (*models.PipelineSegment, error) {
	return s.pipelineRepo.GetByID(id)
}

func (s *pipelineService) UpdatePipelineSegment(segment *models.PipelineSegment) error {
	return s.pipelineRepo.Update(segment)
}

func (s *pipelineService) DeletePipelineSegment(id string) error {
	return s.pipelineRepo.Delete(id)
}
