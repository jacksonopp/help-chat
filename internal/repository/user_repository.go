package repository

import (
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(limit, offset int) ([]*models.User, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *database.Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.Database) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	return r.db.DB.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(user *models.User) error {
	return r.db.DB.Save(user).Error
}

// Delete deletes a user by ID
func (r *userRepository) Delete(id string) error {
	return r.db.DB.Where("id = ?", id).Delete(&models.User{}).Error
}

// List retrieves a list of users with pagination
func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.DB.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}
