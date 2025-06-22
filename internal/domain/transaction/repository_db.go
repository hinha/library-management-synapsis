package transaction

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/internal/domain"

	"gorm.io/gorm"
)

var (
	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrAlreadyReturned is returned when a book has already been returned
	ErrAlreadyReturned = errors.New("book already returned")
)

// IDbRepository defines the interface for transaction data access
type IDbRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Transaction, error)
	MarkAsReturned(ctx context.Context, id string) error
	Ping(ctx context.Context) (err error)
}

// DBRepository implements IDbRepository using GORM
type DBRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new DBRepository
func NewGormRepository(db *gorm.DB) IDbRepository {
	return &DBRepository{db: db}
}

// Create creates a new transaction
func (r *DBRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetByID retrieves a transaction by ID
func (r *DBRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

// GetByUserID retrieves all transactions for a user
func (r *DBRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// MarkAsReturned marks a transaction as returned
func (r *DBRepository) MarkAsReturned(ctx context.Context, id string) error {
	transaction, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if transaction.IsReturned() {
		return ErrAlreadyReturned
	}

	transaction.MarkAsReturned()
	return r.db.WithContext(ctx).Save(transaction).Error
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
