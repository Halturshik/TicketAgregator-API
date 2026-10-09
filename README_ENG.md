# Ticket Aggregator API

A Go backend pet project for an air, rail, and bus ticket aggregator. It supports user registration, searching offers from multiple suppliers, multi-passenger orders, mock payments, loyalty points, partial and full refunds, and guest access to bookings.

The project is built as a modular monolith with a separate gRPC process that simulates external suppliers. PostgreSQL stores business data. Redis is used for short-lived state, search result caching, and rate limiting. Supplier interactions use a Protocol Buffers contract.

## Features

- Two-step registration and login, and password recovery by email code.
- Access and refresh JWTs, refresh-token rotation, and reuse detection.
- Search across air, rail, and bus routes, including round trips and air connections.
- Concurrent requests to three mock suppliers and offer aggregation.
- Passenger-level and group pricing, fare types, and refund rules.
- Multi-passenger orders with document validation.
- Mock payments, 2% loyalty-point accrual, and redemption of up to 50% of the order price.
- History of paid and refunded orders.
- Partial and full refunds with idempotency and retries for unfinished operations.
- Public booking and trip lookup without exposing personal data.
- Email verification for guest access to booking details and refunds.
- Health checks, graceful shutdown, structured logging, and request IDs.
- Unit, integration, and race tests in GitHub Actions.

## Tech Stack

- Go 1.27.1
- chi for HTTP routing
- PostgreSQL 17 and Goose migrations
- Redis 7.4
- gRPC and Protocol Buffers
- Docker and Docker Compose
- GitHub Actions

## Quick Start

Docker and Docker Compose are required.

1. Create a local configuration file:

   ```bash
   cp .env.example .env
   ```

   In PowerShell:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Replace `JWT_SECRET` and `DOCUMENT_VERIFICATION_SECRET` in `.env` with long, randomly generated values.

3. Build and start the full stack:

   ```bash
   docker compose up --build
   ```

Migrations for the API and supplier databases run automatically at startup. Once the services are ready, the API is available at `http://localhost:8080`.

Check service health:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

Stop the services without deleting data:

```bash
docker compose down
```

Delete local environment data:

```bash
docker compose down -v
```

## Running Go Locally

Start the infrastructure separately:

```bash
docker compose up -d api-db supplier-db redis
```

Then start the supplier simulator and API in separate terminals:

```bash
go run ./cmd/supplier-simulator
go run ./cmd/api
```

This setup uses the addresses from `.env`: API PostgreSQL on port `5432`, supplier PostgreSQL on `5433`, Redis on `6379`, and the supplier gRPC service on `9090`.

The mock mailer prints verification codes to the application console. The mock payment provider declines approximately 3% of payments.

## Tests

Run the standard test suite:

```bash
go test ./... -count=1
```

Run PostgreSQL integration tests in PowerShell:

```powershell
$env:TEST_POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable'
go test -tags=integration ./internal/integration -count=1 -timeout=10m
```

On Linux/macOS:

```bash
TEST_POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable' \
  go test -tags=integration ./internal/integration -count=1 -timeout=10m
```

GitHub Actions also runs `go vet`, the race detector, PostgreSQL integration tests, and builds both Docker images.

## Documentation

- [Architecture](docs/architecture_eng.md) — domains, dependencies, storage, gRPC, supplier model, and concurrency.
- [Main flows](docs/flows_eng.md) — authentication, search, orders, payment, refunds, history, and guest access.
- [OpenAPI 3.0](docs/openapi_eng.yaml) — the HTTP contract, also available through the embedded Swagger UI.
- [Supplier protobuf](proto/supplier/v1/supplier.proto) — the internal gRPC contract.

After starting the API, Swagger UI is available at `http://localhost:8080/swagger/`. The English OpenAPI copy is available in the repository at `docs/openapi_eng.yaml`.

## Project Structure

```text
cmd/api/                    HTTP API and background tasks
cmd/supplier-simulator/     gRPC supplier simulator
internal/                   business domains and platform code
migrations/                 main PostgreSQL schema
supplier-migrations/        minimal supplier simulator schema
proto/                      Protocol Buffers contracts
docs/                       architecture, flows, and OpenAPI
```

## Project Scope

This is an educational aggregator, not a real booking system. Search offers, email delivery, document checks, and payment processing are mocked. Transaction boundaries, refund idempotency, ticket snapshots, guest access protection, health checks, and API-to-supplier communication are implemented as independent engineering mechanisms.
