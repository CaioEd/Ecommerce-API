# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
cp .env.example .env      # required on first run; defaults assume local Postgres
go mod tidy               # sync dependencies
go run .                  # start server on :8080 (SERVER_PORT), runs AutoMigrate on boot
go build -o bin/api .     # compile
go vet ./...              # static checks
gofmt -l .                # list unformatted files
```

```bash
make test                                  # all unit tests
make test-one T=TestName PKG=./internal/service   # filter by regex; T also matches subtests
make cover                                 # per-function coverage
make test-api                              # Postman collection via Newman (needs the API running)
```

Unit tests cover `internal/service` and `internal/handler`, each with a hand-written double of the layer below defined in the test file itself (`fakeUserRepository`, `fakeUserService`) — no mocking library. Both doubles carry a compile-time `var _ Interface = (*fake)(nil)` assertion, so changing a layer's interface breaks its tests loudly. `internal/repository` has no unit tests by design: it is thin GORM calls, and the Postman collection in `postman/` covers that path end to end.

A running PostgreSQL instance is required to start the app — `database.Connect` fails fast and `main` calls `log.Fatalf`.

## Architecture

Layered architecture; each layer depends only on the one below it:

```
HTTP -> internal/handler -> internal/service -> internal/repository -> Postgres (GORM)
```

All dependencies are wired manually in `main.go` (`NewUserRepository` -> `NewUserService` -> `NewUserHandler` -> `router.New`). There is no DI container. `repository` and `service` each export an interface plus an unexported struct implementation, so the layer above depends on the interface and can be faked in tests.

**Error handling is the load-bearing convention.** Sentinel errors are declared by the layer that owns the condition and bubble up unchanged:
- `repository.ErrUserNotFound` — returned in place of `gorm.ErrRecordNotFound`; GORM errors never escape the repository.
- `service.ErrEmailAlreadyExists` — uniqueness is enforced in the service via `FindByEmail`, not by catching the DB unique-index violation.

Only the handler maps these to HTTP status codes (`errors.Is` -> 404 / 409, everything else -> 500 with a generic message). Do not translate errors to status codes anywhere else.

**DTO boundary.** Handlers and services speak `internal/dto` types; only the repository touches `internal/model`. `model.User.Password` is `json:"-"` and `dto.UserResponse` omits it entirely — never return `model` structs from a handler. Request validation uses Gin `binding` tags on the DTOs (`ShouldBindJSON`), and partial updates use pointer fields (`*string`) so "absent" is distinguishable from "empty".

**Schema.** `database.Migrate` calls `AutoMigrate` — there are no migration files. New entities must be registered there or their table will not exist. Deletes are soft (`gorm.DeletedAt`), so records survive `DELETE /users/:id` and are filtered out of subsequent queries automatically.

**Routes** live only in `internal/router/router.go`, under the `/api/v1` group; `/health` sits outside the versioned group.

## Adding a new entity

Replicate the `User` slice end to end: `model` (with `TableName()`) -> register in `database.Migrate` -> `repository` (interface + GORM impl, translating `gorm.ErrRecordNotFound`) -> `dto` (request/response + `To*Response` converters) -> `service` -> `handler` -> wiring in `main.go` -> route group in `router.New`.

## Conventions

Code comments, log messages, and user-facing error strings are written in Portuguese (pt-BR); identifiers are in English. Match this when editing.
