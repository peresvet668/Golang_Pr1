.DEFAULT_GOAL := help

LOCAL_TEST_DATABASE_URL ?= postgres://wallet:local_wallet_password@localhost:5432/wallet?sslmode=disable

.PHONY: help up down build logs test test-race db-up test-integration test-all benchmark

help:
	@echo "make up                 Build and start the API and PostgreSQL"
	@echo "make down               Stop the Compose services"
	@echo "make test               Run the fast Go test suite"
	@echo "make test-integration   Start PostgreSQL and run the concurrent HTTP test"
	@echo "make test-all           Run fast and PostgreSQL integration tests"
	@echo "make benchmark          Measure concurrent writes per second"

up:
	docker compose up --build

down:
	docker compose down

build:
	docker compose build api

logs:
	docker compose logs --follow api postgres

test:
	go test ./...

test-race:
	go test -race ./...

db-up:
	docker compose up -d --wait postgres

test-integration: export TEST_DATABASE_URL = $(LOCAL_TEST_DATABASE_URL)
test-integration: db-up
	go test -race -count=1 -run '^TestConcurrentHTTPRequestsKeepWalletBalanceConsistent$$' ./internal/repository/postgres

test-all: test test-integration

benchmark: export TEST_DATABASE_URL = $(LOCAL_TEST_DATABASE_URL)
benchmark: db-up
	go test -run '^$$' -bench BenchmarkConcurrentWalletWrites -benchtime=10s ./internal/repository/postgres
