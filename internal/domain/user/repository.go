package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists is returned when a user with the same email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// IRepository defines the interface for user data access
type IRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// Repository implements IRepository using GORM
type Repository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IRepository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	// Check if user with the same email already exists
	var count int64
	if err := r.db.Model(&User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrEmailAlreadyExists
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *Repository) Update(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete deletes a user
func (r *Repository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
