package repository

import (
	"p9e.in/ugcl/masters/pipeline/models"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type NodeRepository interface {
	GetByZoneID(zoneID string) ([]models.Node, error)
	GetByLabel(label string) ([]models.Node, error)
	GetByID(id string) (*models.Node, error)
	Create(node *models.Node) error
	Update(node *models.Node) error
	Delete(id string) error
}

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var NodeRepositoryModule = fx.Provide(NewNodeRepository)

func (r *nodeRepository) GetByZoneID(zoneID string) ([]models.Node, error) {
	var nodes []models.Node
	err := r.db.Where("zone_id = ?", zoneID).Find(&nodes).Error
	return nodes, err
}

func (r *nodeRepository) GetByLabel(label string) ([]models.Node, error) {
	var nodes []models.Node

	err := r.db.Table("nodes").
		Joins("JOIN pipeline_segments ps1 ON nodes.id = ps1.start_node_id OR nodes.id = ps1.stop_node_id").
		Where("ps1.label = ?", label).
		Distinct().
		Find(&nodes).Error

	return nodes, err
}

func (r *nodeRepository) GetByID(id string) (*models.Node, error) {
	var node models.Node
	err := r.db.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}

func (r *nodeRepository) Update(node *models.Node) error {
	return r.db.Save(node).Error
}

func (r *nodeRepository) Delete(id string) error {
	return r.db.Delete(&models.Node{}, id).Error
}
