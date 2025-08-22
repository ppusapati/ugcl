package repository

import (
	"p9e.in/ugcl/masters/pipeline/models"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type PipelineRepository interface {
	GetSegments(filter SegmentFilter) ([]models.PipelineSegment, int64, error)
	GetLabelsByZoneID(zoneID string) ([]LabelInfo, error)
	GetByID(id string) (*models.PipelineSegment, error)
	Create(segment *models.PipelineSegment) error
	Update(segment *models.PipelineSegment) error
	Delete(id string) error
	GetPipelineView(siteName, zoneName, label string) ([]models.PipelineSegmentView, error)
}

type SegmentFilter struct {
	SiteID string
	ZoneID string
	Label  string
	Limit  int
	Offset int
}

type LabelInfo struct {
	Label        string   `json:"label"`
	SegmentCount int64    `json:"segment_count"`
	TotalLength  float64  `json:"total_length"`
	PipeUsed     []string `json:"pipe_used"`
}

type pipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) PipelineRepository {
	return &pipelineRepository{db: db}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var PipelineRepositoryModule = fx.Provide(NewPipelineRepository)

func (r *pipelineRepository) GetSegments(filter SegmentFilter) ([]models.PipelineSegment, int64, error) {
	var segments []models.PipelineSegment
	var total int64

	query := r.db.Model(&models.PipelineSegment{}).
		Preload("Site").
		Preload("Zone").
		Preload("StartNode").
		Preload("StopNode").
		Preload("Material")

	if filter.SiteID == "" {
		query = query.Where("site_id = ?", filter.SiteID)
	}
	if filter.ZoneID == "" {
		query = query.Where("zone_id = ?", filter.ZoneID)
	}
	if filter.Label != "" {
		query = query.Where("label = ?", filter.Label)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err = query.Find(&segments).Error
	return segments, total, err
}

func (r *pipelineRepository) GetLabelsByZoneID(zoneID string) ([]LabelInfo, error) {
	var results []LabelInfo
	err := r.db.Table("pipeline_segments ps").
		Select(`ps.label,
				COUNT(ps.segment_id) as segment_count,
				SUM(ps.length_m) as total_length,
				GROUP_CONCAT(DISTINCT m.pipe_name) as pipe_used`).
		Joins("JOIN pipes m ON ps.pipe_id = m.pipe_id").
		Where("ps.zone_id = ?", zoneID).
		Group("ps.label").
		Order("ps.label").
		Find(&results).Error

	// Process pipe_used field to convert to slice
	for i := range results {
		if len(results[i].PipeUsed) > 0 {
			// Split comma-separated string into slice
			// This is a simplified version - you might want to use strings.Split
		}
	}

	return results, err
}

func (r *pipelineRepository) GetByID(id string) (*models.PipelineSegment, error) {
	var segment models.PipelineSegment
	err := r.db.Preload("Site").
		Preload("Zone").
		Preload("StartNode").
		Preload("StopNode").
		Preload("Material").
		First(&segment, id).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

func (r *pipelineRepository) Create(segment *models.PipelineSegment) error {
	return r.db.Create(segment).Error
}

func (r *pipelineRepository) Update(segment *models.PipelineSegment) error {
	return r.db.Save(segment).Error
}

func (r *pipelineRepository) Delete(id string) error {
	return r.db.Delete(&models.PipelineSegment{}, id).Error
}

func (r *pipelineRepository) GetPipelineView(siteName, zoneName, label string) ([]models.PipelineSegmentView, error) {
	var results []models.PipelineSegmentView

	query := r.db.Table("vw_pipeline_segments")

	// Logic: if only siteName provided, get zones; if siteName+zoneName, get labels; if all three, get full details
	if siteName != "" && zoneName != "" && label != "" {
		// Get full pipeline details
		query = query.Select("*")
	} else if siteName != "" && zoneName != "" {
		// Get distinct labels for this zone
		query = query.Select("DISTINCT label")
	} else if siteName != "" {
		// Get distinct zones for this site
		query = query.Select("DISTINCT zone_name")
	} else {
		// Get distinct sites
		query = query.Select("DISTINCT site_name")
	}

	if siteName != "" {
		query = query.Where("site_name = ?", siteName)
	}
	if zoneName != "" {
		query = query.Where("zone_name = ?", zoneName)
	}
	if label != "" {
		query = query.Where("label = ?", label)
	}

	err := query.Find(&results).Error
	return results, err
}
