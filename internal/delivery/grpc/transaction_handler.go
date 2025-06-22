package grpc

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/pkg/validator"

	pb "github.com/hinha/library-management-synapsis/gen/api/proto/transaction"
	"github.com/hinha/library-management-synapsis/internal/domain/book"
	domain "github.com/hinha/library-management-synapsis/internal/domain/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TransactionHandler implements the TransactionService gRPC interface
type TransactionHandler struct {
	pb.TransactionServiceServer
	service domain.Service
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(service domain.Service) *TransactionHandler {
	return &TransactionHandler{
		service: service,
	}
}

// Borrow handles book borrowing
func (h *TransactionHandler) Borrow(ctx context.Context, req *pb.BorrowRequest) (*pb.TransactionResponse, error) {
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	transaction, err := h.service.BorrowBook(ctx, req.GetUserId(), req.GetBookId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		case errors.Is(err, book.ErrBookNotFound):
			return nil, status.Error(codes.NotFound, "book not found")
		case errors.Is(err, domain.ErrBookNotAvailable):
			return nil, status.Error(codes.FailedPrecondition, "book not available")
		default:
			return nil, status.Error(codes.Internal, "failed to borrow book")
		}
	}

	return transaction.ToProto(), nil
}

// Return handles book returning
func (h *TransactionHandler) Return(ctx context.Context, req *pb.ReturnRequest) (*pb.TransactionResponse, error) {
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	transaction, err := h.service.ReturnBook(ctx, req.GetTransactionId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		case errors.Is(err, domain.ErrTransactionNotFound):
			return nil, status.Error(codes.NotFound, "transaction not found")
		case errors.Is(err, domain.ErrAlreadyReturned):
			return nil, status.Error(codes.FailedPrecondition, "book already returned")
		default:
			return nil, status.Error(codes.Internal, "failed to return book")
		}
	}

	return transaction.ToProto(), nil
}

// History handles retrieving a user's transaction history
func (h *TransactionHandler) History(ctx context.Context, req *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	transactions, err := h.service.GetUserHistory(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to get transaction history")
	}

	response := &pb.HistoryResponse{
		Transactions: make([]*pb.TransactionResponse, len(transactions)),
	}

	for i, transaction := range transactions {
		response.Transactions[i] = transaction.ToProto()
	}

	return response, nil
}
