# IronLogs Backend — Copilot Instructions

## Project

Go REST API for a gym tracking app. Stack: **Golang, GIN, GORM, Supabase**.

## Goals

- Apply Go best practices, not just make it work
- Structure must be production-ready and unit-testable
- Always consider testability when suggesting or generating code

## Project Structure

```
internal/<domain>/   # Feature domains (auth, user, routine, exercise, session)
  handler.go         # HTTP translation only — no business logic
  service.go         # Business logic — no HTTP, no DB details
  repository.go      # Data access — GORM queries only
  dto.go             # Request/response types for this domain
  model.go           # DB model struct
  errors.go          # Sentinel errors produced by this domain

pkg/                 # Generic utilities — zero imports from internal/
routes/              # URL registration only — no dependency wiring
main.go              # Owns all dependency wiring (wire everything here)
## Architecture Rules

- **Dependency injection only** — never instantiate dependencies inside handlers or services; accept them via constructor parameters
- **Interface-based dependencies** — every service and repository must be defined behind an interface so it can be mocked in tests
- **Strict layer separation**:
  - Handlers translate HTTP ↔ domain types; call services, nothing else
  - Services own business logic; call repository interfaces, nothing else
  - Repositories own data access; call GORM, nothing else
- **Errors live in the domain that produces them** — e.g., `internal/auth/errors.go` defines `ErrInvalidCredentials`
- **Each layer only knows the interface of the layer below it** — no skipping layers

## Conventions

- Return `error` as the last return value; wrap errors with `fmt.Errorf("context: %w", err)`
- Use Go table-driven tests with `t.Run` subtests for all new test files
- Mock interfaces with `testify/mock` or hand-written fakes — never hit a real DB in unit tests
- Validate request DTOs at the handler boundary using GIN binding tags
- Keep `pkg/` packages importable from anywhere — they must not import `internal/`
- Place new domains under `internal/<domain>/` following the six-file layout above
- `*config.Config` is passed to services and contains `cfg.DB (*gorm.DB)` — do not suggest adding a separate db field to structs that already hold cfg
## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Short, meaningful variable names; avoid stutter (`auth.AuthService` → `auth.Service`)
- Export only what consumers need; keep internals unexported
