# Bootstrap Guide

## Pré-requisitos
- Go 1.25+ instalado e presente no `PATH`.
- `golangci-lint` 2.6.2+ instalado (veja `AGENTS.md`).
- Docker + Docker Compose para bancos isolados por módulo.
- `make` disponível para rodar os atalhos do `Makefile`.
- Variáveis de ambiente documentadas em `.env.example` acessíveis.

## Configuração rápida
1. **Instale dependências Go**
   ```bash
   go mod download
   ```
2. **Configure variáveis de ambiente específicas**
   - Copie `.env.example` para `.env` e ajuste DSNs por módulo: `GATEWAY_POSTGRES_DSN`, `TENANT_POSTGRES_DSN`, `GUARDRAIL_POSTGRES_DSN`, `OBSERVABILITY_POSTGRES_DSN`.
   - Defina servidores com `GATEWAY_SERVER_*`, `TENANT_SERVER_*`, `GUARDRAIL_SERVER_*`, `OBSERVABILITY_SERVER_*` (evite reuso de portas ao subir múltiplos serviços).
   - `SERVER_*` também controla `BASE_PATH`, TLS (`SERVER_ENABLE_TLS`, `SERVER_TLS_CERTFILE`, `SERVER_TLS_KEYFILE`) e pprof (`SERVER_ENABLE_PPROF`).
   - Ajuste sinks de uso/auditoria: `USAGE_MODE`/`AUDIT_MODE` (`file` ou `kafka`), `USAGE_TOPIC`/`AUDIT_TOPIC` e `KAFKA_BROKERS` quando Kafka estiver habilitado.
   - `RATELIMIT_*` para rate limiting e `LLM_*` para provedores upstream e modelos permitidos.
3. **Suba dependências locais (opcional)**
   ```bash
   docker compose up -d gateway-db tenant-db guardrail-db observability-db
   ```

## Rodando os serviços
- Use os binários gerados com `make build`:
  ```bash
  ./bin/gateway        # proxy HTTP
  ./bin/tenant         # API e tarefas de tenants/chaves
  ./bin/guardrails     # motor de guardrails
  ./bin/observability  # publishers de usage/audit
  ```
  Ajuste `SERVER_PORT` para rodar múltiplos serviços em paralelo. O `SERVER_BASE_PATH` padrão é `/api/v1` e todas as rotas são registradas relativas a ele.
- Para desenvolvimento com Docker Compose focando no gateway:
  ```bash
  make docker-up
  ```

## Migrations
- Cada módulo mantém migrations via `go:embed` e roda na própria conexão:
  - Gateway: `internal/gateway/infra/migrations` (uso/telemetria).
  - Tenant: `internal/tenant/infra/migrations` (tenants, API keys, policies).
  - Guardrails: `internal/guardrails/infra/migrations` (regras e seeds).
- `core.RunAllMigrations` aplica as migrations ao iniciar a aplicação; não há fallback global de banco.

## Makefiles úteis
- `make build`: compila binários em `./bin`.
- `make lint`: executa `golangci-lint run`.
- `make test`: roda `go test ./...`.
- `make fmt`: aplica formatação.
- `make kafka-topics-init`: cria tópicos obrigatórios do Kafka quando o cluster local está ativo.

## Serviços externos
- **LLM Providers:** configure endpoints/keys em `LLM_*` conforme `pkg/toolkit/config`.
- **Kafka:** tópicos definidos em `docs/KAFKA-TOPICS.md`; habilite `KAFKA_*` apenas se necessário.
- **Observability:** exportadores OTel configurados via `OBSERVE_*` (métricas/traces). Logs de auditoria/uso padrão gravam em `logs/audit-events.log` e `logs/usage-events.log` quando `MODE=file`.

## Dicas úteis
- Evite usar `*` em `SERVER_ALLOWED_ORIGINS` em produção; defina origens explícitas.
- O Docker Compose inclui `kafka-ui` em `http://localhost:8082`; use-o para inspecionar tópicos ao testar publicações Kafka.
- Habilite TLS e pprof apenas quando necessário em ambientes de desenvolvimento controlados.
