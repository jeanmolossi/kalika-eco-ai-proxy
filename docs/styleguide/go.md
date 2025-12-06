# Go Styleguide

## Estrutura
- Pacotes pequenos, focados em um motivo. Evite pacotes utilitários gigantes.
- Interfaces no domínio, implementações na `infra` ou `remote`.
- Nomes de arquivos descrevem o papel (`store_pg.go`, `handler_chat.go`).

## Código
- Use `context.Context` como primeiro argumento em métodos que tocam rede/DB.
- Sempre defina timeout usando config (`context.WithTimeout`).
- Erros: `return fmt.Errorf("<ação>: %w", err)` e não compare mensagens, use `errors.Is`.
- Logs com `slog`: `log.InfoContext(ctx, "message", "key", value)`.
- Não use variáveis globais; registre tudo no `core.Container`.
- Não use `panic` fora de bootstrap.

## Concurrency
- Prefira `context` para cancelamento. Channels com buffer explícito.
- Proteja dados compartilhados com mutex ou use tipos imutáveis.
- Sempre trate leaks: `defer cancel()` e `defer rows.Close()`.

## Tests
- Estrutura `Arrange-Act-Assert`. Use tabelas de teste.
- Mocks devem ser mínimos, preferir fakes simples.
- Nome de teste: `Test<Struct>_<Metodo>`.
- Garanta cobrir invariantes e erros de infraestrutura (timeouts, violação de contrato).

## Comentários e Doc
- Comente interfaces públicas e funções exportadas.
- TODOs precisam de owner/issue: `// TODO(@owner): ...`.

## Dependências
- Evite dependências novas sem discussão. Priorize stdlib + libs já usadas (slog, echo, pgx).
- Não envolva imports em blocos `try/catch` (Go não usa) e não silencie erros.
