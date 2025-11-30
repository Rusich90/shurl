include .env
export

MIGRATIONS_DIR := ./migrations

check-database-dsn:
ifndef DATABASE_DSN
	$(error DATABASE_DSN is not set. Please create .env file or set environment variable)
endif

build:
	cd cmd/shortener && go build -o shortener *.go

test:
	SERVER_PORT=8888 shortenertest -test.v -test.run=^TestIteration4$$ -binary-path=cmd/shortener/shortener -server-port=$SERVER_PORT

my-test:
	go test -v ./...

my-test-cov:
	go test -v ./... -coverprofile=coverage.out -coverpkg=./...
	go tool cover -func=coverage.out | grep "total:"

show-test-cov:
	go tool cover -html=coverage.out

run:
	go run cmd/shortener/*.go

migrate-up: check-database-dsn
	@echo "Applying migrations"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" up

migrate-down: check-database-dsn
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" down

migrate-version: check-database-dsn
	@echo "Current migration version:"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" version

migrate-create: check-database-dsn
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=migration_name)
endif
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)

.PHONY: build test run