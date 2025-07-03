package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateTicketRequest represents a request to create a new ticket
type CreateTicketRequest struct {
	Title       string         `json:"title" validate:"required,min=1,max=255"`
	Description string         `json:"description" validate:"required,min=1"`
	Priority    TicketPriority `json:"priority" validate:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	CategoryID  *uuid.UUID     `json:"category_id"`
	DueDate     *time.Time     `json:"due_date"`
}

// UpdateTicketRequest represents a request to update a ticket
type UpdateTicketRequest struct {
	Title       *string         `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,min=1"`
	Priority    *TicketPriority `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	CategoryID  *uuid.UUID      `json:"category_id"`
	DueDate     *time.Time      `json:"due_date"`
}

// UpdateTicketStatusRequest represents a request to update ticket status
type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" validate:"required,oneof=OPEN IN_PROGRESS RESOLVED CLOSED"`
}

// AssignTicketRequest represents a request to assign a ticket to an agent
type AssignTicketRequest struct {
	AgentID uuid.UUID `json:"agent_id" validate:"required"`
}

// EscalateTicketRequest represents a request to escalate a ticket
type EscalateTicketRequest struct {
	EscalatedTo uuid.UUID `json:"escalated_to" validate:"required"`
	Reason      string    `json:"reason" validate:"required,min=1"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content    string `json:"content" validate:"required,min=1"`
	IsInternal bool   `json:"is_internal"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

// TicketFilter represents filters for ticket queries
type TicketFilter struct {
	Status      *TicketStatus   `json:"status"`
	Priority    *TicketPriority `json:"priority"`
	CategoryID  *uuid.UUID      `json:"category_id"`
	AssignedTo  *uuid.UUID      `json:"assigned_to"`
	CreatedBy   *uuid.UUID      `json:"created_by"`
	IsEscalated *bool           `json:"is_escalated"`
	IsOverdue   *bool           `json:"is_overdue"`
	DateFrom    *time.Time      `json:"date_from"`
	DateTo      *time.Time      `json:"date_to"`
	Search      string          `json:"search"`
}

// TicketSort represents sorting options for ticket queries
type TicketSort struct {
	Field     string `json:"field" validate:"required,oneof=created_at updated_at priority status title"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"`
}

// TicketQuery represents a complete ticket query with filters, sorting, and pagination
type TicketQuery struct {
	Filter   *TicketFilter `json:"filter"`
	Sort     *TicketSort   `json:"sort"`
	Page     int           `json:"page" validate:"min=1"`
	PageSize int           `json:"page_size" validate:"min=1,max=100"`
}

// TicketListResponse represents a paginated list of tickets
type TicketListResponse struct {
	Tickets    []Ticket `json:"tickets"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// TicketStats represents ticket statistics
type TicketStats struct {
	TotalTickets      int64 `json:"total_tickets"`
	OpenTickets       int64 `json:"open_tickets"`
	InProgressTickets int64 `json:"in_progress_tickets"`
	ResolvedTickets   int64 `json:"resolved_tickets"`
	ClosedTickets     int64 `json:"closed_tickets"`
	EscalatedTickets  int64 `json:"escalated_tickets"`
	OverdueTickets    int64 `json:"overdue_tickets"`
}

// CategoryRequest represents a request to create or update a category
type CategoryRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=100"`
	Description string     `json:"description" validate:"max=500"`
	ParentID    *uuid.UUID `json:"parent_id"`
	IsActive    bool       `json:"is_active"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []Category `json:"categories"`
}

// CommentListResponse represents a list of comments
type CommentListResponse struct {
	Comments []Comment `json:"comments"`
}

// AttachmentListResponse represents a list of attachments
type AttachmentListResponse struct {
	Attachments []Attachment `json:"attachments"`
}

// TicketHistory represents a ticket history entry
type TicketHistory struct {
	ID        uuid.UUID `json:"id"`
	Action    string    `json:"action"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
}

// TicketHistoryResponse represents a list of ticket history entries
type TicketHistoryResponse struct {
	History []TicketHistory `json:"history"`
}
