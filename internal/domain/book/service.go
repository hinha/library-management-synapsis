package book

import (
	"context"
	"errors"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/book"
	"github.com/hinha/library-management-synapsis/internal/domain"
)

var (
	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")
)

// Service defines the interface for book business logic
//
//go:generate mockery --name=IService --output=../../delivery/mocks --outpkg=mocks
type Service interface {
	CreateBook(ctx context.Context, title, author, category string, stock int32) (*domain.Book, error)
	GetBook(ctx context.Context, id string) (*domain.Book, error)
	ListBooks(ctx context.Context) ([]*domain.Book, error)
	UpdateBook(ctx context.Context, id, title, author, category string, stock int32) (*domain.Book, error)
	DeleteBook(ctx context.Context, id string) error
	RecommendBooks(ctx context.Context) ([]*domain.Book, error)
	Health(ctx context.Context) (*pb.HealthCheckResponse, error)
}

// DefaultService implements Service
type DefaultService struct {
	repoDb IDbRepository
}

// NewService creates a new DefaultService
func NewService(repo IDbRepository) *DefaultService {
	return &DefaultService{
		repoDb: repo,
	}
}

// CreateBook creates a new book
func (s *DefaultService) CreateBook(ctx context.Context, title, author, category string, stock int32) (*domain.Book, error) {
	book := domain.NewBook(title, author, category, stock)
	if err := s.repoDb.Create(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

// GetBook retrieves a book by ID
func (s *DefaultService) GetBook(ctx context.Context, id string) (*domain.Book, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	return s.repoDb.GetByID(ctx, id)
}

// ListBooks retrieves all books
func (s *DefaultService) ListBooks(ctx context.Context) ([]*domain.Book, error) {
	return s.repoDb.List(ctx)
}

// UpdateBook updates a book
func (s *DefaultService) UpdateBook(ctx context.Context, id, title, author, category string, stock int32) (*domain.Book, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	book, err := s.repoDb.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		book.Title = title
	}

	if author != "" {
		book.Author = author
	}

	if category != "" {
		book.Category = category
	}

	if stock >= 0 {
		book.Stock = stock
	}

	if err := s.repoDb.Update(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

// DeleteBook deletes a book
func (s *DefaultService) DeleteBook(ctx context.Context, id string) error {
	return s.repoDb.Delete(ctx, id)
}

// RecommendBooks recommends books based on popularity or other criteria
// This is a simple implementation that just returns all books
// In a real application, this would use more sophisticated logic
func (s *DefaultService) RecommendBooks(ctx context.Context) ([]*domain.Book, error) {
	// For now, just return all books
	// In a real application, this would use more sophisticated logic
	return s.repoDb.List(ctx)
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
