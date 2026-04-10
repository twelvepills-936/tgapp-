# CyberMate — Telegram Mini App Backend

Go micro-service that powers the **CyberMate** Telegram Mini App.  
Exposes a REST (gRPC-Gateway) API consumed by the Mini App frontend and a Telegram bot.

## Stack

- **Go 1.24** — language
- **gRPC + grpc-gateway** — gRPC server with auto-generated REST/JSON gateway
- **PostgreSQL 15** — primary database
- **Flyway** — SQL migrations (one-shot task via Docker Compose)
- **OpenTelemetry + otelpgx** — tracing & DB metrics
- **S3-compatible storage** — avatar / file uploads (Yandex Cloud, MinIO, AWS)
- **Telegram Bot API** — bot integration (optional)

## Quick Start

### Requirements

- Go 1.24+
- Docker + Docker Compose

### 1. Configure environment

```bash
cp .env .env.local   # edit as needed
```

Key variables to change for local dev (defaults already match Docker Compose):

```
PG_PASSWORD=postgres
TELEGRAM_BOT_TOKEN=   # get from @BotFather, leave empty to disable
CORS_ALLOWED_ORIGINS=*
```

### 2. Start Postgres and run migrations

```bash
docker compose up -d postgres
docker compose run --rm migrate
# or via Make:
make db.up
```

The `migrate` container is a one-shot task: it applies all SQL files from `internal/migrations/` and then exits with code `0`.

### 3. Run the service

```bash
go run cmd/service/main.go
```

Service endpoints:

| | URL |
|---|---|
| REST API | http://localhost:8090 |
| gRPC | localhost:8091 |
| Swagger UI | http://localhost:8090/swagger |

## Project Structure

```
cmd/service/          Entry point
internal/
  domain.go           Repository + UseCase interfaces (Clean Architecture)
  bot/                Telegram bot (polling, /start handler)
  migrations/         SQL migrations (applied by Flyway)
  repository/         PostgreSQL data access
  service/            gRPC handlers (transport layer)
  usecase/            Business logic
    models/           Input/output/error models
api/                  Proto definition + Swagger JSON
pkg/
  api/                Generated gRPC/gateway code
  app/                App server lifecycle (gRPC + HTTP gateway + CORS)
  config/             Environment-based configuration
  logger/             slog helper attributes
  s3/                 S3 upload helpers
errorcodes/           gRPC error code constants
docs/                 Configuration reference
```

## API

### CyberMate Endpoints

| Method | URL | Description |
|---|---|---|
| `POST` | `/v1/register` | Register user via Telegram WebApp `initData` |
| `GET` | `/v1/users/telegram/{telegram_id}` | Get profile by Telegram ID |

All requests/responses are JSON. See `api/service.swagger.json` or the Swagger UI for full schema.

### Register request example

```json
POST /v1/register
{
  "init_data_raw": "<base64-encoded Telegram initData>",
  "start_param": "<referrer telegram_id, optional>"
}
```

## Migrations

SQL files live in `internal/migrations/` named with Flyway convention:

```
V20251103000100__cybermate_core.sql   — profiles, wallets, referrals, transactions
V20251103000200__admin_resources.sql — admins, channels, projects, proposals
```

Apply manually (if not using Docker Compose):

```bash
flyway -url=jdbc:postgresql://localhost:5433/myapp_db \
       -user=postgres -password=postgres \
       -locations=filesystem:internal/migrations migrate
```

## Testing

```bash
go test ./...            # all tests
go test ./... -cover     # with coverage
make test.cover
```

## Linting & Formatting

```bash
make fmt           # gofmt -w .
make lint          # golangci-lint via Docker
```

## Proto Generation

To regenerate `pkg/api/` after editing `api/service.proto`:

```bash
make proto.gen
```

Requires Docker (uses `bufbuild/buf`). Install proto plugins once:

```bash
go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

## Authors
The Authors and Improvements: @twelvepills-936
The Authors and Improvements: @chopic82region


## Configuration

Full environment variable reference: [docs/configuration.md](docs/configuration.md)
