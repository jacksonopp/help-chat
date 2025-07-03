package repository

import (
	"context"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/google/uuid"
)

// TicketRepository defines the interface for ticket data operations
type TicketRepository interface {
	Create(ctx context.Context, ticket *models.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Ticket, error)
	Update(ctx context.Context, ticket *models.Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, query *models.TicketQuery) (*models.TicketListResponse, error)
	GetStats(ctx context.Context) (*models.TicketStats, error)
	AssignToAgent(ctx context.Context, ticketID, agentID uuid.UUID) error
	UpdateStatus(ctx context.Context, ticketID uuid.UUID, status models.TicketStatus) error
	Escalate(ctx context.Context, ticketID, escalatedTo uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error)
	GetByAgent(ctx context.Context, agentID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error)
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]models.Category, error)
	ListActive(ctx context.Context) ([]models.Category, error)
	GetWithChildren(ctx context.Context, id uuid.UUID) (*models.Category, error)
}

// CommentRepository defines the interface for comment data operations
type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTicket(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]models.Comment, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Comment, error)
}

// AttachmentRepository defines the interface for attachment data operations
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *models.Attachment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Attachment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]models.Attachment, error)
	UpdateVirusScan(ctx context.Context, id uuid.UUID, isScanned, isSafe bool) error
}
