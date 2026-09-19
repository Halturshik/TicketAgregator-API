# Ticket Aggregator API

Backend пет-проекта агрегатора авиа-, железнодорожных и автобусных билетов. Сервис умеет регистрировать пользователей, искать предложения нескольких поставщиков, оформлять мультипассажирные заказы, проводить mock-оплату, начислять бонусы, выполнять частичные и полные возвраты и предоставлять гостевой доступ к купленным билетам.

Проект построен как модульный монолит с отдельным gRPC-процессом, который имитирует внешних поставщиков. PostgreSQL хранит бизнес-данные, Redis используется для короткоживущего состояния, кеша поиска и rate limit, а взаимодействие с поставщиками вынесено в protobuf-контракт.

## Возможности

- регистрация, двухэтапный вход и восстановление пароля по коду;
- access- и refresh-JWT, ротация refresh-токенов и защита от повторного использования;
- поиск билетов по трём видам транспорта с поездками туда-обратно и авиапересадками;
- конкурентный запрос трёх mock-поставщиков и объединение их предложений;
- цены для одного пассажира и всей группы, тарифы и правила возврата;
- заказы на нескольких пассажиров с проверкой документов;
- mock-оплата, начисление 2% бонусов и списание до 50% стоимости;
- история оплаченных и возвращённых заказов;
- частичный или полный возврат с идемпотентностью и повторной обработкой незавершённых операций;
- публичный поиск бронирования и рейса без раскрытия персональных данных;
- подтверждение email для доступа гостя к деталям заказа и возврату;
- health checks, graceful shutdown, structured logging и request ID;
- unit-, integration- и race-тесты в GitHub Actions.

## Технологии

- Go 1.27.1;
- chi для HTTP routing;
- PostgreSQL 17 и goose migrations;
- Redis 7.4;
- gRPC и Protocol Buffers;
- Docker и Docker Compose;
- GitHub Actions.

## Быстрый запуск

Понадобятся Docker и Docker Compose.

1. Создайте локальную конфигурацию:

   ```bash
   cp .env.example .env
   ```

   В PowerShell:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Замените значения `JWT_SECRET` и `DOCUMENT_VERIFICATION_SECRET` в `.env` на длинные случайные строки.

3. Соберите и запустите весь стек:

   ```bash
   docker compose up --build
   ```

Миграции основной и supplier-БД применяются автоматически при старте процессов. После готовности сервис доступен по адресу `http://localhost:8080`.

Проверка состояния:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

Остановка без удаления данных:

```bash
docker compose down
```

Для полного сброса данных локального окружения:

```bash
docker compose down -v
```

## Локальный запуск Go

Инфраструктуру можно поднять отдельно:

```bash
docker compose up -d api-db supplier-db redis
```

Затем в разных терминалах запустить supplier и API:

```bash
go run ./cmd/supplier-simulator
go run ./cmd/api
```

При таком запуске используются адреса из локального `.env`: PostgreSQL API на `5432`, PostgreSQL supplier на `5433`, Redis на `6379`, gRPC supplier на `9090`.

Mock-почта выводит verification-коды в консоль приложения. Mock-оплата отклоняется примерно в 3% случаев.

## Тесты

Обычные тесты:

```bash
go test ./... -count=1
```

PostgreSQL integration tests в PowerShell:

```powershell
$env:TEST_POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable'
go test -tags=integration ./internal/integration -count=1 -timeout=10m
```

В Linux/macOS:

```bash
TEST_POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable' \
  go test -tags=integration ./internal/integration -count=1 -timeout=10m
```

GitHub Actions дополнительно запускает `go vet`, race detector, PostgreSQL integration tests и сборку обоих Docker-образов.

## Документация

- [Архитектура](docs/architecture.md) — домены, зависимости, хранилища, gRPC, supplier-модель и concurrency.
- [Основные сценарии](docs/flows.md) — auth, поиск, заказ, оплата, возврат, история и гостевой доступ.
- [OpenAPI 3.0](docs/openapi.yaml) — формальный HTTP-контракт, доступный через встроенный Swagger UI.
- [Supplier protobuf](proto/supplier/v1/supplier.proto) — внутренний gRPC-контракт.

После запуска API Swagger UI доступен по адресу `http://localhost:8080/swagger/`, а исходный OpenAPI-файл — по адресу `http://localhost:8080/swagger/openapi.yaml`.

## Структура

```text
cmd/api/                    HTTP API и фоновые процессы
cmd/supplier-simulator/     gRPC-симулятор поставщиков
internal/                   домены и платформенный код
migrations/                 схема основной PostgreSQL
supplier-migrations/        минимальная схема supplier simulator
proto/                      protobuf-контракты
docs/                       архитектура, сценарии и OpenAPI
```

## Границы проекта

Это учебный агрегатор, а не система реального бронирования. Поисковые предложения, отправка email, проверка документов и проведение оплаты имитируются. При этом транзакционные границы, идемпотентность возвратов, snapshots билетов, защита гостевого доступа, health checks и взаимодействие API с supplier реализованы как самостоятельные инженерные механизмы.
