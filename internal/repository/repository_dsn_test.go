package repository

import (
    "strings"
    "testing"

    "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func TestGetDSN(t *testing.T) {
    dsn := getDSN(models.ConfigPostgres{
        Host: "localhost",
        Port: "5432",
        User: "postgres",
        Pass: "pass",
        DBName: "db",
        SSLMode: "disable",
    })
    if !strings.Contains(dsn, "postgresql://") { t.Fatalf("unexpected scheme: %s", dsn) }
    if !strings.Contains(dsn, "localhost:5432") { t.Fatalf("unexpected host: %s", dsn) }
    if !strings.Contains(dsn, "sslmode=disable") { t.Fatalf("missing sslmode: %s", dsn) }
}


