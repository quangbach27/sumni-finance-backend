# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Sumni Finance Backend is a Go monolith (module `sumni-finance-backend`, Go 1.27) built as a set of self-contained
modules behind a single HTTP server (Echo v5). It currently has one active module: **treasury** (fund sources,
wallets, transactions/ledger). There is no `Makefile` — all tasks are run through `task` (Taskfile.yml). Auth is not
wired up yet (`internal.ExternalService` is an empty placeholder in `cmd/main.go`, and HTTP handlers use a hardcoded
`mockedTenantContext()` in `internal/treasury/ports/http/handler.go` until real auth middleware lands).

## Commands

All commands go through [Task](https://taskfile.dev) (`task <name>`), not `make`.

```bash
task up                 # docker compose up (app + postgres)
task up-clean            # down then up (fresh containers)
task down                # docker compose down
task down-volumes        # docker compose down -v (wipes DB volume)

task test                # integration tests (superset of unit) + component tests
task test-unit            # go test ./internal/...
task test-integration     # go test -tags integration ./internal/... (needs .env.test / a running postgres)
task test-component       # go test ./tests/...
task domain-coverage      # enforce >=85% coverage on internal/**/domain and internal/common/shared

task gen                 # go generate ./internal/... (sqlc + oapi-codegen), then task fmt
task lint                 # golangci-lint run ./...
task fmt                  # go tool gofumpt -l -w .
task tidy                 # go mod tidy
```

Run a single test the normal Go way, e.g.:
```bash
go test ./internal/treasury/domain/... -run TestFundSource_Reserve_Success -v
go test -tags integration ./internal/treasury/adapters/db/... -run TestCreateWallet_SavesWalletAndAllocations -v
```

**Test build tags matter.** Files under `internal/*/adapters/db/*_test.go` use `//go:build integration` and need a
live Postgres (`.env.test`: `postgres://user:password@localhost:5432/sumni-finance?sslmode=disable`, i.e. `task up`
first). Files under `tests/**` use `//go:build component`. Domain package tests (`internal/*/domain/*_test.go`) have
no build tag and are pure unit tests with no DB dependency.

Migrations run automatically on module `Init()` (see `internal/treasury/module.go`) via
`internal.common.MigrateDatabaseUp`, embedding `adapters/db/migrations/*.sql` with golang-migrate — there's no
separate `migrate up` command to remember.

`lefthook.yml` runs `task test-unit`, `task lint`, and `task fmt` automatically as a git `pre-push` hook —
expect a push to fail if any of those fail locally.

### Code generation

Two independent `go generate` pipelines per module, both invoked by `task gen`:
- `internal/treasury/adapters/db/generate.go` → `sqlc generate` (reads `sqlc.yaml` + `queries/*.sql`, writes
  `dbmodels/*.sql.go`). Edit `.sql` files in `queries/`, never the generated `dbmodels/` output by hand. `engine:
  "postgresql"` in `sqlc.yaml`, so query params use the `@arg_name` shorthand for `sqlc.arg(arg_name)` (e.g.
  `WHERE tenant_id = @tenant_id`) — not supported under a MySQL engine, but fine here.
- `internal/treasury/ports/http/generate.go` → `oapi-codegen --config=oapi-codegen.yaml openapi.yaml`, writing
  `openapi.gen.go` (server-side strict interfaces). There's a matching client generator under
  `internal/treasury/ports/http/client/`. Edit `openapi.yaml`, never `openapi.gen.go` by hand.

A PR labeled `publish-contract` triggers a CI workflow that copies every `**/ports/http/openapi.yaml` into the
`sumni-finance-frontend-vite` repo and opens a PR there — so keep `openapi.yaml` the source of truth for the wire
contract.

## Architecture

Hexagonal / ports-and-adapters, one Go package tree per bounded-context **module**, wired together in
`internal/svc.go`. Each module implements `internal/common/module.Module`:

```go
type Module interface {
    Name() Name
    Init(ctx context.Context) error                                            // migrate DB, build repos/handlers
    RegisterHttp(ctx context.Context, publicRouter, protectedRouter common.EchoRouter) error
    RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error // cross-module in-process contracts
}
```

`internal.New()` builds all modules, calls `Init` → `RegisterContracts` → (verify contracts) → `RegisterHttp`, in that
order, for every module in the `modules` slice. **To add a new module**, mirror `internal/treasury/module.go` and
append it to the `modules` slice in `internal/svc.go`. `RegisterContracts`/`contracts.Contracts` is the seam for
future cross-module, in-process calls (currently unused/empty — see `internal/common/module/contracts`).


### Layout inside a module (using `internal/treasury/` as the reference)

```
domain/            pure business logic: entities, value objects, repository interfaces. No DB/HTTP imports.
app/command/        write use cases (CQRS "commands"), one Handlers struct holding repo dependencies, one file per command
app/query/           read use cases, backed by dedicated read models (not the write repositories)
adapters/db/         repository implementations (pgx + sqlc), migrations/, queries/, generated dbmodels/
ports/http/           oapi-codegen strict-server interface + Handler implementing it, openapi.yaml is the contract
module.go            wires domain/app/adapters together, implements common/module.Module
```

Reads and writes are deliberately split: `app/query` handlers use `*_read_model.go` types
(`adapters/db/fund_source_read_model.go`, `wallet_read_model.go`) that query the DB directly for DTOs, while
`app/command` handlers go through domain-layer `Repository` interfaces (`domain.WalletRepository`,
`domain.FundSourceRepository`) whose implementations reconstruct/persist aggregates.

Repository interfaces frequently take a closure so invariants are enforced with data loaded inside a single
transaction, e.g. `WalletRepository.CreateWallet(ctx, tenantContext, fundSourceUUIDs, func(fundSources map[...]) (*Wallet, error))`
— the closure receives the fund sources loaded (and locked) inside the same DB transaction and returns the
constructed aggregate to persist. Follow this pattern instead of doing separate load-then-save calls when an
invariant spans multiple aggregates.

Treasury-specific domain conventions (ID/enum/error/money/tenancy patterns) and the transaction lifecycle state
machine are documented in `internal/treasury/CLAUDE.md` — read that when working under `internal/treasury/domain/`.

### HTTP layer

Each module's `openapi.yaml` is the API contract; `oapi-codegen` generates strict request/response types and a
`StrictServerInterface` that `ports/http/handler.go`'s `Handler` implements. `module.go`'s `RegisterHttp` calls
`treasuryhttp.Register(protectedRouter, handler)` to mount routes (`internal/treasury/ports/http/openapi.gen.go` has
the generated `Register*` function). Pagination on list endpoints follows `Page`/`PageSize` query params with a
`defaultPageSize` constant per handler package.

## Domain conventions (internal/treasury/domain, internal/common)

- **IDs**: every aggregate has a `<Name>UUID struct { common.UUID }` wrapper type (`WalletUUID`, `FundSourceUUID`,
  `TransactionUUID`) rather than passing raw `uuid.UUID`/`shortuuid` around. New UUIDs are created with
  `common.NewUUIDv7()`.
- **Enums**: modeled with the generic `common.Enum[T]` (`internal/common/enum.go`) — a type implements
  `Values() []string`, then `common.MustEnum[Wrapper, T]("VALUE")` builds validated constants (see
  `TransactionStatus`/`TransactionStatusValues` in `domain/transaction.go`, `shared.EntryType`,
  `shared.Currency`). `Enum` implements `Scan`/`Value` (DB) and `MarshalText`/`UnmarshalText` (JSON), so new enums
  get persistence + JSON handling for free — don't hand-roll string constants for domain enums.
- **Errors**: domain/app code returns `common.Error` (`internal/common/errors.go`) built via
  `common.NewInvalidInputError`, `NewNotFoundError`, `NewConflictError`, etc. Each carries an `ErrorSlug` and an
  HTTP status code baked in; `internal/common/errors_echo.go` translates these to Echo responses at the edge. Prefer
  these constructors over bare `errors.New`/`fmt.Errorf` for anything that crosses into `app/` or is returned from a
  handler; collect multiple validation failures into `[]common.ErrorDetails` and attach with `.WithDetails(...)`
  (see `Transaction`/`TenantContext` constructors) rather than returning on the first error.
- **Money/Currency**: `internal/common/shared/money.go` + `currency.go` — `shared.Money` wraps `decimal.Decimal` +
  `shared.Currency` and validates on construction (`shared.NewMoney`, `shared.MustNewCurrency` for tests). Always
  build amounts through these constructors instead of using `decimal.Decimal` directly in domain code, so
  currency-mismatch and sign checks happen consistently.
- **Multi-tenancy**: `shared.TenantContext` (tenantID + officeID) is threaded explicitly through command/query
  structs and repository methods — there's no ambient context lookup. Real auth isn't wired up yet, so HTTP handlers
  currently synthesize it via `mockedTenantContext()`; when wiring real auth, that's the seam to replace.

## CI (.github/workflows)

- `commit-stage.yml`: on push/PR to `main` — spins up a Postgres service container, runs `golangci-lint`
  (v2.11.4), `task test`, `task domain-coverage` (`continue-on-error: true`, i.e. informational only), `go build
  ./...`, then an Anchore container scan; on push to `main` it also builds/publishes the Docker image
  (`docker/app-prod/Dockerfile`) and dispatches to the acceptance stage, which triggers a deploy to staging in a
  separate `sumni-finance-deployment` repo.

## Local dev environment

`docker-compose.yml` runs the app (hot-reload via `reflex`, see `docker/app-local/`) plus a `postgres:17.6-alpine`
container; `POSTGRES_URL` is read from `.env`. 
