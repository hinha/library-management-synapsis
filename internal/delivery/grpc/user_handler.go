package grpc

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/pkg/validator"
	"github.com/rs/zerolog/log"
	"strings"

	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	domain "github.com/hinha/library-management-synapsis/internal/domain/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserHandler implements the UserService gRPC interface
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service domain.Service
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service domain.Service) *UserHandler {
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
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
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
		if errors.Is(err, domain.ErrInvalidCredentials) {
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
	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	claims, err := h.service.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Check if user is requesting their own data or is an admin
	if claims.UserID != req.GetId() && claims.Role != string(domain.RoleAdmin) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	u, err := h.service.GetUser(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return u.ToProto(), nil
}

// Update updates a user's information
func (h *UserHandler) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	claims, err := h.service.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Check if user is updating their own data or is an admin
	if claims.UserID != req.GetId() && claims.Role != string(domain.RoleAdmin) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	u, err := h.service.UpdateUser(ctx, req.GetId(), req.GetName(), req.GetEmail())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return u.ToProto(), nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return h.service.Health(ctx)
}
