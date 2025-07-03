package repository

import (
	"context"
	"fmt"
	"time"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/google/uuid"
)

// TimeSeriesRepositoryImpl provides a generic implementation of TimeSeriesRepository
type TimeSeriesRepositoryImpl[T models.TimeSeriesEntity] struct {
	db *database.Database
}

// NewTimeSeriesRepository creates a new time-series repository
func NewTimeSeriesRepository[T models.TimeSeriesEntity](db *database.Database) *TimeSeriesRepositoryImpl[T] {
	return &TimeSeriesRepositoryImpl[T]{db: db}
}

// Create creates a new version of an entity
func (r *TimeSeriesRepositoryImpl[T]) Create(ctx context.Context, entity T) error {
	return r.db.DB.WithContext(ctx).Create(entity).Error
}

// GetCurrentByID retrieves the current version of an entity by its logical ID
// This finds the version where ExpirationTime is null
func (r *TimeSeriesRepositoryImpl[T]) GetCurrentByID(ctx context.Context, id uuid.UUID) (T, error) {
	var entity T
	err := r.db.DB.WithContext(ctx).
		Where("id = ? AND expiration_time IS NULL", id).
		First(&entity).Error

	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// GetByID retrieves a specific version of an entity by its version ID
func (r *TimeSeriesRepositoryImpl[T]) GetByID(ctx context.Context, id uuid.UUID) (T, error) {
	var entity T
	err := r.db.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&entity).Error

	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// GetHistory retrieves all versions of an entity by its logical ID
// This finds all versions with the same business ID, ordered by creation time
func (r *TimeSeriesRepositoryImpl[T]) GetHistory(ctx context.Context, id uuid.UUID) ([]T, error) {
	var entities []T
	err := r.db.DB.WithContext(ctx).
		Where("id = ?", id).
		Order("creation_time ASC").
		Find(&entities).Error

	return entities, err
}

// Update creates a new version by cloning the current version and applying updates
func (r *TimeSeriesRepositoryImpl[T]) Update(ctx context.Context, id uuid.UUID, updates func(T) error) (T, error) {
	// Start a transaction
	tx := r.db.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		var zero T
		return zero, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the current version
	var current T
	err := tx.Where("id = ? AND expiration_time IS NULL", id).First(&current).Error
	if err != nil {
		tx.Rollback()
		var zero T
		return zero, fmt.Errorf("failed to get current version: %w", err)
	}

	// Clone the current version using the Cloneable interface
	cloned := current.Clone().(T)

	// Apply updates to the cloned version
	if err := updates(cloned); err != nil {
		tx.Rollback()
		var zero T
		return zero, fmt.Errorf("failed to apply updates: %w", err)
	}

	// Set expiration time on the current version
	now := time.Now()
	current.SetExpirationTime(&now)

	// Update the current version to expire it
	if err := tx.Save(&current).Error; err != nil {
		tx.Rollback()
		var zero T
		return zero, fmt.Errorf("failed to expire current version: %w", err)
	}

	// Create the new version
	if err := tx.Create(&cloned).Error; err != nil {
		tx.Rollback()
		var zero T
		return zero, fmt.Errorf("failed to create new version: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return cloned, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cloned, nil
}

// Archive marks the current version as expired (archives it)
func (r *TimeSeriesRepositoryImpl[T]) Archive(ctx context.Context, id uuid.UUID) error {
	// Start a transaction
	tx := r.db.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the current version
	var current T
	err := tx.Where("id = ? AND expiration_time IS NULL", id).First(&current).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Set expiration time to now
	now := time.Now()
	current.SetExpirationTime(&now)

	// Update the current version to archive it
	if err := tx.Save(&current).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to archive current version: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HardDelete permanently removes all versions of an entity
func (r *TimeSeriesRepositoryImpl[T]) HardDelete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error
}

// GetCurrentByBusinessID retrieves the current version of an entity by its business ID
// This is useful when you have a separate business ID field
func (r *TimeSeriesRepositoryImpl[T]) GetCurrentByBusinessID(ctx context.Context, businessID uuid.UUID) (T, error) {
	var entity T
	err := r.db.DB.WithContext(ctx).
		Where("business_id = ? AND expiration_time IS NULL", businessID).
		First(&entity).Error

	if err != nil {
		var zero T
		return zero, err
	}

	return entity, nil
}

// GetHistoryByBusinessID retrieves all versions of an entity by its business ID
func (r *TimeSeriesRepositoryImpl[T]) GetHistoryByBusinessID(ctx context.Context, businessID uuid.UUID) ([]T, error) {
	var entities []T
	err := r.db.DB.WithContext(ctx).
		Where("business_id = ?", businessID).
		Order("creation_time ASC").
		Find(&entities).Error

	return entities, err
}
