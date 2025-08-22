package repository

import (
	"p9e.in/ugcl/masters/pipeline/models"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type SiteRepository interface {
	GetAll() ([]models.Site, error)
	GetByID(id uint) (*models.Site, error)
	Create(site *models.Site) error
	Update(site *models.Site) error
	Delete(id uint) error
	GetSitesWithZoneCounts() ([]SiteWithCount, error)
}

type SiteWithCount struct {
	models.Site
	ZoneCount int64 `json:"zone_count"`
}

type siteRepository struct {
	db *gorm.DB
}

func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{db: db}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var SiteRepositoryModule = fx.Provide(NewSiteRepository)

func (r *siteRepository) GetAll() ([]models.Site, error) {
	var sites []models.Site
	err := r.db.Find(&sites).Error
	return sites, err
}

func (r *siteRepository) GetByID(id uint) (*models.Site, error) {
	var site models.Site
	err := r.db.First(&site, id).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *siteRepository) Create(site *models.Site) error {
	return r.db.Create(site).Error
}

func (r *siteRepository) Update(site *models.Site) error {
	return r.db.Save(site).Error
}

func (r *siteRepository) Delete(id uint) error {
	return r.db.Delete(&models.Site{}, id).Error
}

func (r *siteRepository) GetSitesWithZoneCounts() ([]SiteWithCount, error) {
	var results []SiteWithCount
	err := r.db.Table("sites").
		Select("sites.*, COUNT(zones.id) as zone_count").
		Joins("LEFT JOIN zones ON sites.id = zones.id").
		Group("sites.id").
		Find(&results).Error
	return results, err
}
