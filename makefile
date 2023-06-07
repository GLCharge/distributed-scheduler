
API_NAME=manager
RUNNER_NAME=runner

.PHONY: build/api
build/api: docs
	@echo "Building..."
	@go build -o bin/$(API_NAME) cmd/$(API_NAME)/main.go

.PHONY: build/runner
build/runner:
	@echo "Building..."
	@go build -o bin/$(RUNNER_NAME) cmd/$(RUNNER_NAME)/main.go

.PHONY: run/api
run/api: dev/up
	@echo "Running..."
	@bin/$(API_NAME)

.PHONY: run/runner
run/runner: dev/up
	@echo "Running..."
	@bin/$(RUNNER_NAME) --max-concurrent-jobs=2000


.PHONY: dev/up
dev/up:
	@echo "Starting dev environment..."
	@docker compose -f docker-compose.dev.yml up -d

.PHONY: dev/down
dev/down:
	@echo "Stopping dev environment..."
	@docker compose -f docker-compose.dev.yml down

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v --race ./...

.PHONY: db/migrate
db/migrate: dev/up
	@echo "Running migrations..."
	@go run cmd/tooling/main.go migrate --host=localhost:5436

.PHONY: get/api/flags
get/api/flags:
	@echo "Getting flags..."
	@go run cmd/$(API_NAME)/main.go --help

.PHONY: get/runner/flags
get/runner/flags:
	@echo "Getting flags..."
	@go run cmd/$(RUNNER_NAME)/main.go --help

.PHONY: docs
docs:
	@echo "Generating docs..."
	@swag init -g handlers/doc.go