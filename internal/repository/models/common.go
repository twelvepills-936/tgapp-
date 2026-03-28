package models

import (
    "time"
)

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

var (
)
