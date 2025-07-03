package repository

import (
	"context"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// categoryRepository implements CategoryRepository
type categoryRepository struct {
	db *database.Database
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *database.Database) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.DB.WithContext(ctx).Create(category).Error
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.DB.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Preload("Tickets").
		Where("id = ?", id).
		First(&category).Error

	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Update updates an existing category
func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.DB.WithContext(ctx).Save(category).Error
}

// Delete deletes a category by ID
func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if category has children
	var childCount int64
	if err := r.db.DB.WithContext(ctx).Model(&models.Category{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return err
	}

	if childCount > 0 {
		return gorm.ErrInvalidData
	}

	// Check if category has tickets
	var ticketCount int64
	if err := r.db.DB.WithContext(ctx).Model(&models.Ticket{}).Where("category_id = ?", id).Count(&ticketCount).Error; err != nil {
		return err
	}

	if ticketCount > 0 {
		return gorm.ErrInvalidData
	}

	return r.db.DB.WithContext(ctx).Delete(&models.Category{}, id).Error
}

// List retrieves all categories
func (r *categoryRepository) List(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.DB.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

// ListActive retrieves only active categories
func (r *categoryRepository) ListActive(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.DB.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

// GetWithChildren retrieves a category with all its children
func (r *categoryRepository) GetWithChildren(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.DB.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Preload("Children.Children").
		Where("id = ?", id).
		First(&category).Error

	if err != nil {
		return nil, err
	}
	return &category, nil
}
