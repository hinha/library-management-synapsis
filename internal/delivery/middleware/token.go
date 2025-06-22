package middleware

import (
	"context"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type Middleware struct {
	service    user.IService
	authClient pb.UserServiceClient
}

func NewMiddleware(service user.IService, grpcClient *grpc.ClientConn) *Middleware {
	return &Middleware{
		service:    service,
		authClient: pb.NewUserServiceClient(grpcClient),
	}
}

func (m *Middleware) AuthValidateToken() grpc.UnaryServerInterceptor {
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
			"/user.UserService/Login":         true,
			"/user.UserService/Register":      true,
			"/user.UserService/ValidateToken": true,
			"/user.UserService/HealthCheck":   true,
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

func (m *Middleware) CrossValidateToken() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		whitelist := map[string]bool{
			"/book.BookService/HealthCheck":               true,
			"/transaction.TransactionService/HealthCheck": true,
		}
		if whitelist[info.FullMethod] {
			return handler(ctx, req)
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		token := strings.TrimPrefix(authHeader[0], "Bearer ")

		response, err := m.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Error().Err(err).Msg("Failed to validate token")
				return nil, status.Error(codes.Internal, "Internal server error")
			}
			if st.Code() == codes.Unavailable {
				log.Error().Err(err).Msg("Auth service unavailable")
				return nil, status.Error(codes.Unavailable, "error connecting another service")
			} else if st.Code() == codes.Unauthenticated {
				log.Error().Err(err).Msg("Invalid token")
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			} else if st.Code() == codes.PermissionDenied {
				log.Error().Err(err).Msg("Permission denied")
				return nil, status.Error(codes.PermissionDenied, "permission denied")
			} else {
				log.Error().Err(err).Msg("Unknown error during token validation")
			}

			return nil, err
		}

		log.Info().Str("path", info.FullMethod).
			Dur("duration", time.Since(start)).
			Interface("request", req).
			Interface("response", response.String()).
			Err(err).
			Msg("ValidateToken")

		return handler(ctx, req)
	}
}

func checkPermission(srcId, dstId, role string) error {
	if srcId != dstId && role != string(domain.RoleAdmin) {
		return status.Error(codes.PermissionDenied, "permission denied")
	}
	return nil
}
