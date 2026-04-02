SERVICE := sumni-finance-backend

.PHONY: test
test:
	@./scripts/test.sh .e2e.env

.PHONY: test-ci
test-ci: 
	@CI=true ./scripts/test.sh .e2e.env

.PHONY: dev
dev:
	DEBUG=$(DEBUG) docker compose up --build $(SERVICE) -d	

.PHONY: logs
logs:
	docker logs -f $(SERVICE)

.PHONY: stop
stop:
	docker compose down $(SERVICE)

.PHONY: down
down:
	docker compose down -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate -f ./internal/$(DOMAIN)/adapter/db/store/sqlc.yml

.PHONY: openapi_http
openapi_http:
	@./scripts/openapi-http.sh $(SERVICE) $(OUTPUT) $(PACKAGE)

## -------------------------------------
# Database Config Variables (Use environment variables if set, otherwise use default placeholders)
POSTGRES_USER ?= sumni
POSTGRES_PASSWORD ?= sumni
POSTGRES_HOST ?= localhost
POSTGRES_DATABASE ?= sumni-finance
POSTGRES_PORT ?= 5432

DB_URL := postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable
MIGRATE_PATH := db/migrations

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(NAME)

.PHONY: migrate-up
migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATE_PATH) up

.PHONY: migrate-down
migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATE_PATH) down 1

.PHONY: migrate-status
migrate-status:
	@echo "Checking migration status..."
	@migrate -database "$(DB_URL)" -path $(MIGRATE_PATH) version
