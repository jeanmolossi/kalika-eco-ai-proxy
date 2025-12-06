# ADR 002: OpenTelemetry para traces/métricas

## Contexto
Precisamos de rastreamento distribuído consistente entre gateway, guardrails e tenant.

## Decisão
- Adotar OTel SDK com propagação W3C em todas as entradas HTTP.
- Métricas e traces exportados para backend configurável via env (`OBSERVE_*`).
- Spans nomeados `<module>.<operation>` com atributos de domínio (tenant_id, provider, model).

## Consequências
- Requer instrumentar rotas e clientes (LLM, DB, Kafka).
- Alertas podem ser configurados sobre métricas expostas.
- Custo de observabilidade deve ser monitorado; amostragem configurável por ambiente.
