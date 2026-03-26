package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Postgres ConfigPostgres
	App      ConfigApp
}

type ConfigApp struct {
	HTTPPort int
	GRPCPort int
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

func LoadConfig() Config {
	return Config{
		Postgres: LoadPostgresConfig(),
		App:      LoadAppConfig(),
	}
}

func LoadAppConfig() ConfigApp {
	return ConfigApp{
		HTTPPort: int(getenvInt64("APP_HTTP_PORT", 8090)),
		GRPCPort: int(getenvInt64("APP_GRPC_PORT", 8091)),
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
