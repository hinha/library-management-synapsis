package user

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestDefaultService_Register(t *testing.T) {
	type args struct {
		name     string
		email    string
		password string
		isAdmin  bool
	}
	testCases := []struct {
		name          string
		args          args
		setupMock     func(repo *mocks.IDbRepository, user *domain.User)
		expectedRole  domain.Role
		expectedError error
	}{
		{
			name: "Success as Admin",
			args: args{
				name:     "Admin User",
				email:    "admin@example.com",
				password: "password",
				isAdmin:  true,
			},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedRole:  domain.RoleAdmin,
			expectedError: nil,
		},
		{
			name: "Success as Operation",
			args: args{
				name:     "Op User",
				email:    "op@example.com",
				password: "password",
				isAdmin:  false,
			},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedRole:  domain.RoleOperation,
			expectedError: nil,
		},
		{
			name: "NewUser returns error",
			args: args{
				name:     "",
				email:    "bademail",
				password: "",
				isAdmin:  false,
			},
			setupMock:     func(repo *mocks.IDbRepository, user *domain.User) {},
			expectedRole:  domain.RoleOperation,
			expectedError: errors.New("invalid user data"),
		},
		{
			name: "Repo returns error",
			args: args{
				name:     "User",
				email:    "repoerr@example.com",
				password: "password",
				isAdmin:  false,
			},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				// Only set up the mock if user is not nil
				if user != nil {
					repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("repo error"))
				}
			},
			expectedRole:  domain.RoleOperation,
			expectedError: errors.New("repo error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			var user *domain.User
			var err error

			// Simulate NewUser error for invalid data
			if tc.name == "NewUser returns error" {
				user, err = nil, errors.New("invalid user data")
			} else {
				user, err = domain.NewUser(tc.args.name, tc.args.email, tc.args.password, tc.expectedRole)
				assert.NoError(t, err)
			}

			tc.setupMock(repo, user)

			svc := &DefaultService{repoDb: repo}
			gotUser, gotErr := svc.Register(context.Background(), tc.args.name, tc.args.email, tc.args.password, tc.args.isAdmin)

			if tc.expectedError != nil {
				assert.Error(t, gotErr)
				assert.Equal(t, tc.expectedError.Error(), gotErr.Error())
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotUser)
				assert.Equal(t, tc.args.name, gotUser.Name)
				assert.Equal(t, tc.args.email, gotUser.Email)
				assert.Equal(t, tc.expectedRole, gotUser.Role)
			}
			// Only assert expectations if we expect repo.Create to be called
			if tc.name != "NewUser returns error" {
				repo.AssertExpectations(t)
			}
		})
	}
}

func TestDefaultService_Login(t *testing.T) {
	type args struct {
		email    string
		password string
	}
	testCases := []struct {
		name          string
		args          args
		setupMock     func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User)
		expectedError error
	}{
		{
			name: "Success",
			args: args{
				email:    "user@example.com",
				password: "password",
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				cache.On("SaveUser", mock.Anything, user).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			args: args{
				email:    "notfound@example.com",
				password: "password",
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				repo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, ErrUserNotFound)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Wrong password",
			args: args{
				email:    "user@example.com",
				password: "wrongpassword",
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Repo returns error",
			args: args{
				email:    "user@example.com",
				password: "password",
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Cache returns error",
			args: args{
				email:    "user@example.com",
				password: "password",
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				cache.On("SaveUser", mock.Anything, user).Return(errors.New("cache error"))
			},
			expectedError: errors.New("cache error"),
		},
	}

	// Prepare a valid user for positive cases
	validUser, _ := domain.NewUser("Test User", "user@example.com", "password", domain.RoleOperation)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			cache := new(mocks.ICacheRepository)
			var user *domain.User
			if tc.name == "Wrong password" {
				// User with a different password
				user, _ = domain.NewUser("Test User", "user@example.com", "password", domain.RoleOperation)
			} else if tc.name == "Success" || tc.name == "Cache returns error" {
				user = validUser
			} else {
				user = nil
			}

			tc.setupMock(repo, cache, user)

			svc := &DefaultService{
				repoDb:    repo,
				repoCache: cache,
				jwtConfig: JWTConfig{SecretKey: "secret", TokenDuration: 60},
			}

			token, expiresAt, err := svc.Login(context.Background(), tc.args.email, tc.args.password)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Empty(t, token)
				assert.Empty(t, expiresAt)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotEmpty(t, expiresAt)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestDefaultService_GetUser(t *testing.T) {
	type args struct {
		id string
	}
	testCases := []struct {
		name          string
		args          args
		setupMock     func(repo *mocks.IDbRepository, user *domain.User)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name: "Success",
			args: args{id: "1"},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
			},
			expectedUser: &domain.User{
				ID:    1,
				Name:  "Test User",
				Email: "user@example.com",
				Role:  domain.RoleOperation,
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			args: args{id: "2"},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "2").Return(nil, errors.New("not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			var user *domain.User
			if tc.expectedUser != nil {
				user = &domain.User{
					ID:    tc.expectedUser.ID,
					Name:  tc.expectedUser.Name,
					Email: tc.expectedUser.Email,
					Role:  tc.expectedUser.Role,
				}
			}
			tc.setupMock(repo, user)

			svc := &DefaultService{repoDb: repo}

			gotUser, gotErr := svc.GetUser(context.Background(), tc.args.id)

			if tc.expectedError != nil {
				assert.Error(t, gotErr)
				assert.Equal(t, tc.expectedError.Error(), gotErr.Error())
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotUser)
				assert.Equal(t, tc.expectedUser.ID, gotUser.ID)
				assert.Equal(t, tc.expectedUser.Name, gotUser.Name)
				assert.Equal(t, tc.expectedUser.Email, gotUser.Email)
				assert.Equal(t, tc.expectedUser.Role, gotUser.Role)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_UpdateUser(t *testing.T) {
	type args struct {
		id    string
		name  string
		email string
	}
	testCases := []struct {
		name          string
		args          args
		setupMock     func(repo *mocks.IDbRepository, user *domain.User)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name: "Success update name",
			args: args{id: "1", name: "Updated Name", email: ""},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedUser: &domain.User{
				ID:    1,
				Name:  "Updated Name",
				Email: "user@example.com",
				Role:  domain.RoleOperation,
			},
			expectedError: nil,
		},
		{
			name: "Success update email",
			args: args{id: "1", name: "", email: "new@example.com"},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
				repo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, ErrUserNotFound)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedUser: &domain.User{
				ID:    1,
				Name:  "Test User",
				Email: "new@example.com",
				Role:  domain.RoleOperation,
			},
			expectedError: nil,
		},
		{
			name: "Email already exists",
			args: args{id: "1", name: "", email: "exists@example.com"},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
				repo.On("GetByEmail", mock.Anything, "exists@example.com").Return(&domain.User{}, nil)
			},
			expectedUser:  nil,
			expectedError: ErrEmailAlreadyExists,
		},
		{
			name: "GetByID returns error",
			args: args{id: "2", name: "Any", email: ""},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "2").Return(nil, errors.New("not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("not found"),
		},
		{
			name: "Update returns error",
			args: args{id: "1", name: "Updated Name", email: ""},
			setupMock: func(repo *mocks.IDbRepository, user *domain.User) {
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("update error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("update error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			// Use a fresh user for each test
			var user *domain.User
			if tc.args.id == "1" {
				user = &domain.User{
					ID:    1,
					Name:  "Test User",
					Email: "user@example.com",
					Role:  domain.RoleOperation,
				}
			}
			tc.setupMock(repo, user)

			svc := &DefaultService{repoDb: repo}

			gotUser, gotErr := svc.UpdateUser(context.Background(), tc.args.id, tc.args.name, tc.args.email)

			if tc.expectedError != nil {
				assert.Error(t, gotErr)
				assert.Equal(t, tc.expectedError.Error(), gotErr.Error())
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotUser)
				assert.Equal(t, tc.expectedUser.ID, gotUser.ID)
				assert.Equal(t, tc.expectedUser.Name, gotUser.Name)
				assert.Equal(t, tc.expectedUser.Email, gotUser.Email)
				assert.Equal(t, tc.expectedUser.Role, gotUser.Role)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestDefaultService_ValidateToken(t *testing.T) {
	secret := "testsecret"
	jwtConfig := JWTConfig{SecretKey: secret, TokenDuration: 60}
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	// Helper to create a valid JWT token string
	createToken := func(userID, email, role string, secret string, exp time.Time) string {
		claims := &Claims{
			UserID: userID,
			Email:  email,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(exp),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, _ := token.SignedString([]byte(secret))
		return signedToken
	}

	type args struct {
		token string
	}
	testCases := []struct {
		name          string
		args          args
		setupMock     func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User)
		expectedError error
	}{
		{
			name: "Success",
			args: args{
				token: createToken("1", "user@example.com", string(domain.RoleOperation), secret, expiresAt),
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				cache.On("GetUser", mock.Anything, uint(1)).Return(user, nil)
				repo.On("GetByID", mock.Anything, "1").Return(user, nil)
			},
			expectedError: nil,
		},
		{
			name: "Token expired",
			args: args{
				token: createToken("1", "user@example.com", string(domain.RoleOperation), secret, now.Add(-1*time.Hour)),
			},
			setupMock:     func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {},
			expectedError: errors.New("token is expired"),
		},
		{
			name: "Invalid token signature",
			args: args{
				token: createToken("1", "user@example.com", string(domain.RoleOperation), "wrongsecret", expiresAt),
			},
			setupMock:     func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {},
			expectedError: errors.New("signature is invalid"),
		},
		{
			name: "Cache returns error",
			args: args{
				token: createToken("1", "user@example.com", string(domain.RoleOperation), secret, expiresAt),
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				cache.On("GetUser", mock.Anything, uint(1)).Return(nil, errors.New("cache error"))
			},
			expectedError: ErrUnauthorized,
		},
		{
			name: "Repo returns error",
			args: args{
				token: createToken("1", "user@example.com", string(domain.RoleOperation), secret, expiresAt),
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				cache.On("GetUser", mock.Anything, uint(1)).Return(user, nil)
				repo.On("GetByID", mock.Anything, "1").Return(nil, errors.New("db error"))
			},
			expectedError: ErrUnauthorized,
		},
		{
			name: "User inactive",
			args: args{
				token: createToken("2", "inactive@example.com", string(domain.RoleOperation), secret, expiresAt),
			},
			setupMock: func(repo *mocks.IDbRepository, cache *mocks.ICacheRepository, user *domain.User) {
				cache.On("GetUser", mock.Anything, uint(2)).Return(user, nil)
				repo.On("GetByID", mock.Anything, "2").Return(user, nil)
			},
			expectedError: ErrUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			cache := new(mocks.ICacheRepository)
			var user *domain.User
			if tc.name == "User inactive" {
				user = &domain.User{
					ID:     2,
					Name:   "Inactive User",
					Email:  "inactive@example.com",
					Role:   domain.RoleOperation,
					Active: false,
				}
			} else {
				user = &domain.User{
					ID:     1,
					Name:   "Test User",
					Email:  "user@example.com",
					Role:   domain.RoleOperation,
					Active: true,
				}
			}
			tc.setupMock(repo, cache, user)

			svc := &DefaultService{
				repoDb:    repo,
				repoCache: cache,
				jwtConfig: jwtConfig,
			}

			claims, err := svc.ValidateToken(context.Background(), tc.args.token)

			if tc.expectedError != nil {
				assert.Error(t, err)
				// For token parse errors, check substring
				if tc.name == "Token expired" || tc.name == "Invalid token signature" {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				} else {
					assert.Equal(t, tc.expectedError, err)
				}
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.Email, claims.Email)
				assert.Equal(t, user.UserIDString(), claims.UserID)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestDefaultService_Health(t *testing.T) {
	type mockReturn struct {
		dbErr    error
		cacheErr error
	}
	testCases := []struct {
		name           string
		mockReturn     mockReturn
		expectedStatus string
		expectedComps  map[string]string // component name -> status
	}{
		{
			name:           "All healthy",
			mockReturn:     mockReturn{dbErr: nil, cacheErr: nil},
			expectedStatus: "HEALTHY",
			expectedComps:  map[string]string{"db": "UP", "cache": "UP"},
		},
		{
			name:           "DB unhealthy",
			mockReturn:     mockReturn{dbErr: errors.New("db down"), cacheErr: nil},
			expectedStatus: "UNHEALTHY",
			expectedComps:  map[string]string{"db": "DOWN", "cache": "UP"},
		},
		{
			name:           "Cache unhealthy",
			mockReturn:     mockReturn{dbErr: nil, cacheErr: errors.New("cache down")},
			expectedStatus: "UNHEALTHY",
			expectedComps:  map[string]string{"db": "UP", "cache": "DOWN"},
		},
		{
			name:           "Both unhealthy",
			mockReturn:     mockReturn{dbErr: errors.New("db down"), cacheErr: errors.New("cache down")},
			expectedStatus: "UNHEALTHY",
			expectedComps:  map[string]string{"db": "DOWN", "cache": "DOWN"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mocks.IDbRepository)
			cache := new(mocks.ICacheRepository)

			repo.On("Ping", mock.Anything).Return(tc.mockReturn.dbErr)
			cache.On("Ping", mock.Anything).Return(tc.mockReturn.cacheErr)

			svc := &DefaultService{repoDb: repo, repoCache: cache}

			resp, err := svc.Health(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.Status)
			compStatus := map[string]string{}
			for _, comp := range resp.Components {
				compStatus[comp.Name] = comp.Status
			}
			assert.Equal(t, tc.expectedComps, compStatus)

			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}
