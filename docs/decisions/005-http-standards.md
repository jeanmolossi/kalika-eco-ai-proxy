# ADR 005: Padrões HTTP internos

## Contexto
Precisamos de consistência em rotas e contratos HTTP entre serviços.

## Decisão
- Respostas JSON com envelope simples (quando aplicável) e `Content-Type: application/json`.
- Erros usam `problem+json` com campos `type`, `title`, `detail`, `request_id`.
- Paginação padrão: `page`/`page_size` com limites máximos configuráveis.
- Autenticação via API key em header `Authorization: Bearer <key>` para gateway/tenant; guardrails internos podem usar token de serviço.

## Consequências
- Facilita integração com clientes e métricas (rotas estáveis).
- Requer middlewares compartilhados para request_id e validação de headers.
