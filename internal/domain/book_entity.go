package domain

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/book"
	"gorm.io/gorm"
)

// Book represents a book entity in the system
type Book struct {
	ID        string         `gorm:"primaryKey"`
	Title     string         `gorm:"not null"`
	Author    string         `gorm:"not null"`
	Category  string         `gorm:"not null"`
	Stock     int32          `gorm:"not null"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// NewBook creates a new book entity
func NewBook(title, author, category string, stock int32) *Book {
	return &Book{
		ID:        uuid.New().String(),
		Title:     title,
		Author:    author,
		Category:  category,
		Stock:     stock,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToProto converts the book entity to a protobuf book response
func (b *Book) ToProto() *pb.BookResponse {
	return &pb.BookResponse{
		Id:       b.ID,
		Title:    b.Title,
		Author:   b.Author,
		Category: b.Category,
		Stock:    b.Stock,
	}
}

// UpdateStock updates the book's stock
func (b *Book) UpdateStock(stock int32) {
	b.Stock = stock
	b.UpdatedAt = time.Now()
}
