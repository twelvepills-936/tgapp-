# Configuration Reference

All settings are read from environment variables. Copy `.env` and adjust for your environment.

## App

| Variable | Default | Description |
|---|---|---|
| `PORT` | — | HTTP port on Railway (injected automatically; takes precedence) |
| `APP_HTTP_PORT` | `8090` | HTTP port for local/docker when `PORT` is unset. **Do not set on Railway** if it differs from `PORT` — healthchecks will fail |
| `APP_GRPC_PORT` | `8091` | gRPC port |
| `ENVIRONMENT` | `development` | Runtime environment (`development` / `production`) |
| `LOG_LEVEL` | `info` | Log level (`debug` / `info` / `warn` / `error`) |
| `SKIP_AI_WALLET_CHECK` | `false` | When `true`, AI text/image generation does not check or deduct Cybercoins |

## HTTP Server Timeouts

| Variable | Default | Description |
|---|---|---|
| `SERVER_READ_TIMEOUT` | `120s` | Max time to read full request |
| `SERVER_WRITE_TIMEOUT` | `120s` | Max time to write full response (must be ≥ AI provider HTTP timeout, ~90s) |
| `SERVER_IDLE_TIMEOUT` | `60s` | Max keep-alive idle time |

If `SERVER_WRITE_TIMEOUT` is too low (e.g. `30s`), long text generation (DeepSeek, Gemini) is cut off with `context canceled` and the dev proxy may return **502 Bad Gateway**.

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
| `yandexgpt` (default) | YandexGPT (Foundation Models `/completion`) | `YANDEX_GPT_*` |
| `deepseek` | DeepSeek via Yandex AI Studio (`/v1/responses`) | `YANDEX_GPT_*` + `YANDEX_DEEPSEEK_MODEL` |
| `gemini-flash` | Google Gemini | `GEMINI_*` |
| `openai` | OpenAI | `OPENAI_*` |

Token cost per request (in-app wallet): `yandexgpt` = 1, `deepseek` = 2, `gemini-flash` = 1, `openai` = 2.

## YandexGPT

| Variable | Default | Description |
|---|---|---|
| `YANDEX_GPT_API_KEY` | _(empty)_ | API key from [Yandex Cloud](https://yandex.cloud/) |
| `YANDEX_GPT_FOLDER_ID` | _(empty)_ | Folder ID for Foundation Models |
| `YANDEX_GPT_MODEL` | `yandexgpt/latest` | Model path segment |
| `YANDEX_GPT_MODEL_URI` | _(empty)_ | Full URI override, e.g. `gpt://folder/yandexgpt/latest` |
| `YANDEX_GPT_BASE_URL` | _(empty)_ | Foundation Models API base override |
| `YANDEX_DEEPSEEK_MODEL` | `deepseek-v32/latest` | DeepSeek model id in catalog (also reads `YANDEX_CLOUD_MODEL`) |
| `YANDEX_RESPONSES_BASE_URL` | `https://ai.api.cloud.yandex.net/v1` | AI Studio Responses API base (DeepSeek) |

DeepSeek uses the same `YANDEX_GPT_API_KEY` and `YANDEX_GPT_FOLDER_ID` as YandexGPT, but a different HTTP API (`POST /responses`). Request body: `model`, `instructions`, `input`, `max_output_tokens`.

| `AI_TEXT_MAX_OUTPUT_TOKENS` | `4096` | Max completion tokens for YandexGPT and DeepSeek (256–8192). Was `2000` — long answers were cut off mid-text. |

Chat history in `messages[]` is trimmed server-side (assistant turns to ~6k chars) so follow-up prompts like «построй схему HTTPS» do not fail with `message N too long` after a long previous answer.

## Gemini

| Variable | Default | Description |
|---|---|---|
| `GEMINI_API_KEY` | _(empty)_ | API key from [Google AI Studio](https://aistudio.google.com/apikey) |
| `GEMINI_MODEL` | `gemini-2.0-flash-lite` | Model id for text (e.g. `gemini-2.0-flash-lite`, `gemini-2.0-flash-lite-001`). Do **not** use `*-latest` aliases — API returns 404 |
| `GEMINI_IMAGE_MODEL` | `gemini-2.5-flash-image` | Model id for **Nano Banana** image generation |
| `GEMINI_API_BASE_URL` | _(empty)_ | Override API base (optional) |

Text generation uses `POST /v1/generate/text` with `model=gemini-flash`.  
Image generation uses `POST /v1/generate/image`:

| `model` slug | Provider | Env |
|---|---|---|
| `nano-banana` | Gemini image | `GEMINI_API_KEY`, `GEMINI_IMAGE_MODEL` |
| `alice-ai-art` | Alice AI ART (Yandex AI Studio) | `YANDEX_GPT_API_KEY`, `YANDEX_GPT_FOLDER_ID`, `YANDEX_ALICE_AI_ART_MODEL` |

| `YANDEX_ALICE_AI_ART_MODEL` | `aliceai-image-art-3.0/latest` | Alice AI ART model id in catalog |
| `YANDEX_IMAGE_SIZE` | `1024x1024` | Output size for Alice AI ART (`POST /v1/images/generations`) |

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
