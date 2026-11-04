.DEFAULT_GOAL := help

LOCAL_TEST_DATABASE_URL ?= postgres://wallet:local_wallet_password@localhost:5432/wallet?sslmode=disable
API_URL ?= http://localhost:8080
WALLET_ID ?= 11111111-1111-4111-8111-111111111111
AMOUNT ?= 100

.PHONY: help up down balance deposit withdraw build logs test test-race db-up test-integration test-all benchmark

help:
	@echo "Main:       make up | make balance | make deposit AMOUNT=100 | make withdraw AMOUNT=50 | make down"
	@echo "Auxiliary:  make build | make logs | make test | make test-race | make db-up"
	@echo "            make test-integration | make test-all | make benchmark"

up:
	docker compose up --build

down:
	docker compose down

balance:
	curl -sS "$(API_URL)/api/v1/wallets/$(WALLET_ID)"

deposit:
	curl -sS -X POST "$(API_URL)/api/v1/wallet" -H "Content-Type: application/json" -d '{"walletId":"$(WALLET_ID)","operationType":"DEPOSIT","amount":$(AMOUNT)}'

withdraw:
	curl -sS -X POST "$(API_URL)/api/v1/wallet" -H "Content-Type: application/json" -d '{"walletId":"$(WALLET_ID)","operationType":"WITHDRAW","amount":$(AMOUNT)}'

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
