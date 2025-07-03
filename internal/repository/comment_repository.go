package repository

import (
	"context"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/google/uuid"
)

// commentRepository implements CommentRepository
type commentRepository struct {
	db *database.Database
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *database.Database) CommentRepository {
	return &commentRepository{db: db}
}

// Create creates a new comment
func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	return r.db.DB.WithContext(ctx).Create(comment).Error
}

// GetByID retrieves a comment by ID
func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.DB.WithContext(ctx).
		Preload("Ticket").
		Preload("User").
		Where("id = ?", id).
		First(&comment).Error

	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// Update updates an existing comment
func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	return r.db.DB.WithContext(ctx).Save(comment).Error
}

// Delete deletes a comment by ID
func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&models.Comment{}, id).Error
}

// GetByTicket retrieves comments for a specific ticket
func (r *commentRepository) GetByTicket(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]models.Comment, error) {
	var comments []models.Comment
	query := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC")

	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}

	err := query.Find(&comments).Error
	return comments, err
}

// GetByUser retrieves comments created by a specific user
func (r *commentRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.DB.WithContext(ctx).
		Preload("Ticket").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&comments).Error

	return comments, err
}
