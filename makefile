
APP_NAME=scheduler

build:
	@echo "Building..."
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

run: dev-up
	@echo "Running..."
	@bin/$(APP_NAME)

dev-up:
	@echo "Starting dev environment..."
	@docker compose -f docker-compose.dev.yml up -d

dev-down:
	@echo "Stopping dev environment..."
	@docker compose -f docker-compose.dev.yml down

test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: build run dev-up dev-down