package book

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/book"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/book/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestDefaultService_CreateBook(t *testing.T) {
	type args struct {
		ctx      context.Context
		title    string
		author   string
		category string
		stock    int32
	}
	tests := []struct {
		name    string
		args    args
		mockFn  func(repo *mocks.IDbRepository)
		want    *domain.Book
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				title:    "Book Title",
				author:   "Author",
				category: "Fiction",
				stock:    10,
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
					return b.Title == "Book Title" &&
						b.Author == "Author" &&
						b.Category == "Fiction" &&
						b.Stock == 10
				})).Return(nil)
			},
			want: func() *domain.Book {
				return domain.NewBook("Book Title", "Author", "Fiction", 10)
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "repo returns error",
			args: args{
				ctx:      context.Background(),
				title:    "Book Title",
				author:   "Author",
				category: "Fiction",
				stock:    10,
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
					return b.Title == "Book Title" &&
						b.Author == "Author" &&
						b.Category == "Fiction" &&
						b.Stock == 10
				})).Return(errors.New("db error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.CreateBook(tt.args.ctx, tt.args.title, tt.args.author, tt.args.category, tt.args.stock)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateBook(%v, %v, %v, %v, %v)", tt.args.ctx, tt.args.title, tt.args.author, tt.args.category, tt.args.stock)) {
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Author, got.Author)
				assert.Equal(t, tt.want.Category, got.Category)
				assert.Equal(t, tt.want.Stock, got.Stock)
			} else {
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_GetBook(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		mockFn  func(repo *mocks.IDbRepository)
		want    *domain.Book
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				id:  "book-id-1",
			},
			mockFn: func(repo *mocks.IDbRepository) {
				book := &domain.Book{
					ID:       "book-id-1",
					Title:    "Book Title",
					Author:   "Author",
					Category: "Fiction",
					Stock:    10,
				}
				repo.On("GetByID", mock.Anything, "book-id-1").Return(book, nil)
			},
			want: &domain.Book{
				ID:       "book-id-1",
				Title:    "Book Title",
				Author:   "Author",
				Category: "Fiction",
				Stock:    10,
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid input - empty id",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			mockFn:  nil,
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "repo returns error",
			args: args{
				ctx: context.Background(),
				id:  "book-id-2",
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("GetByID", mock.Anything, "book-id-2").Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.GetBook(tt.args.ctx, tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("GetBook(%v, %v)", tt.args.ctx, tt.args.id)) {
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Author, got.Author)
				assert.Equal(t, tt.want.Category, got.Category)
				assert.Equal(t, tt.want.Stock, got.Stock)
			} else {
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestNewService(t *testing.T) {
	repo := new(mocks.IDbRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
	assert.Equal(t, repo, svc.repoDb)
}

func TestDefaultService_ListBooks(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(repo *mocks.IDbRepository)
		want    []*domain.Book
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			mockFn: func(repo *mocks.IDbRepository) {
				books := []*domain.Book{
					{ID: "1", Title: "Book1", Author: "Author1", Category: "Fiction", Stock: 5},
					{ID: "2", Title: "Book2", Author: "Author2", Category: "NonFiction", Stock: 3},
				}
				repo.On("List", mock.Anything).Return(books, nil)
			},
			want: []*domain.Book{
				{ID: "1", Title: "Book1", Author: "Author1", Category: "Fiction", Stock: 5},
				{ID: "2", Title: "Book2", Author: "Author2", Category: "NonFiction", Stock: 3},
			},
			wantErr: assert.NoError,
		},
		{
			name: "repo returns error",
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("List", mock.Anything).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.ListBooks(context.Background())
			if !tt.wantErr(t, err, fmt.Sprintf("ListBooks(%v)", context.Background())) {
				return
			}
			assert.Equalf(t, tt.want, got, "ListBooks(%v)", context.Background())
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_UpdateBook(t *testing.T) {
	type args struct {
		ctx      context.Context
		id       string
		title    string
		author   string
		category string
		stock    int32
	}
	tests := []struct {
		name    string
		args    args
		mockFn  func(repo *mocks.IDbRepository)
		want    *domain.Book
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success update all fields",
			args: args{
				ctx:      context.Background(),
				id:       "book-id-1",
				title:    "New Title",
				author:   "New Author",
				category: "New Category",
				stock:    20,
			},
			mockFn: func(repo *mocks.IDbRepository) {
				origBook := &domain.Book{
					ID:       "book-id-1",
					Title:    "Old Title",
					Author:   "Old Author",
					Category: "Old Category",
					Stock:    10,
				}
				repo.On("GetByID", mock.Anything, "book-id-1").Return(origBook, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
					return b.ID == "book-id-1" &&
						b.Title == "New Title" &&
						b.Author == "New Author" &&
						b.Category == "New Category" &&
						b.Stock == 20
				})).Return(nil)
			},
			want: &domain.Book{
				ID:       "book-id-1",
				Title:    "New Title",
				Author:   "New Author",
				Category: "New Category",
				Stock:    20,
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid input - empty id",
			args: args{
				ctx:      context.Background(),
				id:       "",
				title:    "Title",
				author:   "Author",
				category: "Category",
				stock:    5,
			},
			mockFn:  nil,
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "repo GetByID returns error",
			args: args{
				ctx:      context.Background(),
				id:       "book-id-2",
				title:    "Title",
				author:   "Author",
				category: "Category",
				stock:    5,
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("GetByID", mock.Anything, "book-id-2").Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "repo Update returns error",
			args: args{
				ctx:      context.Background(),
				id:       "book-id-3",
				title:    "Title",
				author:   "Author",
				category: "Category",
				stock:    5,
			},
			mockFn: func(repo *mocks.IDbRepository) {
				origBook := &domain.Book{
					ID:       "book-id-3",
					Title:    "Old Title",
					Author:   "Old Author",
					Category: "Old Category",
					Stock:    1,
				}
				repo.On("GetByID", mock.Anything, "book-id-3").Return(origBook, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Book")).Return(errors.New("update error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "partial update (only title)",
			args: args{
				ctx:      context.Background(),
				id:       "book-id-4",
				title:    "Partial Title",
				author:   "",
				category: "",
				stock:    -1, // should not update stock
			},
			mockFn: func(repo *mocks.IDbRepository) {
				origBook := &domain.Book{
					ID:       "book-id-4",
					Title:    "Old Title",
					Author:   "Old Author",
					Category: "Old Category",
					Stock:    7,
				}
				repo.On("GetByID", mock.Anything, "book-id-4").Return(origBook, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
					return b.ID == "book-id-4" &&
						b.Title == "Partial Title" &&
						b.Author == "Old Author" &&
						b.Category == "Old Category" &&
						b.Stock == 7
				})).Return(nil)
			},
			want: &domain.Book{
				ID:       "book-id-4",
				Title:    "Partial Title",
				Author:   "Old Author",
				Category: "Old Category",
				Stock:    7,
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.UpdateBook(tt.args.ctx, tt.args.id, tt.args.title, tt.args.author, tt.args.category, tt.args.stock)
			if !tt.wantErr(t, err, fmt.Sprintf("UpdateBook(%v, %v, %v, %v, %v, %v)", tt.args.ctx, tt.args.id, tt.args.title, tt.args.author, tt.args.category, tt.args.stock)) {
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Author, got.Author)
				assert.Equal(t, tt.want.Category, got.Category)
				assert.Equal(t, tt.want.Stock, got.Stock)
			} else {
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_DeleteBook(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		mockFn  func(repo *mocks.IDbRepository)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				id:  "book-id-1",
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Delete", mock.Anything, "book-id-1").Return(nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "repo returns error",
			args: args{
				ctx: context.Background(),
				id:  "book-id-2",
			},
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Delete", mock.Anything, "book-id-2").Return(errors.New("db error"))
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			tt.wantErr(t, s.DeleteBook(tt.args.ctx, tt.args.id), fmt.Sprintf("DeleteBook(%v, %v)", tt.args.ctx, tt.args.id))
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_RecommendBooks(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(repo *mocks.IDbRepository)
		want    []*domain.Book
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			mockFn: func(repo *mocks.IDbRepository) {
				books := []*domain.Book{
					{ID: "1", Title: "Book1", Author: "Author1", Category: "Fiction", Stock: 5},
					{ID: "2", Title: "Book2", Author: "Author2", Category: "NonFiction", Stock: 3},
				}
				repo.On("List", mock.Anything).Return(books, nil)
			},
			want: []*domain.Book{
				{ID: "1", Title: "Book1", Author: "Author1", Category: "Fiction", Stock: 5},
				{ID: "2", Title: "Book2", Author: "Author2", Category: "NonFiction", Stock: 3},
			},
			wantErr: assert.NoError,
		},
		{
			name: "repo returns error",
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("List", mock.Anything).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.RecommendBooks(context.Background())
			if !tt.wantErr(t, err, fmt.Sprintf("RecommendBooks(%v)", context.Background())) {
				return
			}
			assert.Equalf(t, tt.want, got, "RecommendBooks(%v)", context.Background())
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_Health(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(repo *mocks.IDbRepository)
		want    *pb.HealthCheckResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "healthy db",
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Ping", mock.Anything).Return(nil)
			},
			want: &pb.HealthCheckResponse{
				Status: "HEALTHY",
				Components: []*pb.ComponentStatus{
					{
						Name:    "db",
						Status:  "UP",
						Message: "Database is healthy",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "unhealthy db",
			mockFn: func(repo *mocks.IDbRepository) {
				repo.On("Ping", mock.Anything).Return(errors.New("db down"))
			},
			want: &pb.HealthCheckResponse{
				Status: "UNHEALTHY",
				Components: []*pb.ComponentStatus{
					{
						Name:    "db",
						Status:  "DOWN",
						Message: "db down",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			if tt.mockFn != nil {
				tt.mockFn(repo)
			}
			s := &DefaultService{
				repoDb: repo,
			}
			got, err := s.Health(context.Background())
			if !tt.wantErr(t, err, fmt.Sprintf("Health(%v)", context.Background())) {
				return
			}
			assert.Equal(t, tt.want.Status, got.Status)
			assert.Equal(t, len(tt.want.Components), len(got.Components))
			for i := range tt.want.Components {
				assert.Equal(t, tt.want.Components[i].Name, got.Components[i].Name)
				assert.Equal(t, tt.want.Components[i].Status, got.Components[i].Status)
				assert.Equal(t, tt.want.Components[i].Message, got.Components[i].Message)
			}
			repo.AssertExpectations(t)
		})
	}
}
