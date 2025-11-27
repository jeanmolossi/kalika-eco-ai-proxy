.PHONY: help build run test fmt cert docker-dev docker-up docker-down migrate kafka-topics-init

GOPATH ?= $(which go)

SERVICES := gateway tenant guardrails observability

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS=":.*?## "} {printf "\t\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build all service binaries
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		echo "building $$svc"; \
		go build -o bin/$$svc apps/$$svc/main.go; \
	done

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
@docker compose logs -f gateway

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

kafka-topics-init: ## Create required Kafka topics in the local cluster
	@docker compose exec kafka sh -c '\
	  kafka-topics --bootstrap-server localhost:9092 \
	    --create --if-not-exists --topic ai-proxy.audit.events \
	    --partitions 3 --replication-factor 1 \
	    --config cleanup.policy=delete --config retention.ms=604800000 && \
	  kafka-topics --bootstrap-server localhost:9092 \
	    --create --if-not-exists --topic ai-proxy.usage.events \
	    --partitions 3 --replication-factor 1 \
	    --config cleanup.policy=compact,delete --config retention.ms=1209600000 && \
	  kafka-topics --bootstrap-server localhost:9092 \
	              --create --if-not-exists --topic ai-proxy.guardrails.verdicts \
	              --partitions 1 --replication-factor 1 \
	              --config cleanup.policy=delete --config retention.ms=1209600000 \
'
