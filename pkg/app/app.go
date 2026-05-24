package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds all startup parameters for the application.
type Config struct {
	GRPCPort     int
	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	CORSOrigins  []string
}

// App holds the gRPC server, HTTP gateway and their lifecycle.
type App struct {
	GrpcServer   *grpc.Server
	ServeMux     *runtime.ServeMux
	GrpcConn     *grpc.ClientConn
	httpServer   *http.Server
	httpRoot     http.Handler
	grpcPort     int
	httpPort     int
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	corsOrigins    []string
	httpListenOnce sync.Once
	httpListenCh   chan struct{}
}

// SetHTTPRootHandler wraps the gRPC gateway (e.g. extra REST routes).
func (a *App) SetHTTPRootHandler(h http.Handler) {
	a.httpRoot = h
}

// LoadConfigFromEnv populates Config from environment variables.
func LoadConfigFromEnv() Config {
	httpPort := getenvInt("PORT", 0)
	if httpPort == 0 {
		httpPort = getenvInt("APP_HTTP_PORT", 8090)
	}
	return Config{
		GRPCPort:     getenvInt("APP_GRPC_PORT", 8091),
		HTTPPort:     httpPort,
		ReadTimeout:  getenvDuration("SERVER_READ_TIMEOUT", 120*time.Second),
		WriteTimeout: getenvDuration("SERVER_WRITE_TIMEOUT", 120*time.Second),
		IdleTimeout:  getenvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		CORSOrigins:  getenvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
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

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getenvSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return def
}

// New creates a new App instance but does not start listening yet.
func New(ctx context.Context, cfg Config) (*App, error) {
	grpcServer := grpc.NewServer()
	mux := runtime.NewServeMux()

	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &App{
		GrpcServer:   grpcServer,
		ServeMux:     mux,
		GrpcConn:     conn,
		grpcPort:     cfg.GRPCPort,
		httpPort:     cfg.HTTPPort,
		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,
		idleTimeout:  cfg.IdleTimeout,
		corsOrigins:  cfg.CORSOrigins,
		httpListenCh: make(chan struct{}),
	}, nil
}

// Init starts gRPC listener and sets up the HTTP gateway server.
func (a *App) Init(ctx context.Context) error {
	errChan := make(chan error, 1)
	readyChan := make(chan struct{})

	go func() {
		grpcAddr := fmt.Sprintf(":%d", a.grpcPort)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
			return
		}
		close(readyChan)
		if err := a.GrpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC serve error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-readyChan:
	case <-time.After(30 * time.Second):
		return fmt.Errorf("gRPC server failed to start in time")
	}

	root := a.httpRoot
	if root == nil {
		root = a.ServeMux
	}

	httpAddr := fmt.Sprintf(":%d", a.httpPort)
	a.httpServer = &http.Server{
		Addr:         httpAddr,
		Handler:      a.corsMiddleware(root),
		ReadTimeout:  a.readTimeout,
		WriteTimeout: a.writeTimeout,
		IdleTimeout:  a.idleTimeout,
	}

	return nil
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	if a.httpServer == nil {
		return fmt.Errorf("app not initialized: call Init first")
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.httpServer.Shutdown(shutCtx); err != nil {
			fmt.Printf("HTTP shutdown error: %v\n", err)
		}
		a.GrpcServer.GracefulStop()
		if err := a.GrpcConn.Close(); err != nil {
			fmt.Printf("gRPC conn close error: %v\n", err)
		}
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", a.httpPort))
	if err != nil {
		return fmt.Errorf("HTTP listen on 0.0.0.0:%d: %w", a.httpPort, err)
	}
	a.markHTTPListening()
	slog.Info("CyberMate backend listening",
		slog.String("addr", listener.Addr().String()),
		slog.Int("http_port", a.httpPort),
	)
	if err := a.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP serve error: %w", err)
	}
	return nil
}

func (a *App) markHTTPListening() {
	a.httpListenOnce.Do(func() { close(a.httpListenCh) })
}

// WaitHTTPListening blocks until Run has bound the HTTP port or ctx is cancelled.
func (a *App) WaitHTTPListening(ctx context.Context) error {
	select {
	case <-a.httpListenCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// corsMiddleware adds CORS headers and handles OPTIONS preflight requests.
func (a *App) corsMiddleware(next http.Handler) http.Handler {
	allowedSet := make(map[string]struct{}, len(a.corsOrigins))
	wildcard := false
	for _, o := range a.corsOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			wildcard = true
		} else {
			allowedSet[o] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := wildcard || origin == ""
		if !allowed && origin != "" {
			_, allowed = allowedSet[origin]
		}
		if !allowed && origin != "" && isTelegramMiniAppOrigin(origin) {
			allowed = true
		}

		if allowed {
			if wildcard && origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if wildcard {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if origin != "" {
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isTelegramMiniAppOrigin(origin string) bool {
	origin = strings.ToLower(strings.TrimSpace(origin))
	return strings.HasSuffix(origin, ".telegram.org")
}
