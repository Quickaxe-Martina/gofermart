MIGRATIONS_DIR=./migrations
MIGRATE_BIN=$(shell go env GOPATH)/bin/migrate
DB_URL=postgres://test:test@localhost:5433/test?sslmode=disable
GO_IMPORTS_BIN=$(shell go env GOPATH)/bin/goimports

start:
	@$(GO_IMPORTS_BIN) -w .
	@go run ./cmd/gophermart/main.go -d="host=127.0.0.1 port=5433 user=test password=test dbname=test sslmode=disable" -r="127.0.0.1:8081"

accrual:
	cmd/accrual/accrual_linux_amd64 -a="127.0.0.1:8081" -d="host=127.0.0.1 port=5433 user=test password=test dbname=accrual sslmode=disable"

build:
	@$(GO_IMPORTS_BIN) -w .
	@go build -o ./cmd/gophermart/gophermart ./cmd/gophermart/main.go

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=seq"; \
		exit 1; \
	fi
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
