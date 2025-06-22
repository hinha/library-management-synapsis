package user

import (
	"context"
	"errors"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidCredentials is returned when the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUnauthorized is returned when a user is not authorized to perform an action
	ErrUnauthorized = errors.New("unauthorized")
)

// JWTConfig contains configuration for JWT token generation
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// Service defines the interface for user business logic
type Service interface {
	Register(ctx context.Context, name, email, password string, isAdmin bool) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	Health(ctx context.Context) (*pb.HealthCheckResponse, error)
}

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// DefaultService implements Service
type DefaultService struct {
	repoDb    IDbRepository
	repoCache ICacheRepository
	jwtConfig JWTConfig
}

// NewService creates a new DefaultService
func NewService(repoDb IDbRepository, repoCache ICacheRepository, jwtConfig JWTConfig) *DefaultService {
	return &DefaultService{
		repoDb:    repoDb,
		repoCache: repoCache,
		jwtConfig: jwtConfig,
	}
}

// Register registers a new user
func (s *DefaultService) Register(ctx context.Context, name, email, password string, isAdmin bool) (*domain.User, error) {
	var role domain.Role
	if isAdmin {
		role = domain.RoleAdmin
	} else {
		role = domain.RoleOperation
	}

	user, err := domain.NewUser(name, email, password, role)
	if err != nil {
		return nil, err
	}

	if err := s.repoDb.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *DefaultService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repoDb.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if !user.ComparePassword(password) {
		return "", "", ErrInvalidCredentials
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtConfig.TokenDuration)
	claims := &Claims{
		UserID: strconv.Itoa(int(user.ID)),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtConfig.SecretKey))
	if err != nil {
		return "", "", err
	}

	// Store token in a cache with expiration
	if err := s.repoCache.SaveUser(ctx, user); err != nil {
		return "", "", err
	}

	return signedToken, expiresAt.String(), nil
}

// GetUser retrieves a user by ID
func (s *DefaultService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repoDb.GetByID(ctx, id)
}

// UpdateUser updates a user's information
func (s *DefaultService) UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error) {
	user, err := s.repoDb.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		user.Name = name
	}

	if email != "" && email != user.Email {
		// Check if email is already taken
		_, err := s.repoDb.GetByEmail(ctx, email)
		if err == nil {
			return nil, ErrEmailAlreadyExists
		} else if !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		user.Email = email
	}

	user.UpdatedAt = time.Now()

	if err := s.repoDb.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *DefaultService) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrUnauthorized
	}

	userID, _ := strconv.Atoi(claims.UserID)
	userCache, err := s.repoCache.GetUser(ctx, uint(userID))
	if err != nil {
		return nil, ErrUnauthorized
	}

	user, err := s.repoDb.GetByID(ctx, strconv.Itoa(int(userCache.ID)))
	if err != nil {
		return nil, ErrUnauthorized
	}

	if !user.Active {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

func (s *DefaultService) Health(ctx context.Context) (*pb.HealthCheckResponse, error) {
	status := "HEALTHY"

	var componentStatus []*pb.ComponentStatus
	if err := s.repoDb.Ping(ctx); err != nil {
		status = "UNHEALTHY"
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "db",
			Status:  "DOWN",
			Message: err.Error(),
		})
	} else {
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "db",
			Status:  "UP",
			Message: "Database is healthy",
		})
	}

	if err := s.repoCache.Ping(ctx); err != nil {
		status = "UNHEALTHY"
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "cache",
			Status:  "DOWN",
			Message: err.Error(),
		})
	} else {
		componentStatus = append(componentStatus, &pb.ComponentStatus{
			Name:    "cache",
			Status:  "UP",
			Message: "Cache is healthy",
		})
	}

	return &pb.HealthCheckResponse{
		Status:     status,
		Components: componentStatus,
	}, nil
}
