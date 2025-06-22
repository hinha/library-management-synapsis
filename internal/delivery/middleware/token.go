package middleware

import (
	"context"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type Middleware struct {
	service user.IService
}

func NewUserMiddleware(service user.IService) *Middleware {
	return &Middleware{
		service: service,
	}
}

func (m *Middleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		whitelist := map[string]bool{
			"/user.IService/Login":       true,
			"/user.IService/Register":    true,
			"/user.IService/HealthCheck": true,
		}

		if whitelist[info.FullMethod] {
			return handler(ctx, req)
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := m.service.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Check if user is requesting their own data or is an admin
		switch request := req.(type) {
		case *pb.GetUserRequest:
			if err := checkPermission(claims.UserID, request.GetId(), claims.Role); err != nil {
				return nil, err
			}
		case *pb.UpdateUserRequest:
			if err := checkPermission(claims.UserID, request.GetId(), claims.Role); err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}

func checkPermission(srcId, dstId, role string) error {
	if srcId != dstId && role != string(domain.RoleAdmin) {
		return status.Error(codes.PermissionDenied, "permission denied")
	}
	return nil
}
