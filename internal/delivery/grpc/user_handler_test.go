package grpc

import (
	"context"
	"errors"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"github.com/hinha/library-management-synapsis/internal/delivery/mocks"
	"github.com/hinha/library-management-synapsis/internal/domain"
	userDomain "github.com/hinha/library-management-synapsis/internal/domain"
	userEntity "github.com/hinha/library-management-synapsis/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestUserHandler_Register(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.RegisterRequest
	}
	validReq := &pb.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
		Role:     pb.UserRole_USER_ROLE_OPERATION,
	}
	validUser := &domain.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  userDomain.RoleOperation,
	}
	validAdminReq := &pb.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
		Role:     pb.UserRole_USER_ROLE_ADMIN,
	}
	tests := []struct {
		name       string
		args       args
		mockSetup  func(svc *mocks.IService)
		want       *pb.UserResponse
		wantErr    bool
		statusCode codes.Code
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("Register", mock.Anything, validReq.Name, validReq.Email, validReq.Password, false).
					Return(validUser, nil)
			},
			want:    validUser.ToProto(),
			wantErr: false,
		},
		{
			name: "success role admin",
			args: args{
				ctx: context.Background(),
				req: validAdminReq,
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("Register", mock.Anything, validAdminReq.Name, validReq.Email, validReq.Password, true).
					Return(validUser, nil)
			},
			want:    validUser.ToProto(),
			wantErr: false,
		},
		{
			name: "email already exists",
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("Register", mock.Anything, validReq.Name, validReq.Email, validReq.Password, false).
					Return(nil, userEntity.ErrEmailAlreadyExists)
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.AlreadyExists,
		},
		{
			name: "validation error",
			args: args{
				ctx: context.Background(),
				req: &pb.RegisterRequest{}, // invalid: missing fields
			},
			mockSetup: func(svc *mocks.IService) {},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "internal error",
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("Register", mock.Anything, validReq.Name, validReq.Email, validReq.Password, false).
					Return(nil, errors.New("db error"))
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.Register(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.statusCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.statusCode, st.Code())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	validReq := &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}
	validToken := "sometoken"
	validExpiredAt := "2025-01-01T00:00:00Z"

	tests := []struct {
		name       string
		req        *pb.LoginRequest
		mockSetup  func(svc *mocks.IService)
		want       *pb.LoginResponse
		wantErr    bool
		statusCode codes.Code
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func(svc *mocks.IService) {
				svc.On("Login", mock.Anything, validReq.Email, validReq.Password).
					Return(validToken, validExpiredAt, nil)
			},
			want: &pb.LoginResponse{
				Token:     validToken,
				ExpiredAt: validExpiredAt,
			},
			wantErr: false,
		},
		{
			name: "invalid credentials",
			req:  validReq,
			mockSetup: func(svc *mocks.IService) {
				svc.On("Login", mock.Anything, validReq.Email, validReq.Password).
					Return("", "", userEntity.ErrInvalidCredentials)
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Unauthenticated,
		},
		{
			name: "internal error",
			req:  validReq,
			mockSetup: func(svc *mocks.IService) {
				svc.On("Login", mock.Anything, validReq.Email, validReq.Password).
					Return("", "", errors.New("db error"))
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.Login(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.statusCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.statusCode, st.Code())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Get(t *testing.T) {
	validUser := &domain.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  userDomain.RoleOperation,
	}
	tests := []struct {
		name       string
		req        *pb.GetUserRequest
		mockSetup  func(svc *mocks.IService)
		want       *pb.UserResponse
		wantErr    bool
		statusCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.GetUserRequest{Id: "1"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("GetUser", mock.Anything, "1").Return(validUser, nil)
			},
			want:    validUser.ToProto(),
			wantErr: false,
		},
		{
			name: "user not found",
			req:  &pb.GetUserRequest{Id: "2"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("GetUser", mock.Anything, "2").Return(nil, userEntity.ErrUserNotFound)
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetUserRequest{Id: "3"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("GetUser", mock.Anything, "3").Return(nil, errors.New("db error"))
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.Get(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.statusCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.statusCode, st.Code())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	validUser := &domain.User{
		ID:    1,
		Name:  "Updated User",
		Email: "updated@example.com",
		Role:  userDomain.RoleOperation,
	}
	tests := []struct {
		name       string
		req        *pb.UpdateUserRequest
		mockSetup  func(svc *mocks.IService)
		want       *pb.UserResponse
		wantErr    bool
		statusCode codes.Code
	}{
		{
			name: "success",
			req: &pb.UpdateUserRequest{
				Id:    "1",
				Name:  "Updated User",
				Email: "updated@example.com",
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("UpdateUser", mock.Anything, "1", "Updated User", "updated@example.com").
					Return(validUser, nil)
			},
			want:    validUser.ToProto(),
			wantErr: false,
		},
		{
			name: "user not found",
			req: &pb.UpdateUserRequest{
				Id:    "2",
				Name:  "Any",
				Email: "any@example.com",
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("UpdateUser", mock.Anything, "2", "Any", "any@example.com").
					Return(nil, userEntity.ErrUserNotFound)
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.NotFound,
		},
		{
			name: "email already exists",
			req: &pb.UpdateUserRequest{
				Id:    "1",
				Name:  "Any",
				Email: "exists@example.com",
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("UpdateUser", mock.Anything, "1", "Any", "exists@example.com").
					Return(nil, userEntity.ErrEmailAlreadyExists)
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.AlreadyExists,
		},
		{
			name: "internal error",
			req: &pb.UpdateUserRequest{
				Id:    "1",
				Name:  "Any",
				Email: "any@example.com",
			},
			mockSetup: func(svc *mocks.IService) {
				svc.On("UpdateUser", mock.Anything, "1", "Any", "any@example.com").
					Return(nil, errors.New("db error"))
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.Update(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.statusCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.statusCode, st.Code())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_HealthCheck(t *testing.T) {
	mockResp := &pb.HealthCheckResponse{
		Status: "HEALTHY",
		Components: []*pb.ComponentStatus{
			{Name: "db", Status: "UP"},
			{Name: "cache", Status: "UP"},
		},
	}
	tests := []struct {
		name      string
		mockSetup func(svc *mocks.IService)
		want      *pb.HealthCheckResponse
		wantErr   bool
	}{
		{
			name: "success",
			mockSetup: func(svc *mocks.IService) {
				svc.On("Health", mock.Anything).Return(mockResp, nil)
			},
			want:    mockResp,
			wantErr: false,
		},
		{
			name: "internal error",
			mockSetup: func(svc *mocks.IService) {
				svc.On("Health", mock.Anything).Return(nil, errors.New("internal error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.HealthCheck(context.Background(), &pb.HealthCheckRequest{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestNewUserHandler(t *testing.T) {
	mockSvc := new(mocks.IService)
	handler := NewUserHandler(mockSvc)
	assert.NotNil(t, handler)
	assert.Equal(t, mockSvc, handler.service)
}

func TestUserHandler_ValidateToken(t *testing.T) {
	mockClaims := &userEntity.Claims{
		UserID: "1",
		Role:   string(domain.RoleOperation),
	}
	tests := []struct {
		name       string
		req        *pb.ValidateTokenRequest
		mockSetup  func(svc *mocks.IService)
		want       *pb.ValidateTokenResponse
		wantErr    bool
		statusCode codes.Code
	}{
		{
			name: "success operation",
			req:  &pb.ValidateTokenRequest{Token: "valid-token"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("ValidateToken", mock.Anything, "valid-token").
					Return(mockClaims, nil)
			},
			want: &pb.ValidateTokenResponse{
				UserId:  "1",
				Role:    pb.UserRole_USER_ROLE_OPERATION,
				IsValid: true,
			},
			wantErr: false,
		},
		{
			name: "success admin",
			req:  &pb.ValidateTokenRequest{Token: "admin-token"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("ValidateToken", mock.Anything, "admin-token").
					Return(&userEntity.Claims{
						UserID: "2",
						Role:   string(domain.RoleAdmin),
					}, nil)
			},
			want: &pb.ValidateTokenResponse{
				UserId:  "2",
				Role:    pb.UserRole_USER_ROLE_ADMIN,
				IsValid: true,
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			req:  &pb.ValidateTokenRequest{Token: "bad-token"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("ValidateToken", mock.Anything, "bad-token").
					Return(nil, errors.New("invalid token"))
			},
			want:       nil,
			wantErr:    true,
			statusCode: codes.Unauthenticated,
		},
		{
			name: "unspecified role",
			req:  &pb.ValidateTokenRequest{Token: "unspecified-token"},
			mockSetup: func(svc *mocks.IService) {
				svc.On("ValidateToken", mock.Anything, "unspecified-token").
					Return(&userEntity.Claims{
						UserID: "3",
						Role:   "unknown",
					}, nil)
			},
			want: &pb.ValidateTokenResponse{
				UserId:  "3",
				Role:    pb.UserRole_USER_ROLE_UNSPECIFIED,
				IsValid: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.IService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &UserHandler{
				service: mockSvc,
			}
			got, err := h.ValidateToken(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.statusCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.statusCode, st.Code())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
