package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ticketRepository implements TicketRepository
type ticketRepository struct {
	db             *database.Database
	timeSeriesRepo *TimeSeriesRepositoryImpl[*models.Ticket]
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *database.Database) TicketRepository {
	return &ticketRepository{
		db:             db,
		timeSeriesRepo: NewTimeSeriesRepository[*models.Ticket](db),
	}
}

// Create creates a new ticket
func (r *ticketRepository) Create(ctx context.Context, ticket *models.Ticket) error {
	return r.timeSeriesRepo.Create(ctx, ticket)
}

// GetByID retrieves the current version of a ticket by ID
func (r *ticketRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Ticket, error) {
	ticketVal, err := r.timeSeriesRepo.GetCurrentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ticket := ticketVal

	// Load relationships
	err = r.db.DB.WithContext(ctx).
		Preload("Category").
		Preload("AssignedAgent").
		Preload("CreatedBy").
		Preload("EscalatedToUser").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("creation_time ASC")
		}).
		Preload("Comments.User").
		Preload("Attachments").
		First(ticket).Error

	if err != nil {
		return nil, err
	}
	return ticket, nil
}

// Update updates an existing ticket (creates a new version and expires the old one)
func (r *ticketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	_, err := r.timeSeriesRepo.Update(ctx, ticket.ID, func(clone *models.Ticket) error {
		// Copy updatable fields from the input ticket to the clone
		clone.Title = ticket.Title
		clone.Description = ticket.Description
		clone.Status = ticket.Status
		clone.Priority = ticket.Priority
		clone.CategoryID = ticket.CategoryID
		clone.AssignedAgentID = ticket.AssignedAgentID
		clone.EscalatedAt = ticket.EscalatedAt
		clone.EscalatedTo = ticket.EscalatedTo
		clone.ResolvedAt = ticket.ResolvedAt
		clone.DueDate = ticket.DueDate
		return nil
	})
	return err
}

// Delete archives the current version of a ticket (marks it as expired)
func (r *ticketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.timeSeriesRepo.Archive(ctx, id)
}

// List retrieves tickets with filtering, sorting, and pagination
func (r *ticketRepository) List(ctx context.Context, query *models.TicketQuery) (*models.TicketListResponse, error) {
	db := r.db.DB.WithContext(ctx).
		Preload("Category").
		Preload("AssignedAgent").
		Preload("CreatedBy")

	// Apply filters
	db = r.applyFilters(db, query.Filter)

	// Get total count
	var total int64
	if err := db.Model(&models.Ticket{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply sorting
	if query.Sort != nil {
		orderClause := fmt.Sprintf("%s %s", query.Sort.Field, strings.ToUpper(query.Sort.Direction))
		db = db.Order(orderClause)
	} else {
		db = db.Order("creation_time DESC")
	}

	// Apply pagination
	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	// Execute query
	var tickets []models.Ticket
	if err := db.Find(&tickets).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &models.TicketListResponse{
		Tickets:    tickets,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStats retrieves ticket statistics
func (r *ticketRepository) GetStats(ctx context.Context) (*models.TicketStats, error) {
	var stats models.TicketStats

	// Get counts for each status
	statuses := []models.TicketStatus{
		models.StatusOpen,
		models.StatusInProgress,
		models.StatusResolved,
		models.StatusClosed,
	}

	for _, status := range statuses {
		var count int64
		if err := r.db.DB.WithContext(ctx).Model(&models.Ticket{}).Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}

		switch status {
		case models.StatusOpen:
			stats.OpenTickets = count
		case models.StatusInProgress:
			stats.InProgressTickets = count
		case models.StatusResolved:
			stats.ResolvedTickets = count
		case models.StatusClosed:
			stats.ClosedTickets = count
		}
	}

	// Get total tickets
	if err := r.db.DB.WithContext(ctx).Model(&models.Ticket{}).Count(&stats.TotalTickets).Error; err != nil {
		return nil, err
	}

	// Get escalated tickets
	if err := r.db.DB.WithContext(ctx).Model(&models.Ticket{}).Where("escalated_at IS NOT NULL").Count(&stats.EscalatedTickets).Error; err != nil {
		return nil, err
	}

	// Get overdue tickets
	if err := r.db.DB.WithContext(ctx).Model(&models.Ticket{}).Where("due_date < ?", time.Now()).Count(&stats.OverdueTickets).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// AssignToAgent assigns a ticket to an agent
func (r *ticketRepository) AssignToAgent(ctx context.Context, ticketID, agentID uuid.UUID) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.Ticket{}).
		Where("id = ?", ticketID).
		Update("assigned_agent_id", agentID).Error
}

// UpdateStatus updates the status of a ticket
func (r *ticketRepository) UpdateStatus(ctx context.Context, ticketID uuid.UUID, status models.TicketStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// Set resolved_at if status is resolved or closed
	if status == models.StatusResolved || status == models.StatusClosed {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return r.db.DB.WithContext(ctx).
		Model(&models.Ticket{}).
		Where("id = ?", ticketID).
		Updates(updates).Error
}

// Escalate escalates a ticket to another user
func (r *ticketRepository) Escalate(ctx context.Context, ticketID, escalatedTo uuid.UUID) error {
	now := time.Now()
	return r.db.DB.WithContext(ctx).
		Model(&models.Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{
			"escalated_to": escalatedTo,
			"escalated_at": &now,
		}).Error
}

// GetByUser retrieves tickets created by a specific user
func (r *ticketRepository) GetByUser(ctx context.Context, userID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	if query.Filter == nil {
		query.Filter = &models.TicketFilter{}
	}
	query.Filter.CreatedBy = &userID
	return r.List(ctx, query)
}

// GetByAgent retrieves tickets assigned to a specific agent
func (r *ticketRepository) GetByAgent(ctx context.Context, agentID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	if query.Filter == nil {
		query.Filter = &models.TicketFilter{}
	}
	query.Filter.AssignedTo = &agentID
	return r.List(ctx, query)
}

// applyFilters applies filters to the database query
func (r *ticketRepository) applyFilters(db *gorm.DB, filter *models.TicketFilter) *gorm.DB {
	if filter == nil {
		return db
	}

	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}

	if filter.Priority != nil {
		db = db.Where("priority = ?", *filter.Priority)
	}

	if filter.CategoryID != nil {
		db = db.Where("category_id = ?", *filter.CategoryID)
	}

	if filter.AssignedTo != nil {
		db = db.Where("assigned_agent_id = ?", *filter.AssignedTo)
	}

	if filter.CreatedBy != nil {
		db = db.Where("created_by_id = ?", *filter.CreatedBy)
	}

	if filter.IsEscalated != nil {
		if *filter.IsEscalated {
			db = db.Where("escalated_at IS NOT NULL")
		} else {
			db = db.Where("escalated_at IS NULL")
		}
	}

	if filter.IsOverdue != nil {
		if *filter.IsOverdue {
			db = db.Where("due_date < ?", time.Now())
		} else {
			db = db.Where("(due_date IS NULL OR due_date >= ?)", time.Now())
		}
	}

	if filter.DateFrom != nil {
		db = db.Where("creation_time >= ?", *filter.DateFrom)
	}

	if filter.DateTo != nil {
		db = db.Where("creation_time <= ?", *filter.DateTo)
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	return db
}
