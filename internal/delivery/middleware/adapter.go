package middleware

import (
	"context"
	bookPb "github.com/hinha/library-management-synapsis/gen/api/proto/book"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/book"
	"google.golang.org/grpc"
)

// BookServiceClient is a client for the book service
type BookServiceClient struct {
	client bookPb.BookServiceClient
}

// NewBookServiceClient creates a new BookServiceClient
func NewBookServiceClient(conn *grpc.ClientConn) *BookServiceClient {
	return &BookServiceClient{
		client: bookPb.NewBookServiceClient(conn),
	}
}

// TransactionBookRepository defines a subset of book.IDbRepository needed by the transaction service
type TransactionBookRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	UpdateStock(ctx context.Context, id string, change int32) error
}

// BookRepositoryAdapter adapts the book service client to the TransactionBookRepository interface
type BookRepositoryAdapter struct {
	client *BookServiceClient
}

// NewBookRepositoryAdapter creates a new BookRepositoryAdapter
func NewBookRepositoryAdapter(client *BookServiceClient) TransactionBookRepository {
	return &BookRepositoryAdapter{
		client: client,
	}
}

// GetByID retrieves a book by ID
func (a *BookRepositoryAdapter) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	resp, err := a.client.client.GetBook(ctx, &bookPb.GetBookRequest{Id: id})
	if err != nil {
		return nil, book.ErrBookNotFound
	}

	return &domain.Book{
		ID:       resp.Id,
		Title:    resp.Title,
		Author:   resp.Author,
		Category: resp.Category,
		Stock:    resp.Stock,
	}, nil
}

// UpdateStock updates a book's stock
func (a *BookRepositoryAdapter) UpdateStock(ctx context.Context, id string, change int32) error {
	// Get current book
	b, err := a.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if there's enough stock when decreasing
	if change < 0 && b.Stock < -change {
		return book.ErrInsufficientStock
	}

	// Update book with new stock
	_, err = a.client.client.Create(ctx, &bookPb.CreateBookRequest{
		Title:    b.Title,
		Author:   b.Author,
		Category: b.Category,
		Stock:    b.Stock + change,
	})

	return err
}
