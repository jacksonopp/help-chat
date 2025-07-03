package models

import (
	"time"

	"github.com/google/uuid"
)

// Cloneable defines an interface for types that can be deeply cloned
// Clone should return a new instance with a new ID and appropriate fields
// (for time-series, this means a new version, not a shallow copy)
type Cloneable interface {
	Clone() Cloneable
}

// TimeSeriesEntity defines the interface for any entity that supports time-series versioning
// Now embeds Cloneable
type TimeSeriesEntity interface {
	Cloneable
	// GetID returns the unique identifier of the entity
	GetID() uuid.UUID

	// GetCreationTime returns when this version was created
	GetCreationTime() time.Time

	// GetExpirationTime returns when this version expires (null means current version)
	GetExpirationTime() *time.Time

	// SetExpirationTime sets the expiration time for this version
	SetExpirationTime(expirationTime *time.Time)

	// IsCurrentVersion returns true if this is the current version (expiration time is null)
	IsCurrentVersion() bool
}

// TimeSeriesRepository defines the interface for repository operations on time-series entities
type TimeSeriesRepository[T TimeSeriesEntity] interface {
	// Create creates a new version of an entity
	Create(ctx interface{}, entity T) error

	// GetCurrentByID retrieves the current version of an entity by its logical ID
	GetCurrentByID(ctx interface{}, id uuid.UUID) (T, error)

	// GetByID retrieves a specific version of an entity by its version ID
	GetByID(ctx interface{}, id uuid.UUID) (T, error)

	// GetHistory retrieves all versions of an entity by its logical ID
	GetHistory(ctx interface{}, id uuid.UUID) ([]T, error)

	// Update creates a new version by cloning the current version and applying updates
	Update(ctx interface{}, id uuid.UUID, updates func(T) error) (T, error)

	// Archive marks the current version as expired (archives it)
	Archive(ctx interface{}, id uuid.UUID) error

	// HardDelete permanently removes all versions of an entity
	HardDelete(ctx interface{}, id uuid.UUID) error
}

// TimeSeriesService defines the interface for service operations on time-series entities
type TimeSeriesService[T TimeSeriesEntity] interface {
	// Create creates a new entity
	Create(ctx interface{}, entity T) (T, error)

	// GetCurrent retrieves the current version of an entity
	GetCurrent(ctx interface{}, id uuid.UUID) (T, error)

	// Get retrieves a specific version of an entity
	Get(ctx interface{}, id uuid.UUID) (T, error)

	// GetHistory retrieves all versions of an entity
	GetHistory(ctx interface{}, id uuid.UUID) ([]T, error)

	// Update creates a new version with the provided updates
	Update(ctx interface{}, id uuid.UUID, updates func(T) error) (T, error)

	// Archive archives the current version
	Archive(ctx interface{}, id uuid.UUID) error
}

// BaseTimeSeriesEntity provides a base implementation of TimeSeriesEntity
type BaseTimeSeriesEntity struct {
	ID             uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	CreationTime   time.Time  `json:"creation_time" gorm:"autoCreateTime;not null"`
	ExpirationTime *time.Time `json:"expiration_time" gorm:"index"`
}

// GetID returns the unique identifier of the entity
func (b *BaseTimeSeriesEntity) GetID() uuid.UUID {
	return b.ID
}

// GetCreationTime returns when this version was created
func (b *BaseTimeSeriesEntity) GetCreationTime() time.Time {
	return b.CreationTime
}

// GetExpirationTime returns when this version expires (null means current version)
func (b *BaseTimeSeriesEntity) GetExpirationTime() *time.Time {
	return b.ExpirationTime
}

// SetExpirationTime sets the expiration time for this version
func (b *BaseTimeSeriesEntity) SetExpirationTime(expirationTime *time.Time) {
	b.ExpirationTime = expirationTime
}

// IsCurrentVersion returns true if this is the current version (expiration time is null)
func (b *BaseTimeSeriesEntity) IsCurrentVersion() bool {
	return b.ExpirationTime == nil
}

// Clone creates a new version of this entity with a new ID and current creation time
func (b *BaseTimeSeriesEntity) Clone() Cloneable {
	// This is a base implementation - specific entities should override this
	// to properly clone their specific fields
	return &BaseTimeSeriesEntity{
		CreationTime:   time.Now(),
		ExpirationTime: nil, // New version is current
	}
}
