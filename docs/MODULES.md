# Módulos do Monorepo

## Padrão Organizacional
- Cada módulo expõe uma struct `Deps` em `deps.go` para registrar contratos no `core.Container` e reutilizar wiring em rotas/startup (modelo implementado no `gateway` e replicado em `guardrails`, `tenant` e `observability`).
- Handlers vivem em `infra/http` e só dependem das interfaces públicas da `Deps`, nunca de adapters de outros módulos.
- `module.go` é responsável por criar `Deps`, armazená-la na chave `<modulo>:deps` e registrar rotas/migrations, mantendo a ordem de inicialização via `Weight`.

## Gateway
- **Responsabilidade:** plano de dados HTTP para proxy de LLMs, rate-limit, cache e guardrails.
- **Boundaries:** consome `TenantStore`, `GuardrailEngine`, `Limiter`, `SemanticCache`, `Router`. Não lê banco de outros módulos.
- **Dependências permitidas:** interfaces de `tenant`, `guardrails`, `ratelimit`, `cache`, `llm`. Nunca importar `infra` alheio.
- **Banco:** `database/gateway` com tabelas de uso (`apx.usage_*`).

## Tenant
- **Responsabilidade:** cadastro e políticas de tenants, emissão/rotatividade de API keys.
- **Boundaries:** expõe `Store` para leitura; mutações devem vir por comandos dedicados ou seeds controlados.
- **Dependências permitidas:** somente infra própria e contratos do `core` (logger, config). Não depende de `guardrails`.
- **Banco:** `database/tenant` com `apx.tenants`, `apx.tenant_api_keys`, `apx.tenant_policies`.

## Guardrails
- **Responsabilidade:** regras de segurança/PII, políticas de bloqueio/mascaramento.
- **Boundaries:** expõe `Engine` e `RuleRepository`. Recebe `tenant_id` como UUID opaco; não valida contra banco de tenant.
- **Dependências permitidas:** logger, config e banco próprio.
- **Banco:** `database/guardrail` com `apx.guardrail_rules` e seeds default.

## Observability
- **Responsabilidade:** publishers de métricas, logs e auditoria.
- **Boundaries:** apenas produz eventos; não lê dados de domínio.
- **Dependências permitidas:** conexões Kafka/OTel configuradas por ambiente.
- **Banco:** separado via `database/observability` caso necessário futuramente.

## Infra Transversal
- **Database modules:** um por app, configurado por `cfg.<Module>DB`.
- **Rate limit:** `internal/ratelimit` fornece `Limiter` usado pelo gateway.
- **Cache:** `internal/cache` provê cache semântico.
- **LLM Router:** `internal/llm` decide provedor/modelo com base em políticas.

## Matriz de Dependência Permitida
- Linhas dependem de colunas (✔ permitido):

|       | gateway | tenant | guardrails | observability | infra (db/cache/ratelimit/llm) |
|-------|---------|--------|------------|---------------|---------------------------------|
|gateway|    -    |   ✔    |     ✔      |       ✔       |              ✔                  |
|tenant |    ✖    |   -    |     ✖      |       ✖       |              ✔ (db/logger)      |
|guardrails| ✖    |   ✖    |     -      |       ✖       |              ✔ (db/logger)      |
|observability|✖  |   ✖    |     ✖      |       -       |              ✔                  |
