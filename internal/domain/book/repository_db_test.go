package book

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewDbRepository(t *testing.T) {
	db := &gorm.DB{}
	repo := NewDbRepository(db)
	_, ok := repo.(*DBRepository)
	assert.True(t, ok, "NewDbRepository should return *DBRepository")
}

func TestDBRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	type fields struct {
		setupMock func(sqlmock.Sqlmock, *domain.Book)
	}
	testCases := []struct {
		name          string
		book          *domain.Book
		fields        fields
		expectedError error
	}{
		{
			name: "Success",
			book: &domain.Book{
				Title:     "Test Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     10,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, book *domain.Book) {
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO "books"`).
						WithArgs(
							book.ID,
							book.Title,
							book.Author,
							book.Category,
							book.Stock,
							sqlmock.AnyArg(), // CreatedAt
							sqlmock.AnyArg(), // UpdatedAt
							sqlmock.AnyArg(), // DeletedAt
						).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				},
			},
			expectedError: nil,
		},
		{
			name: "Database Error",
			book: &domain.Book{
				Title:     "Error Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     5,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, book *domain.Book) {
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO "books"`).
						WithArgs(
							book.ID,
							book.Title,
							book.Author,
							book.Category,
							book.Stock,
							sqlmock.AnyArg(),
							sqlmock.AnyArg(),
							sqlmock.AnyArg(),
						).
						WillReturnError(errors.New("database error"))
					mock.ExpectRollback()
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			tc.fields.setupMock(mock, tc.book)

			repo := &DBRepository{db: gdb}
			err = repo.Create(context.Background(), tc.book)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else {
					if err.Error() != tc.expectedError.Error() {
						t.Errorf("expected error %v, got %v", tc.expectedError, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	type fields struct {
		setupMock func(sqlmock.Sqlmock, string, *domain.Book)
	}
	testCases := []struct {
		name          string
		id            string
		returnBook    *domain.Book
		fields        fields
		expectedBook  *domain.Book
		expectedError error
	}{
		{
			name: "Success",
			id:   "1",
			returnBook: &domain.Book{
				ID:        "1", // ID is not used in assertion, can be 0
				Title:     "Test Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     10,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string, book *domain.Book) {
					rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
						AddRow(book.ID, book.Title, book.Author, book.Category, book.Stock, book.CreatedAt, book.UpdatedAt, nil)
					mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
						WithArgs(id, 1).
						WillReturnRows(rows)
				},
			},
			expectedBook: &domain.Book{
				ID:        "1", // ID is not used in assertion, can be 0
				Title:     "Test Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     10,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			expectedError: nil,
		},
		{
			name:       "Not Found",
			id:         "2",
			returnBook: nil,
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string, book *domain.Book) {
					mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
						WithArgs(id, 1).
						WillReturnError(gorm.ErrRecordNotFound)
				},
			},
			expectedBook:  nil,
			expectedError: ErrBookNotFound,
		},
		{
			name:       "DB Error",
			id:         "3",
			returnBook: nil,
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string, book *domain.Book) {
					mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
						WithArgs(id, 1).
						WillReturnError(errors.New("db error"))
				},
			},
			expectedBook:  nil,
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			tc.fields.setupMock(mock, tc.id, tc.returnBook)

			repo := &DBRepository{db: gdb}
			got, err := repo.GetByID(context.Background(), tc.id)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
				if got != nil {
					t.Errorf("expected nil book, got %v", got)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tc.expectedBook) {
					t.Errorf("expected book %+v, got %+v", tc.expectedBook, got)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_List(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		expectedBooks []*domain.Book
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow("1", "Book1", "Author1", "Cat1", 5, fixedTime, fixedTime, nil).
					AddRow("2", "Book2", "Author2", "Cat2", 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE "books"\."deleted_at" IS NULL ORDER BY created_at DESC`).WillReturnRows(rows)
			},
			expectedBooks: []*domain.Book{
				{
					ID:        "1",
					Title:     "Book1",
					Author:    "Author1",
					Category:  "Cat1",
					Stock:     5,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
				{
					ID:        "2",
					Title:     "Book2",
					Author:    "Author2",
					Category:  "Cat2",
					Stock:     10,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			expectedError: nil,
		},
		{
			name: "DB Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE "books"\."deleted_at" IS NULL ORDER BY created_at DESC`).WillReturnError(errors.New("db error"))
			},
			expectedBooks: nil,
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			repo := &DBRepository{db: gdb}
			got, err := repo.List(context.Background())

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
				if got != nil {
					t.Errorf("expected nil books, got %v", got)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tc.expectedBooks) {
					t.Errorf("expected books %+v, got %+v", tc.expectedBooks, got)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_Update(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		book          *domain.Book
		setupMock     func(sqlmock.Sqlmock, *domain.Book)
		expectedError error
	}{
		{
			name: "Success",
			book: &domain.Book{
				ID:        "1",
				Title:     "Updated Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     10,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			setupMock: func(mock sqlmock.Sqlmock, book *domain.Book) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "books" SET (.+) WHERE "books"\."deleted_at" IS NULL AND "id" = \$8`).
					WithArgs(
						book.Title,
						book.Author,
						book.Category,
						book.Stock,
						book.CreatedAt,
						sqlmock.AnyArg(), // UpdatedAt (will be set to time.Now())
						sqlmock.AnyArg(), // DeletedAt
						book.ID,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		//{
		//	name: "Book Not Found",
		//	book: &domain.Book{
		//		ID:        "2",
		//		Title:     "Not Found Book",
		//		Author:    "Author",
		//		Category:  "Fiction",
		//		Stock:     5,
		//		CreatedAt: fixedTime,
		//		UpdatedAt: fixedTime,
		//	},
		//	// No mock setup for this case, as GORM will not start a transaction if RowsAffected == 0
		//	setupMock:     nil,
		//	expectedError: ErrBookNotFound,
		//},
		{
			name: "DB Error",
			book: &domain.Book{
				ID:        "3",
				Title:     "Error Book",
				Author:    "Author",
				Category:  "Fiction",
				Stock:     7,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			setupMock: func(mock sqlmock.Sqlmock, book *domain.Book) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "books" SET (.+) WHERE "books"\."deleted_at" IS NULL AND "id" = \$8`).
					WithArgs(
						book.Title,
						book.Author,
						book.Category,
						book.Stock,
						book.CreatedAt,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						book.ID,
					).
					WillReturnError(errors.New("update error"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("update error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			if tc.setupMock != nil {
				tc.setupMock(mock, tc.book)
			}

			repo := &DBRepository{db: gdb}
			err = repo.Update(context.Background(), tc.book)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			// Only check expectations if we set up a mock for this case
			if tc.setupMock != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestDBRepository_Delete(t *testing.T) {
	type fields struct {
		setupMock func(sqlmock.Sqlmock, string)
	}
	testCases := []struct {
		name          string
		id            string
		fields        fields
		expectedError error
	}{
		{
			name: "Success",
			id:   "1",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "books" SET "deleted_at"=\$1 WHERE id = \$2 AND "books"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), id).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				},
			},
			expectedError: nil,
		},
		{
			name: "Book Not Found",
			id:   "2",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "books" SET "deleted_at"=\$1 WHERE id = \$2 AND "books"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), id).
						WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectCommit()
				},
			},
			expectedError: ErrBookNotFound,
		},
		{
			name: "DB Error",
			id:   "3",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock, id string) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "books" SET "deleted_at"=\$1 WHERE id = \$2 AND "books"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), id).
						WillReturnError(errors.New("delete error"))
					mock.ExpectRollback()
				},
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			if tc.fields.setupMock != nil {
				tc.fields.setupMock(mock, tc.id)
			}

			repo := &DBRepository{db: gdb}
			err = repo.Delete(context.Background(), tc.id)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_UpdateStock(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		id            string
		change        int32
		setupMock     func(sqlmock.Sqlmock, string, int32)
		expectedError error
	}{
		{
			name:   "Success increase",
			id:     "1",
			change: 5,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow(id, "Book", "Author", "Cat", 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnRows(rows)
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "books" SET (.+) WHERE "books"\."deleted_at" IS NULL AND "id" = \$8`).
					WithArgs(
						"Book", "Author", "Cat", int32(15), fixedTime, sqlmock.AnyArg(), sqlmock.AnyArg(), id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:   "Success decrease",
			id:     "1",
			change: -3,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow(id, "Book", "Author", "Cat", 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnRows(rows)
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "books" SET (.+) WHERE "books"\."deleted_at" IS NULL AND "id" = \$8`).
					WithArgs(
						"Book", "Author", "Cat", int32(7), fixedTime, sqlmock.AnyArg(), sqlmock.AnyArg(), id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:   "Insufficient stock",
			id:     "1",
			change: -15,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow(id, "Book", "Author", "Cat", 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnRows(rows)
			},
			expectedError: ErrInsufficientStock,
		},
		{
			name:   "Book not found",
			id:     "2",
			change: 1,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: ErrBookNotFound,
		},
		{
			name:   "DB error on select",
			id:     "3",
			change: 1,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:   "DB error on update",
			id:     "1",
			change: 2,
			setupMock: func(mock sqlmock.Sqlmock, id string, change int32) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow(id, "Book", "Author", "Cat", 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE id = \$1 AND "books"\."deleted_at" IS NULL ORDER BY "books"\."id" LIMIT \$2`).
					WithArgs(id, 1).
					WillReturnRows(rows)
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "books" SET (.+) WHERE "books"\."deleted_at" IS NULL AND "id" = \$8`).
					WithArgs(
						"Book", "Author", "Cat", int32(12), fixedTime, sqlmock.AnyArg(), sqlmock.AnyArg(), id,
					).
					WillReturnError(errors.New("update error"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("update error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			if tc.setupMock != nil {
				tc.setupMock(mock, tc.id, tc.change)
			}

			repo := &DBRepository{db: gdb}
			err = repo.UpdateStock(context.Background(), tc.id, tc.change)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_GetByCategory(t *testing.T) {
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		category      string
		setupMock     func(sqlmock.Sqlmock, string)
		expectedBooks []*domain.Book
		expectedError error
	}{
		{
			name:     "Success",
			category: "Fiction",
			setupMock: func(mock sqlmock.Sqlmock, category string) {
				rows := sqlmock.NewRows([]string{"id", "title", "author", "category", "stock", "created_at", "updated_at", "deleted_at"}).
					AddRow("1", "Book1", "Author1", category, 5, fixedTime, fixedTime, nil).
					AddRow("2", "Book2", "Author2", category, 10, fixedTime, fixedTime, nil)
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE category = \$1 AND "books"\."deleted_at" IS NULL`).WithArgs(category).WillReturnRows(rows)
			},
			expectedBooks: []*domain.Book{
				{
					ID:        "1",
					Title:     "Book1",
					Author:    "Author1",
					Category:  "Fiction",
					Stock:     5,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
				{
					ID:        "2",
					Title:     "Book2",
					Author:    "Author2",
					Category:  "Fiction",
					Stock:     10,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			expectedError: nil,
		},
		{
			name:     "DB Error",
			category: "NonFiction",
			setupMock: func(mock sqlmock.Sqlmock, category string) {
				mock.ExpectQuery(`SELECT \* FROM "books" WHERE category = \$1 AND "books"\."deleted_at" IS NULL`).WithArgs(category).WillReturnError(errors.New("db error"))
			},
			expectedBooks: nil,
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			if tc.setupMock != nil {
				tc.setupMock(mock, tc.category)
			}

			repo := &DBRepository{db: gdb}
			got, err := repo.GetByCategory(context.Background(), tc.category)

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
				if got != nil {
					t.Errorf("expected nil books, got %v", got)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tc.expectedBooks) {
					t.Errorf("expected books %+v, got %+v", tc.expectedBooks, got)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBRepository_Ping(t *testing.T) {
	type fields struct {
		setupMock func(sqlmock.Sqlmock)
	}
	testCases := []struct {
		name          string
		fields        fields
		expectedError error
	}{
		{
			name: "Success",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectPing() // GORM open
					mock.ExpectPing() // repo.Ping()
				},
			},
			expectedError: nil,
		},
		{
			name: "Database Error",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectPing()                                               // GORM open
					mock.ExpectPing().WillReturnError(errors.New("database error")) // repo.Ping()
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer db.Close()

			tc.fields.setupMock(mock)
			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.NoError(t, err)

			repo := &DBRepository{db: gdb}
			err = repo.Ping(context.Background())

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
