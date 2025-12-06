# ADR 003: Logs estruturados

## Contexto
Logs precisam ser consumíveis por sistemas de observabilidade e auditoria.

## Decisão
- Formato JSON via `slog` com handler estruturado.
- Campos obrigatórios: `timestamp`, `level`, `msg`, `module`, `request_id`, `tenant_id` (quando houver), `error`.
- Nenhum payload sensível será logado; apenas metadados.

## Consequências
- Facilita correlação com traces (mesmo `request_id`).
- Requer revisão de handlers para garantir que loguem dados mínimos e mascarados.
