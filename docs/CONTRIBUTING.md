# Contribuindo

## Pull Requests
- Sempre abra PR pequeno e focado. Descreva motivo, impacto e testes rodados.
- Use commits semânticos: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
- Nunca force-push para `main`. Branches de feature devem ter nome `feat/<area>-<descricao>`.

## Onde colocar cada tipo de arquivo
- **Domínio:** `internal/<modulo>` com entidades, portas e casos de uso.
- **Infra local (DB, cache):** `internal/<modulo>/infra`.
- **Integrações remotas:** `internal/<modulo>/remote`.
- **HTTP:** `internal/<modulo>/infra/http` para handlers e DTOs.
- **Migrations:** `internal/<modulo>/infra/migrations/*.sql`, com embed pelo módulo.
- **Templates e ADRs:** `docs/templates/`, `docs/decisions/`.

## Padrões de código
- Siga `docs/styleguide/GO.md`, `docs/styleguide/SQL.md` e `docs/styleguide/typescript.md`.
- Context obrigatório em qualquer operação I/O. Timeout vindo da config.
- Erros sempre com wrapping (`fmt.Errorf("action: %w", err)`).
- Nenhum handler retorna panics; valide inputs e retorne erros estruturados.

## Estrutura de pastas
- Consulte `docs/FOLDERS-GUIDE.md` para criar novos diretórios.
- Não crie pacotes públicos em `pkg/` sem discussão prévia.

## Regras de teste
- Para código Go, mínimo: testes de unidade para serviços e adaptadores críticos.
- Use `make test` ou `go test ./...` localmente; PRs precisam reportar comandos executados.
- Para migrations, use database local isolado do módulo (sem cross-module FKs).

## Revisão e qualidade
- Execute `golangci-lint run` antes de enviar PR.
- Não deixe `TODO` sem contexto ou issue link.
- Qualquer mudança de config/env requer atualização do `.env.example` e documentação em `docs/BOOTSTRAP-GUIDE.md`.

## Processo de release
- Versões são marcadas por tag após merge. Nenhum build manual em produção sem CI.
- Alterações breaking exigem ADR e migração de dados separada.
