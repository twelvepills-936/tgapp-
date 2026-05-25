package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates all runtime configuration for the service.
type Config struct {
	Postgres ConfigPostgres
	App      ConfigApp
	CORS     ConfigCORS
	Server   ConfigServer
	AI       ConfigAI
	Gemini   ConfigGemini
	Yandex   ConfigYandex
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
	// DatabaseURL, when set (e.g. Railway DATABASE_URL), takes precedence over PG_* fields.
	DatabaseURL string
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

type ConfigAI struct {
	APIKey  string
	BaseURL string
	Model   string
}

type ConfigGemini struct {
	APIKey     string
	Model      string
	ImageModel string
	BaseURL    string
}

type ConfigYandex struct {
	APIKey                 string
	FolderID               string
	Model                  string
	ModelURI               string
	BaseURL                string
	DeepSeekModel          string
	AliceAIArtModel        string
	ResponsesBaseURL       string
	ImageSize              string
	TextMaxOutputTokens    int
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
		AI:       LoadAIConfig(),
		Gemini:   LoadGeminiConfig(),
		Yandex:   LoadYandexConfig(),
	}
}

func LoadAIConfig() ConfigAI {
	model := getenv("OPENAI_TEXT_MODEL", "")
	if model == "" {
		model = getenv("OPENAI_MODEL", "gpt-4o-mini")
	}
	return ConfigAI{
		APIKey:  getenv("OPENAI_API_KEY", ""),
		BaseURL: getenv("OPENAI_API_BASE_URL", ""),
		Model:   model,
	}
}

func LoadYandexConfig() ConfigYandex {
	deepSeek := strings.TrimSpace(getenv("YANDEX_DEEPSEEK_MODEL", ""))
	if deepSeek == "" {
		deepSeek = strings.TrimSpace(getenv("YANDEX_CLOUD_MODEL", "deepseek-v32/latest"))
	}
	return ConfigYandex{
		APIKey:              getenv("YANDEX_GPT_API_KEY", ""),
		FolderID:            getenv("YANDEX_GPT_FOLDER_ID", ""),
		Model:               getenv("YANDEX_GPT_MODEL", "yandexgpt/latest"),
		ModelURI:            getenv("YANDEX_GPT_MODEL_URI", ""),
		BaseURL:             getenv("YANDEX_GPT_BASE_URL", ""),
		DeepSeekModel:       deepSeek,
		AliceAIArtModel:     strings.TrimSpace(getenv("YANDEX_ALICE_AI_ART_MODEL", "aliceai-image-art-3.0/latest")),
		ResponsesBaseURL:    getenv("YANDEX_RESPONSES_BASE_URL", "https://ai.api.cloud.yandex.net/v1"),
		ImageSize:           getenv("YANDEX_IMAGE_SIZE", "1024x1024"),
		TextMaxOutputTokens: clampTextMaxOutputTokens(getenvInt("AI_TEXT_MAX_OUTPUT_TOKENS", 4096)),
	}
}

func clampTextMaxOutputTokens(n int) int {
	if n < 256 {
		return 256
	}
	if n > 8192 {
		return 8192
	}
	return n
}

func LoadGeminiConfig() ConfigGemini {
	return ConfigGemini{
		APIKey:     getenv("GEMINI_API_KEY", ""),
		Model:      getenv("GEMINI_MODEL", "gemini-2.0-flash-lite"),
		ImageModel: getenv("GEMINI_IMAGE_MODEL", "gemini-2.5-flash-image"),
		BaseURL:    getenv("GEMINI_API_BASE_URL", ""),
	}
}

func LoadAppConfig() ConfigApp {
	// Railway injects PORT — it must win over APP_HTTP_PORT or healthchecks probe the wrong port.
	httpPort := getenvInt("PORT", 0)
	if httpPort == 0 {
		httpPort = getenvInt("APP_HTTP_PORT", 8090)
	}
	return ConfigApp{
		HTTPPort:    httpPort,
		GRPCPort:    getenvInt("APP_GRPC_PORT", 8091),
		Environment: getenv("ENVIRONMENT", "development"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
}

// SkipRegistrationCheck is true in development unless DEV_SKIP_REGISTRATION=false.
func SkipRegistrationCheck() bool {
	if v := os.Getenv("DEV_SKIP_REGISTRATION"); v != "" {
		return getenvBool("DEV_SKIP_REGISTRATION", true)
	}
	return getenv("ENVIRONMENT", "development") == "development"
}

// SkipAIWalletCheck disables balance checks and deductions for text/image generation.
func SkipAIWalletCheck() bool {
	return getenvBool("SKIP_AI_WALLET_CHECK", false)
}

func LoadServerConfig() ConfigServer {
	return ConfigServer{
		ReadTimeout:  getenvDuration("SERVER_READ_TIMEOUT", 120*time.Second),
		WriteTimeout: getenvDuration("SERVER_WRITE_TIMEOUT", 120*time.Second),
		IdleTimeout:  getenvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
	}
}

func LoadPostgresConfig() ConfigPostgres {
	pool := ConfigPostgres{
		PoolStatPeriod:        getenvDuration("PG_POOL_STAT_PERIOD", 30*time.Second),
		PoolMaxConns:          getenvInt64("PG_POOL_MAX_CONNS", 10),
		PoolMinConns:          getenvInt64("PG_POOL_MIN_CONNS", 2),
		PoolMaxConnLifeTime:   getenvDuration("PG_POOL_MAX_CONN_LIFETIME", time.Hour),
		PoolMaxConnIdleTime:   getenvDuration("PG_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		PoolHealthCheckPeriod: getenvDuration("PG_POOL_HEALTH_CHECK_PERIOD", time.Minute),
		Debug:                 getenvBool("PG_DEBUG", false),
		SSLRootCert:           getenv("PG_SSLROOTCERT", ""),
	}

	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		pool.DatabaseURL = dbURL
		pool.SSLMode = getenv("PG_SSLMODE", "")
		return pool
	}

	// Accept both PG_PASS and PG_PASSWORD; fall back to "postgres" for local dev.
	pass := getenv("PG_PASS", "")
	if pass == "" {
		pass = getenv("PG_PASSWORD", "")
	}
	if pass == "" {
		pass = "postgres"
	}

	pool.Host = getenv("PG_HOST", "localhost")
	pool.Port = getenv("PG_PORT", "5432")
	pool.User = getenv("PG_USER", "postgres")
	pool.Pass = pass
	pool.DBName = getenv("PG_DBNAME", "myapp_db")
	pool.SSLMode = getenv("PG_SSLMODE", "disable")
	return pool
}

func LoadCORSConfig() ConfigCORS {
	return ConfigCORS{
		AllowedOrigins: getenvStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
	}
}
