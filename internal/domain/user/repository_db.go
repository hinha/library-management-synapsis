package user

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists is returned when a user with the same email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// IDbRepository defines the interface for user data access
//
//go:generate mockery --name=IDbRepository --output=mocks --outpkg=mocks
type IDbRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	Ping(ctx context.Context) (err error)
}

// DBRepository implements IDbRepository using GORM
type DBRepository struct {
	db *gorm.DB
}

func NewDbRepository(db *gorm.DB) IDbRepository {
	return &DBRepository{db: db}
}

// Create creates a new user
func (r *DBRepository) Create(ctx context.Context, user *domain.User) error {
	// Check if user with the same email already exists
	var count int64
	if err := r.db.Model(&domain.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrEmailAlreadyExists
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *DBRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *DBRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *DBRepository) Update(ctx context.Context, user *domain.User) error {
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
func (r *DBRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *DBRepository) Ping(ctx context.Context) (err error) {
	sql, err := r.db.DB()
	if err != nil {
		return err
	}

	if err = sql.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
