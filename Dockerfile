# syntax=docker/dockerfile:1

FROM golang:1.27.1-alpine AS source
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

FROM source AS api-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ticket-api ./cmd/api

FROM source AS supplier-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/supplier-simulator ./cmd/supplier-simulator

FROM source AS supplier-healthcheck-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/supplier-healthcheck ./cmd/supplier-healthcheck

FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app

FROM runtime AS api
COPY --from=api-builder --chown=app:app /out/ticket-api ./ticket-api
COPY --chown=app:app migrations ./migrations
USER app
EXPOSE 8080
ENTRYPOINT ["./ticket-api"]

FROM runtime AS supplier
COPY --from=supplier-builder --chown=app:app /out/supplier-simulator ./supplier-simulator
COPY --from=supplier-healthcheck-builder --chown=app:app /out/supplier-healthcheck ./supplier-healthcheck
COPY --chown=app:app supplier-migrations ./supplier-migrations
USER app
EXPOSE 9090
ENTRYPOINT ["./supplier-simulator"]
