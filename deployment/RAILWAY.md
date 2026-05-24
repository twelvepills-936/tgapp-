# Railway (backend)

## PostgreSQL

1. Add **PostgreSQL** plugin to the project (or link service).
2. On the **backend** service set variable **`DATABASE_URL`** (Railway often injects it automatically when you link Postgres).

The app reads `DATABASE_URL` and ignores `PG_HOST` when it is set.

## Migrations

SQL migrations live in `internal/migrations/`. They create `profiles`, `wallets`, `prompt_history`, etc.

**On every backend start** the service runs pending migrations automatically (see `internal/migrate`).

After the first successful deploy you should see in logs:

```text
INFO applied migration version=20251103000100 description=cybermate_core
...
```

### Run migrations manually (optional)

From repo root, with `DATABASE_URL` pointing to Railway Postgres:

```bash
go run ./cmd/migrate
```

Or Docker Flyway (local):

```bash
make db.migrate
```

## Required variables (backend)

| Variable | Notes |
|----------|--------|
| `DATABASE_URL` | From Railway Postgres |
| `ENVIRONMENT` | `production` |
| `CORS_ALLOWED_ORIGINS` | `*` (рекомендуется для Telegram Mini App) или точный URL фронта, напр. `https://tgappfront-production.up.railway.app` |

Если в Mini App при генерации видите **Load failed** — чаще всего CORS (нет заголовков) или неверный `VITE_API_BASE_URL` на фронте. Поставьте `CORS_ALLOWED_ORIGINS=*`, redeploy бэкенд.

Долгие ответы DeepSeek (>60 с) иногда обрывает прокси — попробуйте **yandexgpt** или короче промпт.
| AI keys | See `docs/configuration.md` |
| `TELEGRAM_BOT_TOKEN` | Optional |
| `TELEGRAM_WEBHOOK_URL` | `https://<backend>/v1/telegram/webhook` |

`PORT` is set by Railway — **do not** set `APP_HTTP_PORT=8090` (иначе снаружи порт не совпадёт и запросы «висят» → Load failed). Бэкенд слушает `$PORT` автоматически.

Проверка: `GET https://ваш-бэкенд.up.railway.app/health` → `{"status":"ok"}`.

Профиль/кошелёк/AI идут **напрямую в usecase** (без лишнего gRPC-хода) — быстрее и стабильнее.

## Error: `relation "profiles" does not exist`

Postgres is empty or migrations did not run. Fix:

1. Confirm **`DATABASE_URL`** on the backend service matches the Postgres you use.
2. Redeploy backend — check logs for `applied migration` or migration errors.
3. If the repo on Railway does not include `internal/migrations/`, set **`MIGRATIONS_DIR`** to the correct path or deploy from full repo root.
