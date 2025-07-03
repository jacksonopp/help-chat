package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleEndUser       UserRole = "END_USER"
	RoleSupportAgent  UserRole = "SUPPORT_AGENT"
	RoleAdministrator UserRole = "ADMINISTRATOR"
	RoleManager       UserRole = "MANAGER"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"` // "-" means this field won't be included in JSON
	FirstName    string     `json:"first_name" gorm:"not null"`
	LastName     string     `json:"last_name" gorm:"not null"`
	Role         UserRole   `json:"role" gorm:"not null;default:'END_USER'"`
	IsVerified   bool       `json:"is_verified" gorm:"default:false"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy    *string    `json:"created_by" gorm:"type:char(36)"`
	UpdatedBy    *string    `json:"updated_by" gorm:"type:char(36)"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin returns true if the user has administrator privileges
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdministrator || u.Role == RoleManager
}

// IsAgent returns true if the user is a support agent or higher
func (u *User) IsAgent() bool {
	return u.Role == RoleSupportAgent || u.Role == RoleAdministrator || u.Role == RoleManager
}
