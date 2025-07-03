package database

import (
	"fmt"
	"log"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
)

// RunMigrations runs all database migrations
func RunMigrations(db *Database) error {
	log.Println("Running database migrations...")

	// Auto migrate all models
	err := db.DB.AutoMigrate(
		&models.User{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},
		&models.Category{},
		&models.Ticket{},
		&models.Comment{},
		&models.Attachment{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *Database) error {
	log.Println("Seeding database with initial data...")

	// Check if admin user already exists
	var count int64
	db.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Create default admin user
	adminUser := &models.User{
		Email:        "admin@helpchat.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		FirstName:    "Admin",
		LastName:     "User",
		Role:         models.RoleAdministrator,
		IsVerified:   true,
		IsActive:     true,
	}

	if err := db.DB.Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default categories
	categories := []models.Category{
		{
			Name:        "Technical Support",
			Description: "Technical issues and troubleshooting",
			IsActive:    true,
		},
		{
			Name:        "Account Management",
			Description: "Account-related issues and requests",
			IsActive:    true,
		},
		{
			Name:        "Billing & Payments",
			Description: "Billing and payment-related issues",
			IsActive:    true,
		},
		{
			Name:        "Feature Requests",
			Description: "Requests for new features or improvements",
			IsActive:    true,
		},
		{
			Name:        "Bug Reports",
			Description: "Bug reports and software issues",
			IsActive:    true,
		},
	}

	for _, category := range categories {
		if err := db.DB.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", category.Name, err)
		}
	}

	log.Println("Database seeded successfully")
	return nil
}

// CreateIndexes creates database indexes for better performance
func CreateIndexes(db *Database) error {
	log.Println("Creating database indexes...")

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token ON email_verification_tokens(token)",
		"CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id)",
		// Ticket indexes
		"CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_category_id ON tickets(category_id)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_assigned_agent_id ON tickets(assigned_agent_id)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_created_by_id ON tickets(created_by_id)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_creation_time ON tickets(creation_time)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_escalated_at ON tickets(escalated_at)",
		// Category indexes
		"CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active)",
		// Comment indexes
		"CREATE INDEX IF NOT EXISTS idx_comments_ticket_id ON comments(ticket_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_comments_is_internal ON comments(is_internal)",
		// Attachment indexes
		"CREATE INDEX IF NOT EXISTS idx_attachments_ticket_id ON attachments(ticket_id)",
		"CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by_id ON attachments(uploaded_by_id)",
		"CREATE INDEX IF NOT EXISTS idx_attachments_created_at ON attachments(created_at)",
	}

	for _, index := range indexes {
		if err := db.DB.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}
