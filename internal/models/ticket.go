package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	StatusOpen       TicketStatus = "OPEN"
	StatusInProgress TicketStatus = "IN_PROGRESS"
	StatusResolved   TicketStatus = "RESOLVED"
	StatusClosed     TicketStatus = "CLOSED"
)

// TicketPriority represents the priority of a ticket
type TicketPriority string

const (
	PriorityLow      TicketPriority = "LOW"
	PriorityMedium   TicketPriority = "MEDIUM"
	PriorityHigh     TicketPriority = "HIGH"
	PriorityCritical TicketPriority = "CRITICAL"
)

// Ticket represents a support ticket in the system with time-series versioning
type Ticket struct {
	// Time-series fields
	ID             uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	CreationTime   time.Time  `json:"creation_time" gorm:"autoCreateTime;not null"`
	ExpirationTime *time.Time `json:"expiration_time" gorm:"index"`

	// Business fields
	Title           string         `json:"title" gorm:"not null;size:255"`
	Description     string         `json:"description" gorm:"not null;type:text"`
	Status          TicketStatus   `json:"status" gorm:"not null;default:'OPEN';size:20"`
	Priority        TicketPriority `json:"priority" gorm:"not null;default:'MEDIUM';size:20"`
	CategoryID      *uuid.UUID     `json:"category_id" gorm:"type:char(36)"`
	AssignedAgentID *uuid.UUID     `json:"assigned_agent_id" gorm:"type:char(36)"`
	CreatedByID     uuid.UUID      `json:"created_by_id" gorm:"type:char(36);not null"`
	EscalatedAt     *time.Time     `json:"escalated_at"`
	EscalatedTo     *uuid.UUID     `json:"escalated_to" gorm:"type:char(36)"`
	ResolvedAt      *time.Time     `json:"resolved_at"`
	DueDate         *time.Time     `json:"due_date"`

	// Relationships
	Category        *Category    `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	AssignedAgent   *User        `json:"assigned_agent,omitempty" gorm:"foreignKey:AssignedAgentID"`
	CreatedBy       *User        `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID"`
	EscalatedToUser *User        `json:"escalated_to_user,omitempty" gorm:"foreignKey:EscalatedTo"`
	Comments        []Comment    `json:"comments,omitempty" gorm:"foreignKey:TicketID"`
	Attachments     []Attachment `json:"attachments,omitempty" gorm:"foreignKey:TicketID"`
}

// Category represents a ticket category
type Category struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	Name        string     `json:"name" gorm:"not null;size:100"`
	Description string     `json:"description" gorm:"size:500"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:char(36)"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Tickets  []Ticket   `json:"tickets,omitempty" gorm:"foreignKey:CategoryID"`
}

// Comment represents a comment on a ticket
type Comment struct {
	ID         uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	TicketID   uuid.UUID `json:"ticket_id" gorm:"type:char(36);not null"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	Content    string    `json:"content" gorm:"not null;type:text"`
	IsInternal bool      `json:"is_internal" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Ticket *Ticket `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Attachment represents a file attachment on a ticket
type Attachment struct {
	ID             uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	TicketID       uuid.UUID `json:"ticket_id" gorm:"type:char(36);not null"`
	Filename       string    `json:"filename" gorm:"not null;size:255"`
	FilePath       string    `json:"file_path" gorm:"not null;size:500"`
	FileSize       int64     `json:"file_size" gorm:"not null"`
	MimeType       string    `json:"mime_type" gorm:"not null;size:100"`
	UploadedByID   uuid.UUID `json:"uploaded_by_id" gorm:"type:char(36);not null"`
	IsVirusScanned bool      `json:"is_virus_scanned" gorm:"default:false"`
	IsSafe         bool      `json:"is_safe" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Ticket     *Ticket `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
	UploadedBy *User   `json:"uploaded_by,omitempty" gorm:"foreignKey:UploadedByID"`
}

// TableName specifies the table name for the Ticket model
func (Ticket) TableName() string {
	return "tickets"
}

// TableName specifies the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// TableName specifies the table name for the Comment model
func (Comment) TableName() string {
	return "comments"
}

// TableName specifies the table name for the Attachment model
func (Attachment) TableName() string {
	return "attachments"
}

// BeforeCreate is a GORM hook that runs before creating a ticket
func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a category
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a comment
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating an attachment
func (a *Attachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsOpen returns true if the ticket is open or in progress
func (t *Ticket) IsOpen() bool {
	return t.Status == StatusOpen || t.Status == StatusInProgress
}

// IsResolved returns true if the ticket is resolved or closed
func (t *Ticket) IsResolved() bool {
	return t.Status == StatusResolved || t.Status == StatusClosed
}

// IsEscalated returns true if the ticket has been escalated
func (t *Ticket) IsEscalated() bool {
	return t.EscalatedAt != nil
}

// IsOverdue returns true if the ticket has a due date that has passed
func (t *Ticket) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// TimeSeriesEntity interface implementation

// GetID returns the unique identifier of the ticket
func (t *Ticket) GetID() uuid.UUID {
	return t.ID
}

// GetCreationTime returns when this version was created
func (t *Ticket) GetCreationTime() time.Time {
	return t.CreationTime
}

// GetExpirationTime returns when this version expires (null means current version)
func (t *Ticket) GetExpirationTime() *time.Time {
	return t.ExpirationTime
}

// SetExpirationTime sets the expiration time for this version
func (t *Ticket) SetExpirationTime(expirationTime *time.Time) {
	t.ExpirationTime = expirationTime
}

// IsCurrentVersion returns true if this is the current version (expiration time is null)
func (t *Ticket) IsCurrentVersion() bool {
	return t.ExpirationTime == nil
}

// Clone creates a new version of this ticket with a new ID and current creation time
func (t *Ticket) Clone() Cloneable {
	// Create a new ticket with the same business fields but new time-series fields
	cloned := &Ticket{
		Title:           t.Title,
		Description:     t.Description,
		Status:          t.Status,
		Priority:        t.Priority,
		CategoryID:      t.CategoryID,
		AssignedAgentID: t.AssignedAgentID,
		CreatedByID:     t.CreatedByID,
		EscalatedAt:     t.EscalatedAt,
		EscalatedTo:     t.EscalatedTo,
		ResolvedAt:      t.ResolvedAt,
		DueDate:         t.DueDate,
		CreationTime:    time.Now(),
		ExpirationTime:  nil, // New version is current
	}
	// Generate new ID for the cloned ticket
	cloned.ID = uuid.New()
	return cloned
}
