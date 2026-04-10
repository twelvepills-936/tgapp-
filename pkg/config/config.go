package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates all runtime configuration for the service.
type Config struct {
	Postgres ConfigPostgres
	App      ConfigApp
	CORS     ConfigCORS
	Server   ConfigServer
}

type ConfigApp struct {
	HTTPPort    int
	GRPCPort    int
	Environment string
	LogLevel    string
}

type ConfigServer struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type ConfigPostgres struct {
	Host        string
	Port        string
	User        string
	Pass        string
	DBName      string
	SSLMode     string
	SSLRootCert string
	Debug       bool

	PoolStatPeriod        time.Duration
	PoolMaxConns          int64
	PoolMinConns          int64
	PoolMaxConnLifeTime   time.Duration
	PoolMaxConnIdleTime   time.Duration
	PoolHealthCheckPeriod time.Duration
}

type ConfigCORS struct {
	AllowedOrigins []string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getenvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
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

func getenvStringSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		result := []string{}
		start := 0
		for i, ch := range v {
			if ch == ',' {
				if i > start {
					result = append(result, v[start:i])
				}
				start = i + 1
			}
		}
		if start < len(v) {
			result = append(result, v[start:])
		}
		return result
	}
	return def
}

// LoadConfig reads all configuration from environment (and .env if present).
func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		// .env not found — use system environment (normal in production).
	}
	return Config{
		Postgres: LoadPostgresConfig(),
		App:      LoadAppConfig(),
		CORS:     LoadCORSConfig(),
		Server:   LoadServerConfig(),
	}
}

func LoadAppConfig() ConfigApp {
	return ConfigApp{
		HTTPPort:    getenvInt("APP_HTTP_PORT", 8090),
		GRPCPort:    getenvInt("APP_GRPC_PORT", 8091),
		Environment: getenv("ENVIRONMENT", "development"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
}

func LoadServerConfig() ConfigServer {
	return ConfigServer{
		ReadTimeout:  getenvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getenvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getenvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
	}
}

func LoadPostgresConfig() ConfigPostgres {
	// Accept both PG_PASS and PG_PASSWORD; fall back to "postgres" for local dev.
	pass := getenv("PG_PASS", "")
	if pass == "" {
		pass = getenv("PG_PASSWORD", "")
	}
	if pass == "" {
		pass = "postgres"
	}

	return ConfigPostgres{
		Host:        getenv("PG_HOST", "localhost"),
		Port:        getenv("PG_PORT", "5432"),
		User:        getenv("PG_USER", "postgres"),
		Pass:        pass,
		DBName:      getenv("PG_DBNAME", "myapp_db"),
		SSLMode:     getenv("PG_SSLMODE", "disable"),
		SSLRootCert: getenv("PG_SSLROOTCERT", ""),
		Debug:       getenvBool("PG_DEBUG", false),

		PoolStatPeriod:        getenvDuration("PG_POOL_STAT_PERIOD", 30*time.Second),
		PoolMaxConns:          getenvInt64("PG_POOL_MAX_CONNS", 10),
		PoolMinConns:          getenvInt64("PG_POOL_MIN_CONNS", 2),
		PoolMaxConnLifeTime:   getenvDuration("PG_POOL_MAX_CONN_LIFETIME", time.Hour),
		PoolMaxConnIdleTime:   getenvDuration("PG_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		PoolHealthCheckPeriod: getenvDuration("PG_POOL_HEALTH_CHECK_PERIOD", time.Minute),
	}
}

func LoadCORSConfig() ConfigCORS {
	return ConfigCORS{
		AllowedOrigins: getenvStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
	}
}
