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

## Public Networking (порт в UI, не в Variables)

На скрине **Target port = 8080** — это правильно, если в логах деплоя есть:

```text
starting HTTP server http_port=8080
CyberMate backend listening on http://0.0.0.0:8080
```

**Не** ставьте 8090 в Target port на Railway.

| Где | Что |
|-----|-----|
| Railway → Networking → Target port | **8080** (или Auto) |
| Variables `APP_HTTP_PORT` | **удалить** |
| Variables `PORT` | **не добавлять** (даёт Railway) |

Если в Mini App **Load failed**, а URL API верный — откройте в браузере:

`https://tgapp-production-469a.up.railway.app/health`

- `{"status":"ok"}` — бэкенд жив, ищите CORS / фронт.
- **502 / Application failed to respond** — процесс не отвечает на `PORT` до таймаута healthcheck. Откройте **Deploy Logs** и найдите одно из:
  - `failed to init postgres` — нет или неверный **`DATABASE_URL`** (привяжите Postgres к сервису).
  - `failed to apply database migrations` — нет папки `internal/migrations` (неверный **Root Directory**).
  - `failed to init app` / `gRPC server failed to start` — редко; перезапуск.
  - Нет строки `CyberMate backend listening on http://0.0.0.0:8080` — бинарь не стартовал (`./bin/server` не собран → проверьте **Build Logs**).
  - Сразу после старта должны быть `starting CyberMate backend` и (через несколько секунд) `API routes ready`. **`GET /health`** отвечает `{"status":"ok"}` ещё до подключения к БД.

**Root Directory** сервиса бэкенда = корень репозитория `tgback` (где `cmd/service`, `railway.toml`).

**Start Command** (если задан вручную): **`./bin/server`** или оставьте из `railway.toml`.

### Ошибка: в логах только `migrations complete`, а запросы дают 502

Строка `INFO migrations complete dir=internal/migrations` пишется **утилитой** `cmd/migrate` (`go run ./cmd/migrate`). После неё процесс **завершается** — порт `PORT` никто не слушает → 502.

| Правильно (API) | Неправильно (только миграции) |
|-----------------|-------------------------------|
| `./bin/server` | `go run ./cmd/migrate` |
| Build: `go build -o bin/server ./cmd/service` | Build: `go build -o bin/migrate ./cmd/migrate` |

В логах **рабочего** бэкенда должны быть:

```text
starting CyberMate backend http_port=8080
CyberMate backend listening on http://0.0.0.0:8080
database migrations applied (HTTP server keeps running)
API routes ready
```

Если видите только `migrations complete` — в Railway → **Settings → Deploy → Custom Start Command** сбросьте на `./bin/server` или удалите override.

Проверка: `GET https://ваш-бэкенд.up.railway.app/health` → `{"status":"ok"}`.

Профиль/кошелёк/AI идут **напрямую в usecase** (без лишнего gRPC-хода) — быстрее и стабильнее.

## Error: `relation "profiles" does not exist`

Postgres is empty or migrations did not run. Fix:

1. Confirm **`DATABASE_URL`** on the backend service matches the Postgres you use.
2. Redeploy backend — check logs for `applied migration` or migration errors.
3. If the repo on Railway does not include `internal/migrations/`, set **`MIGRATIONS_DIR`** to the correct path or deploy from full repo root.
