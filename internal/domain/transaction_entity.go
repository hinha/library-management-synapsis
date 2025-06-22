package domain

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/transaction"
	"gorm.io/gorm"
)

// Transaction represents a book borrowing transaction
type Transaction struct {
	ID         string         `gorm:"primaryKey"`
	UserID     string         `gorm:"not null;index"`
	BookID     string         `gorm:"not null;index"`
	BorrowedAt time.Time      `gorm:"not null"`
	ReturnedAt *time.Time     `gorm:"default:null"`
	CreatedAt  time.Time      `gorm:"not null"`
	UpdatedAt  time.Time      `gorm:"not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// NewTransaction creates a new transaction
func NewTransaction(userID, bookID string) *Transaction {
	now := time.Now()
	return &Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		BookID:     bookID,
		BorrowedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// MarkAsReturned marks the transaction as returned
func (t *Transaction) MarkAsReturned() {
	now := time.Now()
	t.ReturnedAt = &now
	t.UpdatedAt = now
}

// IsReturned checks if the book has been returned
func (t *Transaction) IsReturned() bool {
	return t.ReturnedAt != nil
}

// ToProto converts the transaction entity to a protobuf transaction response
func (t *Transaction) ToProto() *pb.TransactionResponse {
	response := &pb.TransactionResponse{
		TransactionId: t.ID,
		UserId:        t.UserID,
		BookId:        t.BookID,
		BorrowedAt:    t.BorrowedAt.Format(time.RFC3339),
	}

	if t.ReturnedAt != nil {
		response.ReturnedAt = t.ReturnedAt.Format(time.RFC3339)
	}

	return response
}
