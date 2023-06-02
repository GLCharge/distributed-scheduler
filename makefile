
API_NAME=manager
RUNNER_NAME=runner

build/api:
	@echo "Building..."
	@go build -o bin/$(API_NAME) cmd/$(API_NAME)/main.go

build/runner:
	@echo "Building..."
	@go build -o bin/$(RUNNER_NAME) cmd/$(RUNNER_NAME)/main.go

run/api: dev/up
	@echo "Running..."
	@bin/$(API_NAME)

run/runner: dev/up
	@echo "Running..."
	@bin/$(RUNNER_NAME)


dev/up:
	@echo "Starting dev environment..."
	@docker compose -f docker-compose.dev.yml up -d

dev/down:
	@echo "Stopping dev environment..."
	@docker compose -f docker-compose.dev.yml down

test:
	@echo "Running tests..."
	@go test -v --race ./...


db/migrate: dev/up
	@echo "Running migrations..."
	@go run cmd/tooling/main.go migrate --host=localhost:5436

get/api/flags:
	@echo "Getting flags..."
	@go run cmd/$(API_NAME)/main.go --help

get/runner/flags:
	@echo "Getting flags..."
	@go run cmd/$(RUNNER_NAME)/main.go --help

.PHONY: build/api build/runner run/api run/runner dev/up dev/down test db/migrate get/api/flags get/runner/flags