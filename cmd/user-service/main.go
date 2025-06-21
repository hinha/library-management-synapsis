package main

import (
	"context"
	"fmt"
	"github.com/hinha/library-management-synapsis/pkg/logger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hinha/library-management-synapsis/cmd/config"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	grpcHandler "github.com/hinha/library-management-synapsis/internal/delivery/grpc"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
	"github.com/hinha/library-management-synapsis/internal/seeder"
)

func main() {
	// Load user service configuration
	cfg := config.LoadUserServiceConfig()

	// Database connection
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DbHost,
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbName,
		cfg.DbPort,
	)

	gormLogger, grpcInterceptor, httpMiddleware := logger.NewLogger()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	// Initialize repositories
	userRepo := user.NewUserRepository(db)

	// Initialize and run seeder
	userSeeder := seeder.NewUserSeeder(userRepo)
	if err := userSeeder.Seed(context.Background()); err != nil {
		log.Error().Err(err).Msg("Failed to seed database")
	}

	// Initialize services
	jwtConfig := user.JWTConfig{
		SecretKey:     cfg.JwtSecret,
		TokenDuration: time.Hour * 24, // 1 day
	}
	userService := user.NewService(userRepo, jwtConfig)

	// Initialize gRPC handlers
	userHandler := grpcHandler.NewUserHandler(userService)

	// Start gRPC server
	grpcReady := make(chan struct{})
	go startGRPCServer(cfg.GrpcAddr, userHandler, grpcInterceptor, grpcReady)

	// Wait for gRPC server to be ready
	<-grpcReady

	// Start HTTP gateway
	go startHTTPServer(cfg.HttpAddr, cfg.GrpcAddr, httpMiddleware)

	// Wait for termination signal
	waitForTermination()
}

func startGRPCServer(addr string, userHandler *grpcHandler.UserHandler, logUnary grpc.UnaryServerInterceptor, ready chan struct{}) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to listen on %s", addr)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(logUnary))
	pb.RegisterUserServiceServer(s, userHandler)

	log.Info().Msgf("User service gRPC server listening at %v", lis.Addr())

	// Signal that the server is ready
	close(ready)

	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

func startHTTPServer(httpAddr, grpcAddr string, httpMiddleware func(http.Handler) http.Handler) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Add a retry mechanism for connecting to the gRPC server
	var err error
	for i := 0; i < 5; i++ {
		err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
		if err == nil {
			break
		}
		log.Warn().Err(err).Msgf("Failed to register gateway, retrying in 1 second (attempt %d/5)", i+1)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register gateway after multiple attempts")
	}

	// Apply HTTP middleware for logging
	handler := httpMiddleware(mux)

	log.Info().Msgf("User service HTTP server listening at %v", httpAddr)
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve HTTP")
	}
}

func waitForTermination() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Info().Msg("Shutting down user service...")
}
