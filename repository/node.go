package repository

import (
	"context"
	"fmt"

	"NTMonitor/models"
	"gorm.io/gorm"
)

type NodeRepository struct {
	DB *gorm.DB
}

// NewNodeRepository instantiates a new repository instance.
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{DB: db}
}

func (r *NodeRepository) Create(node *models.Node) error {
	return r.DB.Create(node).Error
}

func (r *NodeRepository) FindAll() ([]models.Node, error) {
	var nodes []models.Node
	err := r.DB.Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *NodeRepository) FindByID(id string) (*models.Node, error) {
	var node models.Node
	err := r.DB.First(&node, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}


func (r *NodeRepository) UpdateNode(ctx context.Context, node *models.Node) error {
	// WithContext ensures database queries respect timeouts and cancellations.
	if err := r.DB.WithContext(ctx).Save(node).Error; err != nil {
		return fmt.Errorf("failed to update node %d: %w", node.ID, err)
	}
	return nil
}