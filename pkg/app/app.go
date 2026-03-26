package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config represents application configuration
type Config struct {
	GRPCPort int
	HTTPPort int
}

// App represents the application with gRPC server and HTTP gateway
type App struct {
	GrpcServer *grpc.Server
	ServeMux   *runtime.ServeMux
	GrpcConn   *grpc.ClientConn
	httpServer *http.Server
	grpcPort   int
	httpPort   int
}

// LoadConfigFromEnv loads configuration from environment variables
// Replaces gitlab16.skiftrade.kz/libs-go/new-app.LoadConfigFromEnv
func LoadConfigFromEnv() Config {
	return Config{
		GRPCPort: getenvInt("APP_GRPC_PORT", 8091),
		HTTPPort: getenvInt("APP_HTTP_PORT", 8090),
	}
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// New creates a new application instance
// Replaces gitlab16.skiftrade.kz/libs-go/new-app.New
func New(ctx context.Context, cfg Config) (*App, error) {
	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create HTTP mux for gateway
	mux := runtime.NewServeMux()

	// Create gRPC client connection for gateway (will be connected in Init)
	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &App{
		GrpcServer: grpcServer,
		ServeMux:   mux,
		GrpcConn:   conn,
		grpcPort:   cfg.GRPCPort,
		httpPort:   cfg.HTTPPort,
	}, nil
}

// Init initializes the application
// Replaces app.Init
func (a *App) Init(ctx context.Context) error {
	errChan := make(chan error, 1)
	readyChan := make(chan struct{})

	// Start gRPC server in a goroutine
	go func() {
		grpcAddr := fmt.Sprintf(":%d", a.grpcPort)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
			return
		}
		close(readyChan) // Signal that listener is ready
		if err := a.GrpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC on %s: %w", grpcAddr, err)
		}
	}()

	// Wait for gRPC server to be ready or fail
	select {
	case err := <-errChan:
		return err
	case <-readyChan:
		// Server listener is ready
	case <-time.After(5 * time.Second):
		return fmt.Errorf("gRPC server failed to start in time")
	}

	// Setup HTTP server
	httpAddr := fmt.Sprintf(":%d", a.httpPort)
	a.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: a.ServeMux,
	}

	return nil
}

// Run starts the HTTP server and blocks
// Replaces app.Run
func (a *App) Run(ctx context.Context) error {
	if a.httpServer == nil {
		return fmt.Errorf("app not initialized, call Init first")
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		if err := a.httpServer.Shutdown(context.Background()); err != nil {
			fmt.Printf("failed to shutdown HTTP server: %v\n", err)
		}
		a.GrpcServer.GracefulStop()
		if err := a.GrpcConn.Close(); err != nil {
			fmt.Printf("failed to close gRPC connection: %v\n", err)
		}
	}()

	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve HTTP: %w", err)
	}

	return nil
}

