package services

import (
	"p9e.in/ugcl/masters/pipeline/models"
	"p9e.in/ugcl/masters/pipeline/repository"

	"go.uber.org/fx"
)

type HierarchyService interface {
	GetSites() ([]repository.SiteWithCount, error)
	GetZonesBySite(siteID string) ([]repository.ZoneWithCounts, error)
	GetLabelsByZone(zoneID string) ([]repository.LabelInfo, error)
	GetNodesByZone(zoneID string) ([]models.Node, error)
	GetNodesByLabel(label string) ([]models.Node, error)
}

type hierarchyService struct {
	siteRepo     repository.SiteRepository
	zoneRepo     repository.ZoneRepository
	nodeRepo     repository.NodeRepository
	pipelineRepo repository.PipelineRepository
}

func NewHierarchyService(
	siteRepo repository.SiteRepository,
	zoneRepo repository.ZoneRepository,
	nodeRepo repository.NodeRepository,
	pipelineRepo repository.PipelineRepository,
) HierarchyService {
	return &hierarchyService{
		siteRepo:     siteRepo,
		zoneRepo:     zoneRepo,
		nodeRepo:     nodeRepo,
		pipelineRepo: pipelineRepo,
	}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var HierarchyServiceModule = fx.Provide(NewHierarchyService)

func (s *hierarchyService) GetSites() ([]repository.SiteWithCount, error) {
	return s.siteRepo.GetSitesWithZoneCounts()
}

func (s *hierarchyService) GetZonesBySite(siteID string) ([]repository.ZoneWithCounts, error) {
	return s.zoneRepo.GetBySiteID(siteID)
}

func (s *hierarchyService) GetLabelsByZone(zoneID string) ([]repository.LabelInfo, error) {
	return s.pipelineRepo.GetLabelsByZoneID(zoneID)
}

func (s *hierarchyService) GetNodesByZone(zoneID string) ([]models.Node, error) {
	return s.nodeRepo.GetByZoneID(zoneID)
}

func (s *hierarchyService) GetNodesByLabel(label string) ([]models.Node, error) {
	return s.nodeRepo.GetByLabel(label)
}
