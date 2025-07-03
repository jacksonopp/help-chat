package services

import (
	"context"
	"fmt"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/repository"
	"github.com/google/uuid"
)

// TicketService handles ticket-related business logic
type TicketService struct {
	ticketRepo     repository.TicketRepository
	categoryRepo   repository.CategoryRepository
	commentRepo    repository.CommentRepository
	attachmentRepo repository.AttachmentRepository
	userRepo       repository.UserRepository
}

// NewTicketService creates a new ticket service
func NewTicketService(
	ticketRepo repository.TicketRepository,
	categoryRepo repository.CategoryRepository,
	commentRepo repository.CommentRepository,
	attachmentRepo repository.AttachmentRepository,
	userRepo repository.UserRepository,
) *TicketService {
	return &TicketService{
		ticketRepo:     ticketRepo,
		categoryRepo:   categoryRepo,
		commentRepo:    commentRepo,
		attachmentRepo: attachmentRepo,
		userRepo:       userRepo,
	}
}

// CreateTicket creates a new ticket
func (s *TicketService) CreateTicket(ctx context.Context, req *models.CreateTicketRequest, createdByID uuid.UUID) (*models.Ticket, error) {
	// Validate category if provided
	if req.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get category: %w", err)
		}
		if category == nil {
			return nil, fmt.Errorf("category not found")
		}
		if !category.IsActive {
			return nil, fmt.Errorf("category is not active")
		}
	}

	// Create ticket
	ticket := &models.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		CategoryID:  req.CategoryID,
		CreatedByID: createdByID,
		Status:      models.StatusOpen,
		DueDate:     req.DueDate,
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Get the created ticket with relationships
	return s.ticketRepo.GetByID(ctx, ticket.ID)
}

// GetTicket retrieves a ticket by ID
func (s *TicketService) GetTicket(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, ticketID)
}

// UpdateTicket updates an existing ticket
func (s *TicketService) UpdateTicket(ctx context.Context, ticketID uuid.UUID, req *models.UpdateTicketRequest, updatedByID uuid.UUID) (*models.Ticket, error) {
	// Get existing ticket
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// Validate category if provided
	if req.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get category: %w", err)
		}
		if category == nil {
			return nil, fmt.Errorf("category not found")
		}
		if !category.IsActive {
			return nil, fmt.Errorf("category is not active")
		}
		ticket.CategoryID = req.CategoryID
	}

	// Update fields
	if req.Title != nil {
		ticket.Title = *req.Title
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.DueDate != nil {
		ticket.DueDate = req.DueDate
	}

	// Update ticket
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	// Get the updated ticket with relationships
	return s.ticketRepo.GetByID(ctx, ticket.ID)
}

// DeleteTicket deletes a ticket
func (s *TicketService) DeleteTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) error {
	// Check if ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	// Get user to check authorization
	user, err := s.userRepo.GetByID(userID.String())
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Only admins can delete tickets
	if !user.IsAdmin() {
		return fmt.Errorf("insufficient permissions: only administrators can delete tickets")
	}

	// Only allow deletion of open tickets
	if ticket.Status != models.StatusOpen {
		return fmt.Errorf("can only delete open tickets")
	}

	return s.ticketRepo.Delete(ctx, ticketID)
}

// ListTickets retrieves tickets with filtering and pagination
func (s *TicketService) ListTickets(ctx context.Context, query *models.TicketQuery) (*models.TicketListResponse, error) {
	// Set default pagination if not provided
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	return s.ticketRepo.List(ctx, query)
}

// GetTicketStats retrieves ticket statistics
func (s *TicketService) GetTicketStats(ctx context.Context) (*models.TicketStats, error) {
	return s.ticketRepo.GetStats(ctx)
}

// AssignTicket assigns a ticket to an agent
func (s *TicketService) AssignTicket(ctx context.Context, ticketID, agentID uuid.UUID, assignedByID uuid.UUID) error {
	// Check if ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	// Check if agent exists and is a support agent
	agent, err := s.userRepo.GetByID(agentID.String())
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return fmt.Errorf("agent not found")
	}
	if !agent.IsAgent() {
		return fmt.Errorf("user is not a support agent")
	}

	// Assign ticket
	if err := s.ticketRepo.AssignToAgent(ctx, ticketID, agentID); err != nil {
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	return nil
}

// UpdateTicketStatus updates the status of a ticket
func (s *TicketService) UpdateTicketStatus(ctx context.Context, ticketID uuid.UUID, req *models.UpdateTicketStatusRequest, updatedByID uuid.UUID) error {
	// Check if ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	// Validate status transition
	if !s.isValidStatusTransition(ticket.Status, req.Status) {
		return fmt.Errorf("invalid status transition from %s to %s", ticket.Status, req.Status)
	}

	// Update status
	if err := s.ticketRepo.UpdateStatus(ctx, ticketID, req.Status); err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	return nil
}

// EscalateTicket escalates a ticket to another user
func (s *TicketService) EscalateTicket(ctx context.Context, ticketID uuid.UUID, req *models.EscalateTicketRequest, escalatedByID uuid.UUID) error {
	// Check if ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	// Check if ticket is already escalated
	if ticket.IsEscalated() {
		return fmt.Errorf("ticket is already escalated")
	}

	// Check if target user exists and is a manager or admin
	targetUser, err := s.userRepo.GetByID(req.EscalatedTo.String())
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}
	if targetUser == nil {
		return fmt.Errorf("target user not found")
	}
	if !targetUser.IsAdmin() {
		return fmt.Errorf("target user is not a manager or administrator")
	}

	// Escalate ticket
	if err := s.ticketRepo.Escalate(ctx, ticketID, req.EscalatedTo); err != nil {
		return fmt.Errorf("failed to escalate ticket: %w", err)
	}

	return nil
}

// GetTicketsByUser retrieves tickets created by a specific user
func (s *TicketService) GetTicketsByUser(ctx context.Context, userID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	return s.ticketRepo.GetByUser(ctx, userID, query)
}

// GetTicketsByAgent retrieves tickets assigned to a specific agent
func (s *TicketService) GetTicketsByAgent(ctx context.Context, agentID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	return s.ticketRepo.GetByAgent(ctx, agentID, query)
}

// isValidStatusTransition checks if a status transition is valid
func (s *TicketService) isValidStatusTransition(from, to models.TicketStatus) bool {
	validTransitions := map[models.TicketStatus][]models.TicketStatus{
		models.StatusOpen: {
			models.StatusInProgress,
			models.StatusResolved,
			models.StatusClosed,
		},
		models.StatusInProgress: {
			models.StatusOpen,
			models.StatusResolved,
			models.StatusClosed,
		},
		models.StatusResolved: {
			models.StatusInProgress,
			models.StatusClosed,
		},
		models.StatusClosed: {
			models.StatusOpen,
			models.StatusInProgress,
		},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}

	return false
}
