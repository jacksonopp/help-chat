package repository

import (
	"context"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/google/uuid"
)

// attachmentRepository implements AttachmentRepository
type attachmentRepository struct {
	db *database.Database
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *database.Database) AttachmentRepository {
	return &attachmentRepository{db: db}
}

// Create creates a new attachment
func (r *attachmentRepository) Create(ctx context.Context, attachment *models.Attachment) error {
	return r.db.DB.WithContext(ctx).Create(attachment).Error
}

// GetByID retrieves an attachment by ID
func (r *attachmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.DB.WithContext(ctx).
		Preload("Ticket").
		Preload("UploadedBy").
		Where("id = ?", id).
		First(&attachment).Error

	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// Delete deletes an attachment by ID
func (r *attachmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&models.Attachment{}, id).Error
}

// GetByTicket retrieves attachments for a specific ticket
func (r *attachmentRepository) GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.DB.WithContext(ctx).
		Preload("UploadedBy").
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&attachments).Error

	return attachments, err
}

// UpdateVirusScan updates the virus scan status of an attachment
func (r *attachmentRepository) UpdateVirusScan(ctx context.Context, id uuid.UUID, isScanned, isSafe bool) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.Attachment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_virus_scanned": isScanned,
			"is_safe":          isSafe,
		}).Error
}
