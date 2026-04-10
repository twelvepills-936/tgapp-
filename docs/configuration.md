# Configuration Reference

All settings are read from environment variables. Copy `.env` and adjust for your environment.

## App

| Variable | Default | Description |
|---|---|---|
| `APP_HTTP_PORT` | `8090` | HTTP (REST gateway) port |
| `APP_GRPC_PORT` | `8091` | gRPC port |
| `ENVIRONMENT` | `development` | Runtime environment (`development` / `production`) |
| `LOG_LEVEL` | `info` | Log level (`debug` / `info` / `warn` / `error`) |

## HTTP Server Timeouts

| Variable | Default | Description |
|---|---|---|
| `SERVER_READ_TIMEOUT` | `30s` | Max time to read full request |
| `SERVER_WRITE_TIMEOUT` | `30s` | Max time to write full response |
| `SERVER_IDLE_TIMEOUT` | `60s` | Max keep-alive idle time |

## PostgreSQL

| Variable | Default | Description |
|---|---|---|
| `PG_HOST` | `localhost` | Database host |
| `PG_PORT` | `5432` | Database port |
| `PG_USER` | `postgres` | Database user |
| `PG_PASSWORD` / `PG_PASS` | `postgres` | Database password |
| `PG_DBNAME` | `myapp_db` | Database name |
| `PG_SSLMODE` | `disable` | SSL mode (`disable` / `require` / `verify-full`) |
| `PG_SSLROOTCERT` | _(empty)_ | Path to SSL root certificate |
| `PG_DEBUG` | `false` | Log all SQL queries |

### Connection Pool

| Variable | Default | Description |
|---|---|---|
| `PG_POOL_MAX_CONNS` | `10` | Maximum open connections |
| `PG_POOL_MIN_CONNS` | `2` | Minimum idle connections |
| `PG_POOL_MAX_CONN_LIFETIME` | `1h` | Max lifetime per connection |
| `PG_POOL_MAX_CONN_IDLE_TIME` | `30m` | Max idle time per connection |
| `PG_POOL_HEALTH_CHECK_PERIOD` | `1m` | Pool health-check interval |

## CORS

| Variable | Default | Description |
|---|---|---|
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins. Use `*` for all (dev only) or explicit origins in production, e.g. `https://t.me,https://myapp.com` |

## Telegram Bot

| Variable | Default | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | _(empty)_ | Bot token from @BotFather. Leave empty to disable the bot. |

## S3 (optional, for file/avatar uploads)

| Variable | Default | Description |
|---|---|---|
| `S3_REGION` | _(empty)_ | AWS / S3-compatible region |
| `S3_ENDPOINT` | _(empty)_ | Custom endpoint URL (e.g. Yandex Cloud, MinIO) |
| `S3_ACCESS_KEY_ID` | _(empty)_ | Access key |
| `S3_SECRET_ACCESS_KEY` | _(empty)_ | Secret key |
| `S3_PUBLIC_BASE` | `https://storage.yandexcloud.net` | Public base URL for generated file links |
