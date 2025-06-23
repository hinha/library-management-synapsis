package main

import (
	"context"
	"github.com/hinha/library-management-synapsis/internal/delivery/middleware"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"github.com/hinha/library-management-synapsis/internal/infrastructure/client"
	"github.com/hinha/library-management-synapsis/internal/infrastructure/persistance"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hinha/library-management-synapsis/cmd/config"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/transaction"
	grpcHandler "github.com/hinha/library-management-synapsis/internal/delivery/grpc"
	"github.com/hinha/library-management-synapsis/internal/domain/transaction"
	"github.com/hinha/library-management-synapsis/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("Warning: .env file not found")
	}

	// Load transaction service configuration
	cfg := config.LoadTransactionServiceConfig()

	// Initialize loggers
	gormLogger, grpcInterceptor, httpMiddleware := logger.NewLogger()

	db, err := persistance.NewDatabaseConnection(cfg, gormLogger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	dbClose, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database connection")
	}
	defer dbClose.Close()

	// Auto migrate the schema
	if err := db.AutoMigrate(&domain.Transaction{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to book service
	bookConn, err := client.NewGRPCClient(ctx, config.SharedGrpcBookServiceAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to book service")
	}
	defer bookConn.Close()

	// Connect to auth service
	authConn, err := client.NewGRPCClient(ctx, config.SharedGrpcAuthServiceAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to book service")
	}
	defer bookConn.Close()

	middlewareHandler := middleware.NewMiddleware(nil, authConn)
	bookClient := middleware.NewBookServiceClient(bookConn)
	bookRepo := middleware.NewBookRepositoryAdapter(bookClient)

	// Initialize repositories
	transactionRepo := transaction.NewGormRepository(db)

	// Initialize services
	transactionService := transaction.NewService(transactionRepo, bookRepo)

	// Initialize gRPC handlers
	transactionHandler := grpcHandler.NewTransactionHandler(transactionService)

	// Start gRPC server
	grpcReady := make(chan struct{})
	go startGRPCServer(cfg.GrpcAddr, transactionHandler, middlewareHandler, grpcInterceptor, grpcReady)

	// Wait for gRPC server to be ready
	<-grpcReady

	// Start HTTP gateway
	go startHTTPServer(cfg.HttpAddr, cfg.GrpcAddr, httpMiddleware)

	// Wait for termination signal
	waitForTermination()
}

func startGRPCServer(addr string, transactionHandler *grpcHandler.TransactionHandler, mw *middleware.Middleware, logUnary grpc.UnaryServerInterceptor, ready chan struct{}) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to listen on %s", addr)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(logUnary, mw.CrossValidateToken()))
	pb.RegisterTransactionServiceServer(s, transactionHandler)

	log.Info().Msgf("Transaction service gRPC server listening at %v", lis.Addr())

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

	if err := pb.RegisterTransactionServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		log.Fatal().Err(err).Msg("Failed to register gateway")
	}

	// Apply HTTP middleware for logging
	handler := httpMiddleware(mux)

	log.Info().Msgf("Transaction service HTTP server listening at %v", httpAddr)
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve HTTP")
	}
}

func waitForTermination() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Info().Msg("Shutting down transaction service...")
}
