package grpc

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/pkg/validator"

	pb "github.com/hinha/library-management-synapsis/gen/api/proto/book"
	domain "github.com/hinha/library-management-synapsis/internal/domain/book"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BookHandler implements the BookService gRPC interface
type BookHandler struct {
	pb.BookServiceServer
	service domain.Service
}

// NewBookHandler creates a new BookHandler
func NewBookHandler(service domain.Service) *BookHandler {
	return &BookHandler{
		service: service,
	}
}

// Create handles book creation
func (h *BookHandler) Create(ctx context.Context, req *pb.CreateBookRequest) (*pb.BookResponse, error) {
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	book, err := h.service.CreateBook(ctx, req.GetTitle(), req.GetAuthor(), req.GetCategory(), req.GetStock())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to create book")
	}

	return book.ToProto(), nil
}

// ListBooks handles listing all books
func (h *BookHandler) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	books, err := h.service.ListBooks(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list books")
	}

	response := &pb.ListBooksResponse{
		Books: make([]*pb.BookResponse, len(books)),
	}

	for i, book := range books {
		response.Books[i] = book.ToProto()
	}

	return response, nil
}

// GetBook handles retrieving a book by ID
func (h *BookHandler) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.BookResponse, error) {
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	book, err := h.service.GetBook(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to get book")
	}

	return book.ToProto(), nil
}

// Recommend handles book recommendations
func (h *BookHandler) Recommend(ctx context.Context, req *pb.RecommendRequest) (*pb.ListBooksResponse, error) {
	books, err := h.service.RecommendBooks(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to recommend books")
	}

	response := &pb.ListBooksResponse{
		Books: make([]*pb.BookResponse, len(books)),
	}

	for i, book := range books {
		response.Books[i] = book.ToProto()
	}

	return response, nil
}

func (h *BookHandler) HealthCheck(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return h.service.Health(ctx)
}
