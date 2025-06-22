package grpc

import (
	"context"
	"errors"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
	"github.com/hinha/library-management-synapsis/pkg/validator"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler implements the IService gRPC interface
type UserHandler struct {
	pb.UserServiceServer
	service user.IService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service user.IService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Register handles user registration
func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {

	// In the future, when UserRole is available in the generated code:
	if err := validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var isAdmin bool
	if req.Role == pb.UserRole_USER_ROLE_ADMIN {
		isAdmin = true
	}
	// Default to operation user if not specified
	u, err := h.service.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), isAdmin)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		log.Debug().Err(err).Msg("failed to login")
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return u.ToProto(), nil
}

// Login handles user authentication
func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	token, expiredAt, err := h.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		log.Debug().Err(err).Msg("failed to login")
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &pb.LoginResponse{
		Token:     token,
		ExpiredAt: expiredAt,
	}, nil
}

// Get retrieves a user by ID
func (h *UserHandler) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	u, err := h.service.GetUser(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return u.ToProto(), nil
}

// Update updates a user's information
func (h *UserHandler) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	u, err := h.service.UpdateUser(ctx, req.GetId(), req.GetName(), req.GetEmail())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return u.ToProto(), nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.service.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	var role pb.UserRole
	if domain.Role(claims.Role) == domain.RoleAdmin {
		role = pb.UserRole_USER_ROLE_ADMIN
	} else if domain.Role(claims.Role) == domain.RoleOperation {
		role = pb.UserRole_USER_ROLE_OPERATION
	} else {
		role = pb.UserRole_USER_ROLE_UNSPECIFIED
	}

	return &pb.ValidateTokenResponse{
		UserId:  claims.UserID,
		Role:    role,
		IsValid: true,
	}, nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return h.service.Health(ctx)
}
