# AGENTS.md — Mutt

Open-source error tracking backend in Go (Fiber + GORM + PostgreSQL + Redis).

## Quick start

```bash
cp .env.example .env   # fill in real values
go run ./cmd/main.go   # or: make run
```

Server starts on `PORT` from `.env` (default `:8080`). Health check: `GET /ping`.

## Running tests

```bash
go test ./...                                          # all tests
go test ./server/handler/ -v                            # handler tests
go test ./server/handler/ -run TestBackup -v            # single test (also: make BackupTest)
go test ./internal/config/ -v                           # config tests
go test ./internal/middleware/ -v                       # middleware tests
```

Tests use **SQLite in-memory** + **miniredis** — no external services needed. Env vars are set in `TestMain` helpers (`server/handler/helper_test.go`, `consts/server_test.go`).

## Architecture

```
cmd/main.go            Entry point: loads env, connects DB/Redis, inits tracing, starts Fiber
internal/config/       DB (GORM/Postgres), Redis, env loading, OTel tracing setup
internal/middleware/    JWT auth, API key auth, rate limiting, security headers
internal/service/      Business logic: password hashing, JWT, API keys, fingerprinting, Redis ops
internal/utils/        Validation helpers, backup compression
server/handler/        HTTP handlers: auth, projects, error groups, backup, ingest
server/routes/         Route registration (all routes defined here)
models/                GORM models + request/response DTOs
consts/                Env-backed constants (port, hash cost, limits, backup thresholds)
```

## Key conventions

- **Init order matters**: `cmd/main.go:init()` calls `MustLoadEnv` → `MustConnectToDB` → `MustSyncDatabase` → `MustConnectRedis` → `MustInitJWT` → `InitTracing`. All `Must*` functions panic on failure.
- **Auto-migration**: Schema is auto-synced on startup via `DB.AutoMigrate()` in `internal/config/syncDatabase.go`. No migration files.
- **Error grouping**: Fingerprint = SHA-256 of `(stackTrace + title)` via `service.ComputeFingerprint()`.
- **Auth split**: Dashboard uses JWT (access/refresh tokens in cookies). SDK ingestion uses API key via `X-Mutt-Key` header (SHA-256 hashed for lookup).
- **Ownership checks**: All project/error queries filter by `user_id` to prevent IDOR.
- **Rate limits**: Redis-backed. Ingest: 100/min per project. Backup export: 5/min per user. Backup import: 2/min per user.
- **Validation**: Uses `go-playground/validator` via `.Validate()` methods on request DTOs in `models/`.

## Testing patterns

- `setupTestDB(t)` / `teardownTestDB(t)` in `server/handler/helper_test.go` — creates in-memory SQLite, registers otelgorm plugin, runs migrations.
- `seedUser(t, ...)` helper for creating test users.
- Each test file builds its own `fiber.App` with only the routes it needs (not the full app).
- Token is extracted from `access_token` cookie after login via `app.Test()`.

## Gotchas

- **`.env` contains real credentials** — never commit secrets. The checked-in `.env` has a NeonDB connection string and JWT secret.
- **No linter or formatter configured** — follow existing code style manually.
- **No CI/CD** — no GitHub Actions or pre-commit hooks.
- **Go 1.25** required (per `go.mod`), not 1.21 as `CONTRIBUTING.md` states.
- **Backup exports gzip** when payload exceeds 10KB (`consts.GzipThreshold`). Import max: 5MB / 100k records.
- **Tracing**: OTel is initialized on startup but gracefully degrades if collector is unreachable. `OTEL_EXPORTER_OTLP_INSECURE=true` needed for local dev.
- **Soft deletes**: Projects use `gorm.Model` which includes `Deleted_at`. Queries must account for this.
