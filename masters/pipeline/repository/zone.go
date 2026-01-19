package repository

import (
	"p9e.in/ugcl/masters/pipeline/models"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ZoneRepository interface {
	GetBySiteID(siteID string) ([]ZoneWithCounts, error)
	GetByID(id string) (*models.Zone, error)
	Create(zone *models.Zone) error
	Update(zone *models.Zone) error
	Delete(id string) error
}

type ZoneWithCounts struct {
	models.Zone
	LabelCount   int64 `json:"label_count"`
	SegmentCount int64 `json:"segment_count"`
}

type zoneRepository struct {
	db *gorm.DB
}

func NewZoneRepository(db *gorm.DB) ZoneRepository {
	return &zoneRepository{db: db}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var ZoneRepositoryModule = fx.Provide(NewZoneRepository)

func (r *zoneRepository) GetBySiteID(siteID string) ([]ZoneWithCounts, error) {
	var results []ZoneWithCounts
	err := r.db.Table("zones").
		Select(`zones.*, 
				COUNT(DISTINCT pipeline_segments.label) as label_count,
				COUNT(pipeline_segments.segment_id) as segment_count`).
		Joins("LEFT JOIN pipeline_segments ON zones.id = pipeline_segments.zone_id").
		Where("zones.site_id = ?", siteID).
		Group("zones.id").
		Find(&results).Error
	return results, err
}

func (r *zoneRepository) GetByID(id string) (*models.Zone, error) {
	var zone models.Zone
	err := r.db.First(&zone, id).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *zoneRepository) Create(zone *models.Zone) error {
	return r.db.Create(zone).Error
}

func (r *zoneRepository) Update(zone *models.Zone) error {
	return r.db.Save(zone).Error
}

func (r *zoneRepository) Delete(id string) error {
	return r.db.Delete(&models.Zone{}, id).Error
}
