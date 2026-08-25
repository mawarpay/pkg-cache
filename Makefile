GOLANGCI_LINT_VERSION ?= v2.13.0
GOLANGCI_LINT_IMAGE = golangci/golangci-lint:$(GOLANGCI_LINT_VERSION)

.PHONY: help fmt fmt-check vet test bench build tidy check lint docker-build docker-test docker-lint docker-up docker-down docker-logs clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-14s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

fmt: ## Format Go sources
	go fmt ./...

fmt-check: ## Fail if Go sources are not gofmt-clean
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "gofmt: run 'make fmt' and commit" && exit 1)

vet: ## Run go vet
	go vet ./...

test: ## Run unit tests with race detector and coverage
	go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

bench: ## Run package benchmarks
	go test ./... -bench=. -benchmem -count=1 -run '^$$'

build: ## Compile all packages
	go build ./...

tidy: ## Sync go.mod / go.sum
	go mod tidy

check: fmt-check vet build test ## fmt-check, vet, build, and unit test (CI)

lint: ## Run golangci-lint in Docker
	@mkdir -p "$(HOME)/.cache/golangci-lint"
	docker run --rm -t \
		-v "$(CURDIR):/app" \
		-w /app \
		--user "$(shell id -u):$(shell id -g)" \
		-e GOPROXY \
		-e GOPRIVATE \
		-e GONOSUMDB \
		-v "$(shell go env GOCACHE):/.cache/go-build" \
		-e GOCACHE=/.cache/go-build \
		-v "$(shell go env GOMODCACHE):/.cache/mod" \
		-e GOMODCACHE=/.cache/mod \
		-v "$(HOME)/.cache/golangci-lint:/.cache/golangci-lint" \
		-e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		$(GOLANGCI_LINT_IMAGE) golangci-lint run

docker-lint: lint ## Alias for lint

docker-build: ## Build the test image
	docker compose build test

docker-test: ## Run go test ./... inside Docker (with Redis)
	docker compose run --rm test

docker-up: ## Start Redis with host port (:6379)
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d redis

docker-down: ## Stop Compose services and remove containers
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down

docker-logs: ## Tail Redis logs
	docker compose logs -f redis

clean: ## Remove local caches and Compose leftovers
	go clean -cache -testcache
	rm -f coverage.out
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down --volumes --remove-orphans
