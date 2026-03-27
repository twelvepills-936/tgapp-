# Telegram App Go Service 🚀

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://www.docker.com/)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%20Gateway-244c5a?logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org/)
[![Tests](https://img.shields.io/badge/Tests-8%2F8%20Passing-success)](.)
[![Linter](https://img.shields.io/badge/Linter-0%20Errors-success)](.)

Современный Go микросервис с gRPC API, HTTP Gateway и Docker Compose. Реализует Clean Architecture с полным набором best practices.

## ✨ Основные возможности

- 🎯 **Clean Architecture** - чёткое разделение слоёв (domain, usecase, repository, service)
- 🔌 **gRPC + HTTP Gateway** - одновременная поддержка gRPC и REST API
- 🐳 **Docker Compose** - простой запуск с PostgreSQL
- 📊 **OpenTelemetry** - трассировка, метрики, логирование
- 🔒 **Безопасность** - валидация входных данных, безопасная работа с транзакциями
- ✅ **Тесты** - unit-тесты для всех слоёв
- 📖 **Swagger UI** - интерактивная документация API
- 🤖 **Telegram Bot** - интеграция с Telegram Web App

## 🚀 Быстрый старт

### Требования

- Go 1.24+
- Docker и Docker Compose
- Make (опционально)

### Запуск через Docker Compose

```bash
# Клонировать репозиторий
git clone https://github.com/twelvepills-936/goservice.git
cd goservice

# Запустить сервис и PostgreSQL
docker compose up -d

# Проверить логи
docker compose logs -f service
```

Сервис будет доступен:
- 🌐 HTTP API: http://localhost:8090
- 📡 gRPC API: localhost:8091
- 📖 Swagger UI: http://localhost:8090/swagger

### Запуск локально

```bash
# Установить зависимости
go mod download

# Настроить PostgreSQL (или использовать docker compose для БД)
docker compose up -d postgres

# Запустить сервис
go run cmd/service/main.go
```

## 📁 Структура проекта

```
.
├── cmd/service/          # Точка входа приложения
├── internal/
│   ├── domain.go         # Интерфейсы слоёв (Clean Architecture)
│   ├── usecase/          # Бизнес-логика
│   ├── repository/       # Работа с БД (PostgreSQL)
│   ├── service/          # gRPC handlers
│   ├── bot/              # Telegram bot
│   └── migrations/       # SQL миграции
├── api/                  # Proto и Swagger файлы
├── pkg/
│   ├── api/              # Сгенерированные gRPC файлы
│   ├── app/              # Инициализация приложения
│   ├── config/           # Конфигурация
│   ├── logger/           # Логирование
│   └── s3/               # S3 интеграция
├── errorcodes/           # Коды ошибок API
├── Dockerfile            # Multi-stage Docker build
├── compose.yml           # Docker Compose конфигурация
```

## 🎯 API Endpoints

### Users API (template example)

```http
GET /v1/user?user_id=42
```


## 🔧 Конфигурация

Все настройки через переменные окружения:

### Приложение

| Переменная | Тип | По умолчанию | Описание |
|-----------|-----|--------------|----------|
| `APP_HTTP_PORT` | int | `8090` | Порт HTTP API |
| `APP_GRPC_PORT` | int | `8091` | Порт gRPC API |
| `APP_LOG_LEVEL` | string | `debug` | Уровень логирования |

### PostgreSQL

| Переменная | Тип | По умолчанию | Описание |
|-----------|-----|--------------|----------|
| `PG_HOST` | string | `localhost` | Хост БД |
| `PG_PORT` | string | `5432` | Порт БД |
| `PG_USER` | string | `postgres` | Пользователь |
| `PG_PASS` | string | `postgres` | Пароль |
| `PG_DBNAME` | string | `facebase` | Имя БД |
| `PG_SSLMODE` | string | `disable` | Режим SSL |
| `PG_POOL_MAX_CONNS` | int64 | `10` | Макс. соединений |

### Telegram Bot (опционально)

| Переменная | Описание |
|-----------|----------|
| `TELEGRAM_BOT_TOKEN` | Токен бота |

[Полная таблица конфигурации →](docs/configuration.md)

## 🧪 Тестирование

### Unit-тесты

```bash
# Запустить все тесты
go test ./...

# С подробным выводом
go test ./... -v

# С покрытием
go test ./... -cover
```

### API тестирование

```bash
# Импортировать коллекцию в Postman
# Файл: postman_collection.json

# Или использовать curl
curl http://localhost:8090/v1/user?user_id=1
```

## 🛠️ Разработка

### Генерация proto файлов

```bash
# Установить зависимости
go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Сгенерировать
make gen.proto
```

### Добавление нового endpoint

1. Описать метод в `api/service.proto`
2. Сгенерировать код: `make gen.proto`
3. Реализовать в слоях:
   - Repository: `internal/repository/`
   - UseCase: `internal/usecase/`
   - Service: `internal/service/`
4. Добавить тесты
5. Обновить Postman коллекцию

### Миграции БД

SQL миграции находятся в `internal/migrations/`:

```sql
-- V20251103000100__facebase_core.sql
-- Основные таблицы: profiles, wallets, referrals
```

Применяются автоматически при старте контейнера PostgreSQL.

## 📊 База данных

### Схема

```
profiles
├─ id (BIGSERIAL PRIMARY KEY)
├─ name (TEXT)
├─ telegram_id (TEXT UNIQUE)
├─ avatar (TEXT)
├─ username (TEXT)
├─ verified (BOOLEAN)
└─ created_at, updated_at (TIMESTAMPTZ)

wallets
├─ id (BIGSERIAL PRIMARY KEY)
├─ profile_id (BIGINT → profiles)
├─ balance (BIGINT)
├─ total_earned (BIGINT)
└─ balance_available (BIGINT)

referrals
├─ referrer_profile_id (BIGINT → profiles)
├─ referee_profile_id (BIGINT → profiles)
├─ completed_tasks_count (INTEGER)
├─ earnings (BIGINT)
└─ created_at (TIMESTAMPTZ)
```

## 🏗️ Архитектура

### Clean Architecture слои

```
┌─────────────────────────────────────┐
│   Transport (gRPC/HTTP Gateway)     │  internal/service/
├─────────────────────────────────────┤
│   Use Cases (Business Logic)        │  internal/usecase/
├─────────────────────────────────────┤
│   Repository (Data Access)          │  internal/repository/
├─────────────────────────────────────┤
│   Domain (Interfaces)                │  internal/domain.go
└─────────────────────────────────────┘
```

### Принципы

- ✅ Зависимости направлены внутрь
- ✅ Интерфейсы в domain.go
- ✅ Ошибки определены в usecase/models
- ✅ Транзакции управляются в usecase
- ✅ Валидация на уровне usecase

## 🔍 Особенности реализации

### Безопасная работа с транзакциями

```go
defer func() {
    if err != nil && tx != nil {
        _ = tx.Rollback(context.Background())
    }
}()
```

### Валидация входных данных

```go
func (i *RegisterByTelegramInput) Validate() error {
    if i.InitDataRaw == "" {
        return fmt.Errorf("%w: init_data_raw is required", ErrInvalidInput)
    }
    if len(i.InitDataRaw) > 10000 {
        return fmt.Errorf("%w: init_data_raw too long", ErrInvalidInput)
    }
    return nil
}
```

### Graceful error handling

```go
// Нет panic - ошибки возвращаются через каналы
errChan := make(chan error, 1)
go func() {
    if err := startServer(); err != nil {
        errChan <- err
    }
}()
```

## 📚 Дополнительные материалы

- [Детальная конфигурация](docs/configuration.md)
- [API документация (Swagger)](http://localhost:8090/swagger)
- [Postman коллекция](postman_collection.json)

## 🤝 Вклад в проект

Contributions are welcome! Пожалуйста:

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'feat: add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📝 Changelog

### [2025-11-08] - Major improvements

#### Исправлено
- 🔴 Критичные: UseCase дублирование, panic в горутине, обработка ошибок
- 🟠 Важные: deprecated grpc.Dial, time.Sleep, валидация, unsafe defer, дублирование кода
- 🟡 Незначительные: форматирование, обработка ошибок в bot

#### Добавлено
- Docker Compose setup
- Facebase API (profiles, wallets, referrals)
- Unit-тесты (8/8 passing)
- Postman коллекция

## 📄 Лицензия

MIT License - see [LICENSE](LICENSE) file for details

## 👥 Авторы

- Improvements: [@twelvepills-936](https://github.com/twelvepills-936)

---

⭐ Если проект полезен, поставьте звезду!

🐛 Нашли баг? [Создайте issue](https://github.com/twelvepills-936/goservice/issues)

💬 Есть вопросы? [Открывайте discussion](https://github.com/twelvepills-936/goservice/discussions)
