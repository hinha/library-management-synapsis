package book

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/internal/domain"

	"gorm.io/gorm"
)

var (
	// ErrBookNotFound is returned when a book is not found
	ErrBookNotFound = errors.New("book not found")
	// ErrInsufficientStock is returned when a book has insufficient stock
	ErrInsufficientStock = errors.New("insufficient stock")
)

// IDbRepository defines the interface for book data access
//
//go:generate mockery --name=IDbRepository --output=mocks --outpkg=mocks
type IDbRepository interface {
	Create(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	List(ctx context.Context) ([]*domain.Book, error)
	Update(ctx context.Context, book *domain.Book) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, change int32) error
	GetByCategory(ctx context.Context, category string) ([]*domain.Book, error)
	Ping(ctx context.Context) (err error)
}

// DBRepository implements IDbRepository using GORM
type DBRepository struct {
	db *gorm.DB
}

// NewDbRepository creates a new DBRepository
func NewDbRepository(db *gorm.DB) IDbRepository {
	return &DBRepository{db: db}
}

// Create creates a new book
func (r *DBRepository) Create(ctx context.Context, book *domain.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

// GetByID retrieves a book by ID
func (r *DBRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var book domain.Book
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&book).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return &book, nil
}

// List retrieves all books
func (r *DBRepository) List(ctx context.Context) ([]*domain.Book, error) {
	var books []*domain.Book
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// Update updates a book
func (r *DBRepository) Update(ctx context.Context, book *domain.Book) error {
	result := r.db.WithContext(ctx).Save(book)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBookNotFound
	}
	return nil
}

// Delete deletes a book
func (r *DBRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Book{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBookNotFound
	}
	return nil
}

// UpdateStock updates a book's stock
func (r *DBRepository) UpdateStock(ctx context.Context, id string, change int32) error {
	var book domain.Book
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&book).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookNotFound
		}
		return err
	}

	// Check if there's enough stock when decreasing
	if change < 0 && book.Stock < -change {
		return ErrInsufficientStock
	}

	book.Stock += change
	book.UpdatedAt = book.UpdatedAt
	return r.db.WithContext(ctx).Save(&book).Error
}

// GetByCategory retrieves books by category
func (r *DBRepository) GetByCategory(ctx context.Context, category string) ([]*domain.Book, error) {
	var books []*domain.Book
	if err := r.db.WithContext(ctx).Where("category = ?", category).Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
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
