# Guia de Pastas

- `apps/<service>`: bootstrap binário (config, logger, registry). Nenhuma regra de negócio aqui.
- `internal/core`: container DI, ciclo de vida, contratos base (Module, Migration runner).
- `internal/<module>`:
  - `domain` (opcional): entidades e regras. Caso não exista, manter arquivos raiz com nomes claros.
  - `deps.go`: ponto único para montar e registrar dependências do módulo no `core.Container` usando uma struct `Deps` (padrão inaugurado no `gateway` e replicado em `guardrails`, `tenant` e `observability`).
  - `infra/`: adapters locais (DB, cache, http server) e migrations (`infra/migrations`).
  - `infra/http`: handlers, rotas, DTOs de request/response (apenas dependências públicas do módulo).
  - `remote/`: clients para serviços externos (HTTP/Kafka/etc.).
  - `module.go`: implementa `core.Module`, registra `Deps` no container e orquestra rotas/migrations/startup sem acoplar a infra de outros módulos.
- `pkg/`: utilidades compartilhadas estáveis (config, logger, httpx). Não referencie código de domínio daqui.
- `docs/`: decisões e guias obrigatórios; mantenha atualizado.
- `logs/`: pasta de saída local. Nunca commitar arquivos gerados.

## Regras
- Não criar novas pastas top-level sem ADR.
- Nomes de pacotes sempre minúsculos e sem underscores.
- Handlers não podem depender de pastas `infra` de outros módulos.
- Migrations sempre em `internal/<module>/infra/migrations` com embed e chave de conexão específica do módulo.
