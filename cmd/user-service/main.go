package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hinha/library-management-synapsis/cmd/config"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	grpcHandler "github.com/hinha/library-management-synapsis/internal/delivery/grpc"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	userRepo := user.NewGormRepository(db)

	// Initialize services
	jwtConfig := user.JWTConfig{
		SecretKey:     cfg.JwtSecret,
		TokenDuration: time.Hour * 24, // 1 day
	}
	userService := user.NewService(userRepo, jwtConfig)

	// Initialize gRPC handlers
	userHandler := grpcHandler.NewUserHandler(userService)

	// Start gRPC server
	go startGRPCServer(cfg.GrpcAddr, userHandler)

	// Start HTTP gateway
	go startHTTPServer(cfg.HttpAddr, cfg.GrpcAddr)

	// Wait for termination signal
	waitForTermination()
}

func startGRPCServer(addr string, userHandler *grpcHandler.UserHandler) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userHandler)

	log.Printf("User service gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func startHTTPServer(httpAddr, grpcAddr string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	log.Printf("User service HTTP server listening at %v", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}

func waitForTermination() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down user service...")
}
