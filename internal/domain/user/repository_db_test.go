package user

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

//// MockRepository is a mock implementation of IDbRepository for testing
//type MockRepository struct {
//	mock.Mock
//}
//
//// Create mocks the Create method
//func (m *MockRepository) Create(ctx context.Context, user *User) error {
//	args := m.Called(ctx, user)
//	return args.Error(0)
//}
//
//// GetByID mocks the GetByID method
//func (m *MockRepository) GetByID(ctx context.Context, id string) (*User, error) {
//	args := m.Called(ctx, id)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*User), args.Error(1)
//}
//
//// GetByEmail mocks the GetByEmail method
//func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
//	args := m.Called(ctx, email)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*User), args.Error(1)
//}
//
//// Update mocks the Update method
//func (m *MockRepository) Update(ctx context.Context, user *User) error {
//	args := m.Called(ctx, user)
//	return args.Error(0)
//}
//
//// Delete mocks the Delete method
//func (m *MockRepository) Delete(ctx context.Context, id string) error {
//	args := m.Called(ctx, id)
//	return args.Error(0)
//}
//
//// Ping mocks the Ping method
//func (m *MockRepository) Ping(ctx context.Context) error {
//	args := m.Called(ctx)
//	return args.Error(0)
//}

func TestDBRepository_Create(t *testing.T) {
	type fields struct {
		setupMock func(sqlmock.Sqlmock)
	}
	testCases := []struct {
		name          string
		user          *domain.User
		fields        fields
		expectedError error
	}{
		{
			name: "Success",
			user: &domain.User{Email: "test@example.com", Name: "Test User"},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT count(.*) FROM "users" WHERE email = .*`).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
					mock.ExpectBegin()
					mock.ExpectQuery(`INSERT INTO "users"`).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
					mock.ExpectCommit()
				},
			},
			expectedError: nil,
		},
		{
			name: "Email Already Exists",
			user: &domain.User{Email: "existing@example.com", Name: "Existing User"},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT count(.*) FROM "users" WHERE email = .*`).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				},
			},
			expectedError: ErrEmailAlreadyExists,
		},
		{
			name: "Database Error on Count",
			user: &domain.User{Email: "errorcount@example.com", Name: "Error Count"},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT count(.*) FROM "users" WHERE email = .*`).
						WillReturnError(errors.New("database error"))
				},
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "Database Error on Create",
			user: &domain.User{Email: "errorcreate@example.com", Name: "Error Create"},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT count(.*) FROM "users" WHERE email = .*`).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
					mock.ExpectBegin()
					mock.ExpectQuery(`INSERT INTO "users"`).
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
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.NoError(t, err)

			tc.fields.setupMock(mock)

			repo := &DBRepository{db: gdb}
			err = repo.Create(context.Background(), tc.user)

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

func TestDBRepository_GetByID(t *testing.T) {
	// Create a fixed timestamp for testing
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	type fields struct {
		setupMock func(sqlmock.Sqlmock)
	}
	testCases := []struct {
		name          string
		id            string
		fields        fields
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name: "Success",
			id:   "1",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "active", "created_at", "updated_at", "deleted_at"}).
						AddRow("1", "test@example.com", "Test User", "", "", false, fixedTime, fixedTime, nil)
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("1", 1).
						WillReturnRows(rows)
				},
			},
			expectedUser: &domain.User{
				ID:        1,
				Email:     "test@example.com",
				Name:      "Test User",
				Password:  "",
				Role:      "",
				Active:    false,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			expectedError: nil,
		},
		{
			name: "User Not Found",
			id:   "2",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("2", 1).
						WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "active", "created_at", "updated_at", "deleted_at"}))
				},
			},
			expectedUser:  nil,
			expectedError: ErrUserNotFound,
		},
		{
			name: "Database Error",
			id:   "3",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("3", 1).
						WillReturnError(errors.New("database error"))
				},
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.NoError(t, err)

			tc.fields.setupMock(mock)

			repo := &DBRepository{db: gdb}
			user, err := repo.GetByID(context.Background(), tc.id)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepository_GetByEmail(t *testing.T) {
	// Create a fixed timestamp for testing
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)

	type fields struct {
		setupMock func(sqlmock.Sqlmock)
	}
	testCases := []struct {
		name          string
		email         string
		fields        fields
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:  "Success",
			email: "test@example.com",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "active", "created_at", "updated_at", "deleted_at"}).
						AddRow(1, "test@example.com", "Test User", "", "", false, fixedTime, fixedTime, nil)
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("test@example.com", 1).
						WillReturnRows(rows)
				},
			},
			expectedUser: &domain.User{
				ID:        1,
				Email:     "test@example.com",
				Name:      "Test User",
				Password:  "",
				Role:      "",
				Active:    false,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			expectedError: nil,
		},
		{
			name:  "User Not Found",
			email: "notfound@example.com",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("notfound@example.com", 1).
						WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "active", "created_at", "updated_at", "deleted_at"}))
				},
			},
			expectedUser:  nil,
			expectedError: ErrUserNotFound,
		},
		{
			name:  "Database Error",
			email: "error@example.com",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
						WithArgs("error@example.com", 1).
						WillReturnError(errors.New("database error"))
				},
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.NoError(t, err)

			tc.fields.setupMock(mock)

			repo := &DBRepository{db: gdb}
			user, err := repo.GetByEmail(context.Background(), tc.email)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepository_Update(t *testing.T) {
	t.Parallel()
	// Create a fixed timestamp for testing
	fixedTime := time.Date(2025, 6, 22, 9, 0, 0, 0, time.UTC)
	//userDummy := &User{
	//	ID:        999,
	//	Email:     "notfound@example.com",
	//	Name:      "Not Found User",
	//	Role:      "user",
	//	Active:    true,
	//	Password:  "dummy_password",
	//	CreatedAt: fixedTime,
	//	UpdatedAt: fixedTime,
	//}

	type fields struct {
		setupMock func(sqlmock.Sqlmock)
	}
	testCases := []struct {
		name          string
		user          *domain.User
		fields        fields
		expectedError error
	}{
		{
			name: "Success",
			user: &domain.User{
				ID:        1,
				Email:     "updated@example.com",
				Name:      "Updated User",
				Password:  "newpassword",
				Role:      "user",
				Active:    true,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "users" SET .+ WHERE .+`).
						WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				},
			},
			expectedError: nil,
		},
		//{
		//	name: "User Not Found",
		//	user: userDummy,
		//	fields: fields{
		//		setupMock: func(mock sqlmock.Sqlmock) {
		//			mock.ExpectBegin()
		//			mock.ExpectExec(`UPDATE "users" SET .+ WHERE .+`).
		//				WillReturnResult(sqlmock.NewResult(0, 0))
		//			mock.ExpectCommit()
		//		},
		//	},
		//	expectedError: ErrUserNotFound,
		//},
		{
			name: "Database Error",
			user: &domain.User{
				ID:        2,
				Email:     "error@example.com",
				Name:      "Error User",
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "users" SET .+ WHERE .+`).
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
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			gdb = gdb.Debug()
			assert.NoError(t, err)

			tc.fields.setupMock(mock)

			repo := &DBRepository{db: gdb}
			err = repo.Update(context.Background(), tc.user)

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

func TestDBRepository_Delete(t *testing.T) {
	type fields struct {
		setupMock func(sqlmock.Sqlmock)
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
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "users" SET "deleted_at"=\$1 WHERE id = \$2 AND "users"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), "1"). // AnyArg = deleted_at timestamp
						WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				},
			},
			expectedError: nil,
		},
		{
			name: "User Not Found",
			id:   "2",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "users" SET "deleted_at"=\$1 WHERE id = \$2 AND "users"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), "2").
						WillReturnResult(sqlmock.NewResult(0, 0)) // penting: RowsAffected = 0
					mock.ExpectCommit()
				},
			},
			expectedError: ErrUserNotFound,
		},
		{
			name: "Database Error",
			id:   "3",
			fields: fields{
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "users" SET "deleted_at"=\$1 WHERE id = \$2 AND "users"\."deleted_at" IS NULL`).
						WithArgs(sqlmock.AnyArg(), "3").
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
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.NoError(t, err)

			tc.fields.setupMock(mock)

			repo := &DBRepository{db: gdb}
			err = repo.Delete(context.Background(), tc.id)

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

func TestNewDbRepository(t *testing.T) {
	db := &gorm.DB{}
	repo := NewDbRepository(db)
	_, ok := repo.(*DBRepository)
	assert.True(t, ok, "NewDbRepository should return *DBRepository")
}
