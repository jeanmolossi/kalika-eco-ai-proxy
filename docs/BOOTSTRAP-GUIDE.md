# Bootstrap Guide

## Pré-requisitos
- Go 1.22+ e `golangci-lint` 2.6.2+ instalados.
- Docker + Docker Compose para bancos isolados por módulo.
- Acesso a variáveis de ambiente descritas em `.env.example`.

## Configuração
1. Copie `.env.example` para `.env` e configure DSNs específicos:
   - `GATEWAY_POSTGRES_DSN`, `TENANT_POSTGRES_DSN`, `GUARDRAIL_POSTGRES_DSN`, `OBSERVABILITY_POSTGRES_DSN`.
   - Servers: `GATEWAY_SERVER_*`, `TENANT_SERVER_*`, `GUARDRAIL_SERVER_*`, `OBSERVABILITY_SERVER_*`.
2. Suba dependências locais por módulo (exemplo usando compose):
   ```bash
   docker compose up -d gateway-db tenant-db guardrail-db observability-db
   ```
3. Instale deps Go:
   ```bash
   go mod download
   ```

## Rodando os serviços
- **Gateway:**
  ```bash
  go run ./apps/gateway
  ```
- **Tenant:**
  ```bash
  go run ./apps/tenant
  ```
- **Guardrails:**
  ```bash
  go run ./apps/guardrails
  ```
- **Observability:**
  ```bash
  go run ./apps/observability
  ```

Cada app usa apenas sua conexão de banco configurada; não há fallback global.

## Migrations
- Migrations são carregadas via `go:embed` em cada módulo:
  - Gateway: `internal/gateway/infra/migrations` (uso/telemetria).
  - Tenant: `internal/tenant/infra/migrations` (tenants, API keys, policies).
  - Guardrails: `internal/guardrails/infra/migrations` (regras e seeds).
- Para rodar, basta iniciar a aplicação; `core.RunAllMigrations` aplica por módulo na sua conexão.

## Makefiles úteis
- `make lint`: executa `golangci-lint run`.
- `make test`: roda `go test ./...`.
- `make run-gateway`/`make run-tenant`/... (se disponíveis) fazem bootstrap completo do serviço.

## Serviços externos
- **LLM Providers:** configure endpoints/keys em variáveis `LLM_*` conforme `pkg/toolkit/config`.
- **Kafka:** tópicos definidos em `docs/kafka-topics.md`; habilite `KAFKA_*` apenas se necessário.
- **Observability:** Exportadores OTel configurados via env `OBSERVE_*` (métricas/traces).
