package transaction

import (
	"context"
	"errors"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/transaction"
	"github.com/hinha/library-management-synapsis/internal/domain"

	"github.com/hinha/library-management-synapsis/internal/domain/book"
)

var (
	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")
	// ErrBookNotAvailable is returned when a book is not available for borrowing
	ErrBookNotAvailable = errors.New("book not available")
)

// Service defines the interface for transaction business logic
type Service interface {
	BorrowBook(ctx context.Context, userID, bookID string) (*domain.Transaction, error)
	ReturnBook(ctx context.Context, transactionID string) (*domain.Transaction, error)
	GetUserHistory(ctx context.Context, userID string) ([]*domain.Transaction, error)
	Health(ctx context.Context) (*pb.HealthCheckResponse, error)
}

// BookRepository defines the interface for book operations needed by the transaction service
type BookRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	UpdateStock(ctx context.Context, id string, change int32) error
}

// DefaultService implements Service
type DefaultService struct {
	repoDb   IDbRepository
	bookRepo BookRepository
}

// NewService creates a new DefaultService
func NewService(repo IDbRepository, bookRepo BookRepository) *DefaultService {
	return &DefaultService{
		repoDb:   repo,
		bookRepo: bookRepo,
	}
}

// BorrowBook creates a new transaction for borrowing a book
func (s *DefaultService) BorrowBook(ctx context.Context, userID, bookID string) (*domain.Transaction, error) {
	if userID == "" || bookID == "" {
		return nil, ErrInvalidInput
	}

	// Check if book exists and has available stock
	b, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, book.ErrBookNotFound) {
			return nil, book.ErrBookNotFound
		}
		return nil, err
	}

	if b.Stock <= 0 {
		return nil, ErrBookNotAvailable
	}

	// Create transaction
	transaction := domain.NewTransaction(userID, bookID)
	if err := s.repoDb.Create(ctx, transaction); err != nil {
		return nil, err
	}

	// Update book stock
	if err := s.bookRepo.UpdateStock(ctx, bookID, -1); err != nil {
		return nil, err
	}

	return transaction, nil
}

// ReturnBook marks a transaction as returned
func (s *DefaultService) ReturnBook(ctx context.Context, transactionID string) (*domain.Transaction, error) {
	if transactionID == "" {
		return nil, ErrInvalidInput
	}

	// Get transaction
	transaction, err := s.repoDb.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Check if already returned
	if transaction.IsReturned() {
		return nil, ErrAlreadyReturned
	}

	// Mark as returned
	if err := s.repoDb.MarkAsReturned(ctx, transactionID); err != nil {
		return nil, err
	}

	// Update book stock
	if err := s.bookRepo.UpdateStock(ctx, transaction.BookID, 1); err != nil {
		return nil, err
	}

	// Get updated transaction
	return s.repoDb.GetByID(ctx, transactionID)
}

// GetUserHistory retrieves a user's transaction history
func (s *DefaultService) GetUserHistory(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	return s.repoDb.GetByUserID(ctx, userID)
}

func (s *DefaultService) Health(ctx context.Context) (*pb.HealthCheckResponse, error) {
	status := "HEALTHY"

	var componentStatus []*pb.ComponentStatus
	if err := s.repoDb.Ping(ctx); err != nil {
		status = "UNHEALTHY"
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "db",
			Status:  "DOWN",
			Message: err.Error(),
		})
	} else {
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "db",
			Status:  "UP",
			Message: "Database is healthy",
		})
	}

	return &pb.HealthCheckResponse{
		Status:     status,
		Components: componentStatus,
	}, nil
}
