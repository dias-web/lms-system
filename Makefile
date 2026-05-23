.PHONY: help build run tidy test docker-up docker-down docker-build migrate-up migrate-down migrate-create swag swag-install mocks mockery-install

include .env
export

GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSLMODE)
MIGRATIONS_DIR ?= ./migrations

help:
	@echo "Available targets:"
	@echo "  build           - build the Go binary"
	@echo "  run             - run the Go app locally"
	@echo "  tidy            - go mod tidy"
	@echo "  test            - run unit tests"
	@echo "  docker-up       - start full stack via docker-compose"
	@echo "  docker-down     - stop the stack"
	@echo "  docker-build    - rebuild Docker images"
	@echo "  migrate-up      - apply all pending migrations"
	@echo "  migrate-down    - roll back the last migration"
	@echo "  migrate-create  - scaffold a new migration: make migrate-create name=add_users_table"
	@echo "  swag            - regenerate Swagger docs (docs/) from annotations"
	@echo "  swag-install    - install the swag CLI ($$GOPATH/bin/swag)"
	@echo "  mocks           - regenerate testify mocks from .mockery.yaml"
	@echo "  mockery-install - install the mockery CLI ($$GOPATH/bin/mockery)"

build:
	go build -o bin/app ./cmd/app

run:
	go run ./cmd/app

tidy:
	go mod tidy

test:
	go test ./... -v -race -cover

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

migrate-up:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="host=localhost port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSLMODE)" goose -dir $(MIGRATIONS_DIR) up

migrate-down:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="host=localhost port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSLMODE)" goose -dir $(MIGRATIONS_DIR) down

migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=<migration_name>"; exit 1; fi
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

swag:
	swag init -g cmd/app/main.go --parseInternal -o docs

swag-install:
	go install github.com/swaggo/swag/cmd/swag@latest

mocks:
	mockery

mockery-install:
	go install github.com/vektra/mockery/v2@latest