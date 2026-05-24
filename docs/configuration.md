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
| `DATABASE_URL` | _(empty)_ | Full PostgreSQL URL (Railway/Heroku). **Overrides all `PG_*` host settings** when set |
| `PG_HOST` | `localhost` | Database host (ignored if `DATABASE_URL` is set) |
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

## Text generation models (`POST /v1/generate/text`)

| `model` slug | Provider | ENV |
|---|---|---|
| `yandexgpt` (default) | YandexGPT | `YANDEX_GPT_*` |
| `gemini-flash` | Google Gemini | `GEMINI_*` |
| `openai` | OpenAI | `OPENAI_*` |

Token cost per request (in-app wallet): `yandexgpt` = 1, `gemini-flash` = 1, `openai` = 2.

## YandexGPT

| Variable | Default | Description |
|---|---|---|
| `YANDEX_GPT_API_KEY` | _(empty)_ | API key from [Yandex Cloud](https://yandex.cloud/) |
| `YANDEX_GPT_FOLDER_ID` | _(empty)_ | Folder ID for Foundation Models |
| `YANDEX_GPT_MODEL` | `yandexgpt/latest` | Model path segment |
| `YANDEX_GPT_MODEL_URI` | _(empty)_ | Full URI override, e.g. `gpt://folder/yandexgpt/latest` |
| `YANDEX_GPT_BASE_URL` | _(empty)_ | API base override |

## Gemini

| Variable | Default | Description |
|---|---|---|
| `GEMINI_API_KEY` | _(empty)_ | API key from [Google AI Studio](https://aistudio.google.com/apikey) |
| `GEMINI_MODEL` | `gemini-2.0-flash` | Model id for text (`gemini-2.0-flash`, etc.) |
| `GEMINI_IMAGE_MODEL` | `gemini-2.5-flash-image` | Model id for **Nano Banana** image generation |
| `GEMINI_API_BASE_URL` | _(empty)_ | Override API base (optional) |

Text generation uses `POST /v1/generate/text` with `model=gemini-flash`.  
Image generation (**Nano Banana**) uses `POST /v1/generate/image` with `model=nano-banana` (same `GEMINI_API_KEY`).

## OpenAI (other categories, optional)

| Variable | Default | Description |
|---|---|---|
| `OPENAI_API_KEY` | _(empty)_ | OpenAI API key |
| `OPENAI_API_BASE_URL` | _(empty)_ | Custom base URL |
| `OPENAI_MODEL` | `gpt-4o-mini` | Model name |

## Telegram Bot

| Variable | Default | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | _(empty)_ | Bot token from @BotFather. Leave empty to disable the bot. |
| `TELEGRAM_WEBHOOK_URL` | _(empty)_ | Public HTTPS URL for webhook, e.g. `https://your-app.up.railway.app/v1/telegram/webhook`. When set, polling is disabled and the URL is registered on startup. |
| `TELEGRAM_WEBHOOK_SECRET` | _(empty)_ | Secret token sent in `X-Telegram-Bot-Api-Secret-Token` (recommended on production). |

## S3 (optional, for file/avatar uploads)

| Variable | Default | Description |
|---|---|---|
| `S3_REGION` | _(empty)_ | AWS / S3-compatible region |
| `S3_ENDPOINT` | _(empty)_ | Custom endpoint URL (e.g. Yandex Cloud, MinIO) |
| `S3_ACCESS_KEY_ID` | _(empty)_ | Access key |
| `S3_SECRET_ACCESS_KEY` | _(empty)_ | Secret key |
| `S3_BUCKET` | _(empty)_ | Bucket for generated images (optional) |
| `S3_PUBLIC_BASE` | `https://storage.yandexcloud.net` | Public base URL for generated file links |

If `S3_BUCKET` and `S3_ACCESS_KEY_ID` are set, generated images are uploaded to S3. Otherwise the API returns a `data:image/...;base64,...` URL (handy in development).
