SHELL := /bin/bash

TEST_DB_CONTAINER := room-booking-test-db
TEST_DB_IMAGE := postgres:17
TEST_DB_NAME := room_booking_test
TEST_DB_USER := postgres
TEST_DB_PASSWORD := postgres
TEST_DB_PORT := 55432
TEST_DSN := postgres://$(TEST_DB_USER):$(TEST_DB_PASSWORD)@localhost:$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=disable

PERF_DB_CONTAINER := room-booking-perf-db
PERF_DB_IMAGE := postgres:17
PERF_DB_NAME := room_booking_perf
PERF_DB_USER := postgres
PERF_DB_PASSWORD := postgres
PERF_DB_PORT := 55433
PERF_DSN := postgres://$(PERF_DB_USER):$(PERF_DB_PASSWORD)@localhost:$(PERF_DB_PORT)/$(PERF_DB_NAME)?sslmode=disable
PERF_APP_PORT := 8081
PERF_APP_PID_FILE := .tmp/perf-app.pid
PERF_APP_LOG := .tmp/perf-app.log
LOAD_VUS ?= 50
LOAD_DURATION ?= 30s

.PHONY: up seed down down-v test-db-up test-db-down test perf-db-up perf-db-down perf-seed perf-app-up perf-app-down perf-up perf-down load-slots perf-test-slots

up:
	set -a; source .env; set +a; docker compose up --build

seed:
	set -a; source .env; set +a; go run ./cmd/seed

down:
	docker compose down

down-v:
	docker compose down -v

test-db-up:
	-docker rm -f $(TEST_DB_CONTAINER)
	docker run -d \
		--name $(TEST_DB_CONTAINER) \
		-e POSTGRES_DB=$(TEST_DB_NAME) \
		-e POSTGRES_USER=$(TEST_DB_USER) \
		-e POSTGRES_PASSWORD=$(TEST_DB_PASSWORD) \
		-p $(TEST_DB_PORT):5432 \
		$(TEST_DB_IMAGE)
	until docker exec $(TEST_DB_CONTAINER) pg_isready -U "$(TEST_DB_USER)" -d "$(TEST_DB_NAME)" >/dev/null 2>&1; do sleep 1; done

test-db-down:
	-docker rm -f $(TEST_DB_CONTAINER)

test:
	@trap '$(MAKE) test-db-down >/dev/null' EXIT; \
	$(MAKE) test-db-up; \
	APP_POSTGRES_TEST_DSN="$(TEST_DSN)" go test -count=1 -coverpkg=./... -coverprofile=coverage.out ./...; \
	go tool cover -func=coverage.out | tail -n 1

perf-db-up:
	-docker rm -f $(PERF_DB_CONTAINER)
	docker run -d \
		--name $(PERF_DB_CONTAINER) \
		-e POSTGRES_DB=$(PERF_DB_NAME) \
		-e POSTGRES_USER=$(PERF_DB_USER) \
		-e POSTGRES_PASSWORD=$(PERF_DB_PASSWORD) \
		-p $(PERF_DB_PORT):5432 \
		$(PERF_DB_IMAGE)
	until docker exec $(PERF_DB_CONTAINER) pg_isready -U "$(PERF_DB_USER)" -d "$(PERF_DB_NAME)" >/dev/null 2>&1; do sleep 1; done
	cat migrations/000001_init.up.sql | docker exec -i $(PERF_DB_CONTAINER) psql -U "$(PERF_DB_USER)" -d "$(PERF_DB_NAME)"

perf-db-down:
	-docker rm -f $(PERF_DB_CONTAINER)

perf-seed:
	APP_POSTGRES_DSN="$(PERF_DSN)" APP_JWT_SECRET=perf-secret APP_JWT_TTL_MINUTES=60 go run ./cmd/seed

perf-app-up:
	mkdir -p .tmp
	@if [ -f $(PERF_APP_PID_FILE) ]; then kill "$$(cat $(PERF_APP_PID_FILE))" 2>/dev/null || true; rm -f $(PERF_APP_PID_FILE); fi
	APP_HTTP_PORT=$(PERF_APP_PORT) APP_POSTGRES_DSN="$(PERF_DSN)" APP_JWT_SECRET=perf-secret APP_JWT_TTL_MINUTES=60 nohup go run ./cmd/server > $(PERF_APP_LOG) 2>&1 & echo $$! > $(PERF_APP_PID_FILE)
	until curl -fsS http://localhost:$(PERF_APP_PORT)/_info >/dev/null 2>&1; do sleep 1; done

perf-app-down:
	@if [ -f $(PERF_APP_PID_FILE) ]; then kill "$$(cat $(PERF_APP_PID_FILE))" 2>/dev/null || true; rm -f $(PERF_APP_PID_FILE); fi

perf-up: perf-db-up perf-seed perf-app-up

perf-down: perf-app-down perf-db-down

load-slots:
	@command -v k6 >/dev/null 2>&1 || { echo "k6 is not installed on host. Install it first, for example: sudo snap install k6"; exit 1; }
	BASE_URL=http://localhost:$(PERF_APP_PORT) VUS=$(LOAD_VUS) DURATION=$(LOAD_DURATION) k6 run load/k6_slots.js

perf-test-slots:
	@trap '$(MAKE) perf-down >/dev/null' EXIT; \
	$(MAKE) perf-up; \
	$(MAKE) load-slots
