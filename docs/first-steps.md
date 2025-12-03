# First Steps

Este guia apresenta rapidamente o que o produto faz, como o código está organizado e quais convenções seguir ao contribuir.

## O que é o projeto
- **Proxy de IA multitenant**: expõe endpoints HTTP para chat e embeddings, roteando requisições para provedores de LLM e aplicando limites por inquilino, guarda-corpos e auditoria. Agora cada bounded context tem um binário dedicado em `apps/` (`gateway`, `tenant`, `guardrails`, `observability`), todos configuráveis via `SERVER_*` e com wiring de módulos isolados para facilitar a extração como serviços independentes.
- **Pipelines de chat/embeddings**: cada requisição passa por rate limiting por tokens, cache semântico opcional, guarda-corpos, roteamento para o modelo solicitado e publicação de eventos de uso/auditoria para observabilidade e billing.

## Estrutura do código
- `apps/gateway/main.go`: ponto de entrada do gateway HTTP. Carrega configuração, inicializa logger, registra módulos (banco, tenant, guardrails, rate limit/cache, LLM/router/tokenizer, observabilidade, gateway), configura timeouts do servidor HTTP e executa bootstrap e shutdown gracioso.
- `apps/{tenant,guardrails,observability}/main.go`: binários dedicados a cada bounded context. Cada um injeta apenas os módulos necessários (`database` + módulo do domínio) e sobe um servidor HTTP básico (healthcheck) para permitir operação isolada ou em composições de serviços.
- `internal/core/`: abstrações centrais de lifecycle e DI. `App` orquestra registro de dependências, migrações, rotas HTTP e parada ordenada dos módulos. `Registry` resolve e ordena módulos por peso.
- `internal/gateway/`: módulo de domínio que implementa chat e embeddings. O `Service` coordena limitador de tokens, cache, guarda-corpos, roteador de modelos e publicação de eventos de uso/auditoria por inquilino.
- `internal/{tenant,guardrails,ratelimit,cache,llm,observability,database}/`: módulos especializados que expõem o store de tenants, guardrails, limitador/token cache, pool de LLMs+roteador+tokenizer, publishers de observabilidade e conexão com banco. Cada um usa contratos de `pkg/*` e suas respectivas implementações locais.
- `docs/`: guias de bootstrap, roadmap de tarefas e notas de revisão de segurança; este arquivo complementa com visão geral e estilo de código.

## Como começar a mexer
- Instale Go 1.25+ e GolangCI-Lint 2.6.2+ (veja `AGENTS.md`).
- Instale deps e compile:
  ```bash
  go mod download
  make build
  ```
- Escolha o serviço que quer rodar localmente (ajuste `SERVER_PORT` para não conflitar entre eles):
  ```bash
  ./bin/gateway           # proxy HTTP
  ./bin/tenant            # gestão de tenants e chaves
  ./bin/guardrails        # motor de guardrails
  ./bin/observability     # publishers de usage/audit
  ```
  Também é possível usar `make docker-up` para rodar em Docker Compose focando no gateway.
- Rode verificações: `make fmt`, `make lint`, `make test`.

## Configuração de CORS e provedores de LLM
- Defina `SERVER_ALLOWED_ORIGINS` (separado por vírgula) para restringir origens; o padrão evita `*` em produção.
- Configure o provedor upstream padrão via `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_CHAT_MODELS` e `LLM_EMBED_MODELS`. Tenants podem sobrepor via `policy_config.routing` sem depender de clientes stub.

## Eventos de auditoria e uso
- Por padrão, eventos são persistidos em arquivos locais (`logs/audit-events.log` e `logs/usage-events.log`).
- Para publicar em Kafka, defina `USAGE_MODE=kafka` e/ou `AUDIT_MODE=kafka`, informe os tópicos (`USAGE_TOPIC`/`AUDIT_TOPIC`) e os brokers via `KAFKA_BROKERS` (por exemplo, `kafka:9092` no Docker Compose local).

## Code-style guidelines
- **Formatação e lint**: use `gofmt`/`goimports` e garanta que `golangci-lint run` passe; as regras ativas incluem `errcheck`, `staticcheck`, `gosec`, `wsl_v5`, `lll` e outras listadas em `.golangci.yml`.
- **Contexto sempre primeiro**: funções públicas e handlers recebem `context.Context` como primeiro argumento para permitir cancelamento e timeouts herdados do servidor HTTP.
- **Erros bem estruturados**: prefira mensagens contextuais com `fmt.Errorf("contexto: %w", err)` e evite expor detalhes sensíveis em respostas HTTP — logue causas internas e normalize mensagens para clientes.
- **Log estruturado**: use `slog` com chaves consistentes (`module`, `error`, `tenant_id`, etc.) e níveis adequados; evite logs silenciosos ao capturar erros de shutdown.
- **Módulos coesos**: adicione novas features compondo módulos registrados em `core.Registry`, mantendo dependências explícitas via interfaces finas e injeção no `Service` correspondente.
