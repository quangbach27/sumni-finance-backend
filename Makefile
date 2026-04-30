SERVICE := sumni-finance-backend

.PHONY: up
up:
	docker compose up ${SERVICE}

.PHONY: down
down:
	docker compose down

.PHONY: down-volumes
down-volumes:
	docker compose down -v

# Test
.PHONY: test
test:
	$(MAKE) test-unit
	$(MAKE) test-integration
	$(MAKE) test-component

.PHONY: test-unit
test-unit:
	go test -race -v -timeout=2m ./...

.PHONY: test-integration
test-integration:
	go test -tags=integration -v -count=1 ./...

.PHONY: test-component
test-component:
	go test -tags=component -v -count=1 ./tests/...

# Tooling
.PHONY: gen
gen:
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go tool gofumpt -l -w .

.PHONY: mod-update
mod-update:
	go get -u=patch ./...
	go mod tidy