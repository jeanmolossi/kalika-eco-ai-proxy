# Observabilidade

## Logs
- Use `slog` com chaves estruturadas: `request_id`, `tenant_id`, `route`, `status`, `latency_ms`.
- Nível padrão: `INFO`; use `DEBUG` somente em ambiente local.
- Nunca logar payload completo; preferir hashes ou tamanho.

## Métricas
- Exportar via OTel: contadores para requisições, histogramas para latência e tamanho de payload.
- Prefixos por módulo: `gateway_http_requests_total`, `tenant_api_key_revocations_total`, `guardrails_rules_loaded`.
- Sempre rotular com `module`, `route`, `status_code`, `provider` quando aplicável.

## Traces
- Traços distribuídos com propagação W3C. Iniciar span na borda HTTP; propagar contexto para chamadas LLM e DB.
- Nome de span: `<module>.<operation>` (ex.: `gateway.proxy_chat`, `tenant.find_by_api_key`).
- Adicione atributos: `db.system=postgresql`, `db.statement` somente resumido, nunca com valores sensíveis.

## Erros
- Registre exceções nos spans (`recordError`), marque status `Error` e inclua `request_id`.
- Alertas críticos: timeouts para LLM, falha de conexão DB, queda de publish Kafka.

## Correlation
- `request_id` obrigatório em todas as rotas; gere se não vier do cliente.
- `tenant_id` deve fluir em contexto para métricas/traces.
