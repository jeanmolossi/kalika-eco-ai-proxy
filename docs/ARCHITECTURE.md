# Arquitetura

## Visão Macro
- **Estilo:** DDD + Ports & Adapters (Hexagonal). Domínio fica em `internal/<modulo>` sem dependências de frameworks.
- **Boundaries:** Cada módulo (`gateway`, `tenant`, `guardrails`, `observability`, `llm`, `ratelimit`, `cache`, etc.) tem infraestrutura própria, banco próprio e contratos claros.
- **Fluxo de request (happy path Gateway):**
  1. `apps/gateway` recebe HTTP -> roteia para handlers em `internal/gateway/infra/http`.
  2. Handler valida input, chama serviços de domínio em `internal/gateway` (camada de application/use case).
  3. Serviços dependem de portas (interfaces) para cache, rate-limit, roteador LLM, guardrails; implementações ficam em `internal/<module>/infra` ou `remote`.
  4. Resposta serializada com DTO dedicado; não reutilize entidades do domínio para payload.
- **Comunicação entre módulos:** somente via interfaces públicas registradas no container; sem import circular ou acesso direto a detalhes de persistência de outro módulo.

## Camadas
- **Domain (Entidades + Serviços de domínio):** puro Go, sem dependências externas. Regras e invariantes vivem aqui.
- **Application (Use Cases):** orquestram regras, tratam transações, chamam portas e traduzem erros. Mantém DTOs próprios.
- **Ports:** interfaces consumidas/pfornecidas (ex.: `TenantStore`, `RuleRepository`, `Limiter`). Sempre em pacote de domínio ou `internal/<module>/ports`.
- **Adapters:** implementações de portas (Postgres, HTTP clients, cache, mensageria). Ficam em `infra/` ou `remote/` do módulo.

## Padrões e Regras
- **Transversal:** logger estruturado (`slog`), context com deadline em todas as I/O, erros contextualizados com `fmt.Errorf("<context>: %w", err)`.
- **Banco:** cada módulo opera em banco/schema próprio via chave de conexão dedicada. Nenhuma FK entre módulos. Dados compartilhados via sincronização ou eventos, nunca por join cross-module.
- **Config:** nunca sobrescrever config de módulo com valor global. Cada app lê `cfg.<Module>DB` e `cfg.<Module>` para servidor.
- **Migrations:** rodam por módulo, com versão numérica + nome descritivo. Nunca incluir tabela de outro módulo.
- **DTO vs Entidade:** DTO para fronteira (HTTP/Kafka); entidade para regras. Não exponha entidades diretamente.

## Boundaries e Dependências Permitidas
- `gateway` consome portas de `tenant` (somente interface), `guardrails`, `ratelimit`, `cache`, `llm`. Não acessa repositórios diretos de outros módulos.
- `tenant` é fonte de verdade para dados de locatários; outros módulos recebem cópia via APIs/eventos.
- `guardrails` mantém regras e catálogo próprio; apenas aceita `tenant_id` como referência opaca (UUID), sem FK.
- `observability` publica métricas/traces/logs; não lê dados de domínio.
- `database` módulos são infra pura, não podem depender de domínio.

## Exemplos Bons/Ruins
- ✅ Bom: Handler chama use case `ProxyChat` com DTO; use case chama `TenantStore` (interface) e `GuardrailEngine`; implementação Postgres vive em `internal/tenant/infra`.
- ❌ Ruim: Handler importando `internal/tenant/repo_pg.go` direto ou realizando join com tabela de outro módulo.
- ✅ Bom: Migration do `guardrails` cria apenas tabelas de regras e seeds padrão; sem `REFERENCES apx.tenants`.
- ❌ Ruim: Migration do `gateway` criando schema do `tenant`.

## Restrições Arquiteturais
- Sem singletons globais; tudo passa pelo `core.Container` com chave explícita.
- Cada app possui instância única do banco que atende apenas seus módulos.
- Toda chamada externa deve ter timeout configurável; não use clients sem limites.
- Não existe lógica em pacotes `cmd/` ou `apps/`: apenas bootstrap e wiring.
- Qualquer novo módulo precisa de ADR aprovado e documentação em `docs/decisions/` e `docs/MODULES.md`.
