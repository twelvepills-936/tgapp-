package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres ConfigPostgres
	App      ConfigApp
	JWT      ConfigJWT
	Redis    ConfigRedis
	CORS     ConfigCORS
	Server   ConfigServer
}

type ConfigApp struct {
	HTTPPort       int
	GRPCPort       int
	Environment    string
	Debug          bool
	LogLevel       string
	SwaggerEnabled bool
	APITitle       string
	APIVersion     string
}

type ConfigServer struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type ConfigPostgres struct {
	Host           string
	Port           string
	User           string
	Pass           string
	DBName         string
	SSLMode        string
	SSLRootCert    string
	Debug          bool
	DriverLogLevel string

	PoolStatPeriod        time.Duration
	PoolMaxConns          int64
	PoolMinConns          int64
	PoolMaxConnLifeTime   time.Duration
	PoolMaxConnIdleTime   time.Duration
	PoolHealthCheckPeriod time.Duration
}

type ConfigJWT struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type ConfigRedis struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ConfigCORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}

func getenvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func getenvStringSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		// Простой split по запятой
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

func LoadConfig() Config {
	// Загружаем .env файл, если он существует
	if err := godotenv.Load(); err != nil {
		// .env файл не найден, используем системные переменные окружения
		// В продакшене это нормально
	}

	return Config{
		Postgres: LoadPostgresConfig(),
		App:      LoadAppConfig(),
		JWT:      LoadJWTConfig(),
		Redis:    LoadRedisConfig(),
		CORS:     LoadCORSConfig(),
		Server:   LoadServerConfig(),
	}
}

func LoadAppConfig() ConfigApp {
	return ConfigApp{
		HTTPPort:       getenvInt("APP_HTTP_PORT", 8090),
		GRPCPort:       getenvInt("APP_GRPC_PORT", 8091),
		Environment:    getenv("ENVIRONMENT", "development"),
		Debug:          getenvBool("DEBUG", false),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		SwaggerEnabled: getenvBool("SWAGGER_ENABLED", true),
		APITitle:       getenv("API_TITLE", "Your API"),
		APIVersion:     getenv("API_VERSION", "1.0.0"),
	}
}

func LoadServerConfig() ConfigServer {
	return ConfigServer{
		Host:         getenv("SERVER_HOST", "localhost"),
		ReadTimeout:  getenvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getenvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getenvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
	}
}

func LoadPostgresConfig() ConfigPostgres {
	// Поддержка PG_PASS и PG_PASSWORD (как указано в README)
	// Значение по умолчанию "postgres" соответствует настройкам Docker-контейнера
	pass := getenv("PG_PASS", "")
	if pass == "" {
		pass = getenv("PG_PASSWORD", "")
	}
	// Если ни одна переменная не установлена, используем значение по умолчанию
	if pass == "" {
		pass = "postgres"
	}

	return ConfigPostgres{
		Host:           getenv("PG_HOST", "localhost"),
		Port:           getenv("PG_PORT", "5432"),
		User:           getenv("PG_USER", "postgres"),
		Pass:           pass,
		DBName:         getenv("PG_DBNAME", "postgres"),
		SSLMode:        getenv("PG_SSLMODE", "disable"),
		SSLRootCert:    getenv("PG_SSLROOTCERT", ""),
		Debug:          getenvBool("PG_DEBUG", false),
		DriverLogLevel: getenv("PG_DRIVER_LOG_LEVEL", "info"),

		PoolStatPeriod:        getenvDuration("PG_POOL_STAT_PERIOD", 30*time.Second),
		PoolMaxConns:          getenvInt64("PG_POOL_MAX_CONNS", 10),
		PoolMinConns:          getenvInt64("PG_POOL_MIN_CONNS", 1),
		PoolMaxConnLifeTime:   getenvDuration("PG_POOL_MAX_CONN_LIFETIME", time.Hour),
		PoolMaxConnIdleTime:   getenvDuration("PG_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		PoolHealthCheckPeriod: getenvDuration("PG_POOL_HEALTH_CHECK_PERIOD", time.Minute),
	}
}

func LoadJWTConfig() ConfigJWT {
	return ConfigJWT{
		Secret:          getenv("JWT_SECRET", "your-super-secret-jwt-key-change-this"),
		AccessTokenTTL:  getenvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getenvDuration("JWT_REFRESH_TOKEN_TTL", 168*time.Hour), // 7 days
	}
}

func LoadRedisConfig() ConfigRedis {
	return ConfigRedis{
		Host:     getenv("REDIS_HOST", "localhost"),
		Port:     getenv("REDIS_PORT", "6379"),
		Password: getenv("REDIS_PASSWORD", ""),
		DB:       getenvInt("REDIS_DB", 0),
	}
}

func LoadCORSConfig() ConfigCORS {
	return ConfigCORS{
		AllowedOrigins: getenvStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		AllowedMethods: getenvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders: getenvStringSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
	}
}
