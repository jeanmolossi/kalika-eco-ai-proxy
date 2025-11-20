.PHONY: help build run test fmt cert docker-dev docker-up docker-down migrate

GOPATH ?= $(which go)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS=":.*?## "} {printf "\t\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	@go build -o bin/jobs cmd/ai-proxy/main.go

test: ## Run all tests
	@go test ./... -cover

fmt: ## Format code
	@go fmt ./...

cert: ## Generate SSL certificate
	@if [ ! -f "key.pem" ]; then \
		go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost;\
	fi

docker-build: ## Build and run Docker compose for development
	@docker compose up --build

docker-up: ## Start Docker compose services
	@docker compose up -d
	@docker compose logs -f api

docker-down: ## Stop docker compose services
	@docker compose down

docker-rebuild-db: ## Stop database removing volumes
	@docker compose down -v postgres-master; \
		docker compose up -d postgres-master

docker-up-watch: ## Start docker compose services and watch core logs
	@make docker-up
	@docker logs -f api -n30

lint: ## Run golangci-lint fixing issues
	@golangci-lint run --fix
