# Copilot Instructions for Sumni Finance Backend

This document provides context and guidelines for GitHub Copilot when assisting with code development in this project.

---

## 🏗️ Project Architecture

This project follows **Clean Architecture** with **Domain-Driven Design (DDD)** and **Hexagonal Architecture** (Ports & Adapters).

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP/External                         │
│                     (Framework & Drivers)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                          Ports                               │
│              (Interface Adapters - HTTP Handlers)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                      Application Layer                       │
│            (Use Cases - Commands & Queries)                  │
│                      CQRS Pattern                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                       Domain Layer                           │
│          (Entities, Value Objects, Repositories)             │
│              Pure Business Logic - No Dependencies           │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                       Adapters                               │
│         (Infrastructure - Database, External APIs)           │
└─────────────────────────────────────────────────────────────┘
```

### Domain Structure

1. **Finance Domain** (`internal/finance/`)

   - Core business logic for financial operations
   - Entities: AssetSource, Wallet, Transaction
   - Value Objects: Money, Currency
   - CQRS: Commands for writes, Queries for reads

2. **Auth Domain** (`internal/auth/`)

   - Keycloak integration (OAuth2/OIDC)
   - Token management and verification
   - Authentication middleware

3. **Common Domain** (`internal/common/`)
   - Shared utilities (CQRS, DB, Logging, Validation)
   - Cross-cutting concerns
   - Reusable value objects

---

## 🎯 Code Generation Guidelines

### When generating new features:

#### 1. Start with Domain Layer

```go
// internal/finance/domain/{entity}/
// - Define entity struct with business rules
// - Create value objects (immutable, validated)
// - Define repository interface
// - Write unit tests
```

#### 2. Application Layer (CQRS)

```go
// Commands (Write Operations)
// internal/finance/app/command/
type CreateEntityHandler struct {
    repo domain.EntityRepository
}

func (h CreateEntityHandler) Handle(ctx context.Context, cmd CreateEntity) error {
    // Validate, create domain entity, persist
}

// Queries (Read Operations)
// internal/finance/app/query/
type GetEntityHandler struct {
    queries *store.Queries
}

func (h GetEntityHandler) Handle(ctx context.Context, q GetEntity) (*EntityDTO, error) {
    // Fetch and return data
}
```

#### 3. Adapter Layer

```go
// internal/finance/adapter/db/
// - Implement repository interface
// - Use SQLC for type-safe queries
// - SQL queries in adapter/db/store/queries/
```

#### 4. Ports Layer

```go
// internal/finance/ports/
// - HTTP handlers
// - Request/Response DTOs
// - Route registration
```

---

## 📝 Code Style & Conventions

### Naming Conventions

- **Packages**: lowercase, no underscores (e.g., `assetsource`, not `asset_source`)
- **Interfaces**: `-er` suffix when appropriate (Reader, Handler)
- **Repository methods**: Domain language (e.g., `SaveAssetSource`, not `InsertAssetSource`)
- **Handlers**: `{Action}{Entity}Handler` (e.g., `CreateAssetSourceHandler`)

### Error Handling

```go
// ✅ Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create asset source: %w", err)
}

// ✅ Custom domain errors
var ErrAssetSourceNotFound = errors.New("asset source not found")

// ✅ Error types for specific cases
type ValidationError struct {
    Field   string
    Message string
}
```

### Context Usage

```go
// ✅ Always pass context as first parameter
func (h Handler) Handle(ctx context.Context, cmd Command) error

// ✅ Propagate context to all downstream calls
entity, err := h.repo.Find(ctx, id)
```

### Value Objects

```go
// ✅ Immutable, validated in constructor
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
    if amount.IsNegative() {
        return Money{}, errors.New("amount cannot be negative")
    }
    return Money{amount: amount, currency: currency}, nil
}

// ❌ No setters on value objects
```

### Repository Pattern

```go
// ✅ Interface in domain layer
type AssetSourceRepository interface {
    Save(ctx context.Context, source *AssetSource) error
    FindByID(ctx context.Context, id uuid.UUID) (*AssetSource, error)
}

// ✅ Implementation in adapter layer
type assetsourceRepo struct {
    pool    *pgxpool.Pool
    queries *store.Queries
}
```

---

## 🗄️ Database & SQLC

### SQLC Usage

- All queries in `internal/finance/adapter/db/store/queries/*.sql`
- Run `sqlc generate` after adding/modifying queries
- Use generated code in repository implementations

```sql
-- name: GetAssetSource :one
SELECT * FROM asset_sources WHERE id = $1;

-- name: CreateAssetSource :one
INSERT INTO asset_sources (id, name, type, details)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

### Migrations

- Location: `db/migrations/`
- Format: `YYYYMMDDHHMMSS_description.up.sql` / `*.down.sql`
- Always provide both up and down migrations

---

## 🧪 Testing Guidelines

### Test Structure

```go
func TestCreateAssetSource(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateInput
        want    *AssetSource
        wantErr bool
    }{
        {
            name: "returns error when ...",
            // ...
        },
        {
            name: "creates asset source successfully when ...",
            // ...
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Dependency Manager Pattern

Use a Dependency Manager for organizing test dependencies:

```go
type EntityDependenciesManager struct {
    entityRepoMock    *entity_mocks.MockRepository
    otherServiceMock  *service_mocks.MockService
}

func NewEntityDependenciesManager(t *testing.T) *EntityDependenciesManager {
    t.Helper()

    return &EntityDependenciesManager{
        entityRepoMock:   entity_mocks.NewMockRepository(t),
        otherServiceMock: service_mocks.NewMockService(t),
    }
}

func (dm *EntityDependenciesManager) NewHandler() command.EntityHandler {
    return command.NewEntityHandler(dm.entityRepoMock, dm.otherServiceMock)
}
```

### Test Structure Guidelines

1. **Test Table Fields**:
   - `name`: Descriptive test case name
   - `cmd` or `input`: Test input data
   - `setupMock`: Function to configure mocks, receives full dependency manager
   - `hasErr`: Boolean indicating if error is expected
   - `errorContains`: Optional string to validate specific error messages
   - `want`: Expected output (for non-error cases)

2. **Test Phases** (Arrange-Act-Assert):
   - **Arrange**: Setup dependencies and mocks
   - **Act**: Execute the function under test
   - **Assert**: Verify results and errors

3. **Naming Conventions**:
   - Test function: `TestEntity_Method`
   - Test cases: Start with action (e.g., "returns error when...", "creates entity successfully")
   - Use descriptive names that explain the scenario

4. **Mock Setup**:
   - Pass entire Dependency Manager to `setupMock`, not individual mocks
   - Comment when no mock setup is needed
   - Use `mock.Anything` for flexible matching
   - Always call `.Once()` to verify call count

### Test Coverage Requirements

- ✅ Unit tests for domain entities and value objects
- ✅ Handler tests with mocked dependencies (table-driven)
- ✅ Integration tests for repositories
- ✅ Validation error cases
- ✅ Repository failure scenarios
- ✅ Success cases

### Test Patterns to Follow

```go
// ✅ DO: Use table-driven tests
func TestHandler(t *testing.T) {
    tests := []struct{ /* ... */ }{}
    for _, tt := range tests { /* ... */ }
}

// ❌ DON'T: Individual test functions
func TestHandler_Case1(t *testing.T) { /* ... */ }
func TestHandler_Case2(t *testing.T) { /* ... */ }

// ✅ DO: Pass dependency manager to setupMock
setupMock: func(dm *EntityDependenciesManager) {
    dm.entityRepoMock.EXPECT().Create(...)
}

// ❌ DON'T: Pass individual mocks
setupMock: func(repo *MockRepository) {
    repo.EXPECT().Create(...)
}

// ✅ DO: Use clear Arrange-Act-Assert phases
// Arrange
dm := NewEntityDM(t)
// Act
err := handler.Handle(ctx, cmd)
// Assert
require.NoError(t, err)

// ❌ DON'T: Mix phases without clear separation
dm := NewEntityDM(t)
err := handler.Handle(ctx, cmd)
require.NoError(t, err)

// ✅ DO: Use realistic test data when appropriate
cmd: command.CreateFundProviderCmd{
    Name:         "Techcombank7316",
    CurrencyCode: "VND",
}

// ❌ DON'T: Use generic names everywhere
cmd: command.CreateFundProviderCmd{
    Name:         "Test Provider",
    CurrencyCode: "USD",
}
```

### Test Data Guidelines

- Use realistic names for entities (e.g., "Techcombank7316" for fund providers)
- Use appropriate currencies (VND for Vietnamese banks, USD for international)
- Keep validation test cases simple (empty strings, invalid values)
- Use `mock.Anything` for flexible argument matching
- Always verify mock expectations with `.Once()`, `.Times(n)`, etc.

### Running Tests

```bash
make test              # Run all tests
go test ./...          # Run all tests
go test -v ./...       # Verbose output
go test -race ./...    # Race detection
go test -cover ./...   # Coverage report
```

---

## 🔒 Security

### Authentication

- Keycloak for authentication (OAuth2/OIDC)
- Token verification via middleware
- Protected routes use `authHandler.AuthMiddleware`

### Best Practices

- Never log sensitive data (tokens, passwords)
- Use parameterized queries (SQLC handles this)
- Validate all external inputs
- No hardcoded secrets (use environment variables)

---

## 🚀 Common Tasks

### Adding a New Command Handler

1. **Define Command** in `internal/finance/app/command/`

   ```go
   type CreateEntityCmd struct {
       Name         string
       InitBalance  int64
       CurrencyCode string
   }
   ```

2. **Create Handler** in same directory

   ```go
   type CreateEntityHandler cqrs.CommandHandler[CreateEntityCmd]

   type createEntityHandler struct {
       entityRepo entity.Repository
   }

   func NewCreateEntityHandler(entityRepo entity.Repository) *createEntityHandler {
       return &createEntityHandler{
           entityRepo: entityRepo,
       }
   }

   func (h *createEntityHandler) Handle(ctx context.Context, cmd CreateEntityCmd) error {
       entity, err := entity.NewEntity(cmd.Name, cmd.InitBalance, cmd.CurrencyCode)
       if err != nil {
           return httperr.NewIncorrectInputError(err, "invalid-cmd")
       }

       err = h.entityRepo.Create(ctx, entity)
       if err != nil {
           return httperr.NewUnknowError(err, "failed-to-create-entity")
       }

       return nil
   }
   ```

3. **Write Tests** using table-driven pattern

   ```go
   func TestCreateEntity_Handle(t *testing.T) {
       tests := []struct {
           name      string
           cmd       command.CreateEntityCmd
           setupMock func(*EntityDependenciesManager)
           hasErr    bool
       }{
           // Test cases...
       }
       // Test implementation...
   }
   ```

### Adding a New Entity

1. **Domain**: Create entity in `internal/finance/domain/{entity}/`

   ```go
   type Entity struct {
       id   uuid.UUID
       name string
       // fields...
   }

   // Factory function with validation
   func NewEntity(name string, amount int64, currencyCode string) (*Entity, error) {
       v := validator.New()
       v.Required(name, "name")
       
       if err := v.Err(); err != nil {
           return nil, err
       }

       money, err := valueobject.NewMoney(amount, currencyCode)
       if err != nil {
           return nil, err
       }

       return &Entity{
           id:   uuid.New(),
           name: name,
           balance: money,
       }, nil
   }
   ```

2. **Repository Interface**: In domain directory

   ```go
   type Repository interface {
       Create(ctx context.Context, e *Entity) error
       GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
   }
   ```

3. **Migration**: Create in `db/migrations/`

   ```sql
   -- db/migrations/YYYYMMDDHHMMSS_create_entities.up.sql
   CREATE TABLE entities (
       id UUID PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       balance BIGINT NOT NULL,
       currency_code VARCHAR(3) NOT NULL,
       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
   );
   ```

4. **SQLC Queries**: Add to `adapter/db/store/queries/entity.sql`

   ```sql
   -- name: CreateEntity :one
   INSERT INTO entities (id, name, balance, currency_code)
   VALUES ($1, $2, $3, $4)
   RETURNING *;

   -- name: GetEntity :one
   SELECT * FROM entities WHERE id = $1;
   ```

5. **Repository Implementation**: In `adapter/db/entity_repository.go`

   ```go
   type entityRepository struct {
       pool    *pgxpool.Pool
       queries *store.Queries
   }

   func NewEntityRepository(pool *pgxpool.Pool) entity.Repository {
       return &entityRepository{
           pool:    pool,
           queries: store.New(pool),
       }
   }

   func (r *entityRepository) Create(ctx context.Context, e *entity.Entity) error {
       _, err := r.queries.CreateEntity(ctx, store.CreateEntityParams{
           ID:           e.ID(),
           Name:         e.Name(),
           Balance:      e.Balance().Amount(),
           CurrencyCode: e.Balance().Currency().Code(),
       })
       return err
   }
   ```

6. **Commands**: In `app/command/` (see "Adding a New Command Handler" above)

7. **HTTP Handlers**: In `ports/`

   ```go
   func (h *HttpServer) CreateEntity(w http.ResponseWriter, r *http.Request) (response.Response, error) {
       var req CreateEntityRequest
       if err := mapToStruct(r, &req); err != nil {
           return nil, err
       }

       cmd := command.CreateEntityCmd{
           Name:         req.Name,
           InitBalance:  req.InitBalance,
           CurrencyCode: req.CurrencyCode,
       }

       err := h.app.Commands.CreateEntity.Handle(r.Context(), cmd)
       if err != nil {
           return nil, err
       }

       return response.EmptyResponse{}, nil
   }
   ```

8. **Wire Up**: 
   - Update `app/app.go` to add command handler
   - Update `ports/http.go` to register routes

### Adding a New Endpoint

1. Define handler in `internal/finance/ports/`
2. Add route in `ports/http.go`
3. Update `cmd/server/main.go` if needed

---

## 🛠️ Development Workflow

### Local Development

```bash
make dev              # Start with hot reload
make dev DEBUG=true   # Start with debugger
make test             # Run tests
make stop             # Stop containers
```

### Before Committing

```bash
go fmt ./...
go vet ./...
golangci-lint run
go test -race ./...
```

---

## 📚 Key Principles

1. **Dependency Rule**: Dependencies point inward (domain has no external deps)
2. **CQRS**: Separate models for reads (queries) and writes (commands)
3. **Repository Pattern**: Abstract data access, interface in domain
4. **Value Objects**: Immutable, self-validating
5. **Aggregates**: Consistency boundaries for transactions
6. **Single Responsibility**: Each component has one reason to change
7. **Interface Segregation**: Small, focused interfaces
8. **Dependency Inversion**: Depend on abstractions, not concretions

---

## 🎓 Reference Documentation

- [FILE_STRUCTURE.md](FILE_STRUCTURE.md) - Complete project structure
- [CODE_REVIEW_GUIDELINES.md](CODE_REVIEW_GUIDELINES.md) - Review checklist
- [README.md](../README.md) - Setup and running instructions

---

## 📖 Reference Implementations

**When in doubt**: Follow the existing patterns in the codebase.

Reference implementations:
- **Domain Entity**: `internal/finance/domain/fundprovider/` or `internal/finance/domain/wallet/`
- **Command Handler**: `internal/finance/app/command/create_fund_provider.go`
- **Handler Tests**: `internal/finance/app/command/create_fund_provider_test.go` or `create_wallet_test.go`
- **Repository**: `internal/finance/adapter/db/fund_provider_repository.go`
- **HTTP Handler**: `internal/finance/ports/create_fund_provider_handler.go`

These serve as canonical examples of the architecture and testing patterns described in this document.
