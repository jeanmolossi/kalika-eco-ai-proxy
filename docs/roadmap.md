# Roadmap

- [x] Adicionar timeouts explícitos nas consultas de tenant (FindByAPIKey) para evitar espera indefinida em conexões lentas.
- [x] Tornar a lista de origens permitidas configurável em tempo de execução e documentar um padrão seguro (evitar wildcard em produção).
- [x] Implementar testes de integração para os endpoints de chat e embeddings cobrindo guardrails e limites de rate limit.
- [x] Automatizar rotação e invalidação de chaves de API expiradas para reduzir superfícies de ataque.
- [x] Substituir o `StubClient` por clientes de LLM configuráveis por tenant (chat e embedding) com retries, métricas e suporte a streaming.
- [x] Remover fallbacks `stub-model`/`stub-embed-model` e validar modelos contra allowlist por tenant antes de rotear.
- [x] Implementar cálculo real de custo no publisher de uso, com tabela de preços por modelo e propagação de RequestID/TraceID.
- [x] Trocar os publishers de log por filas ou armazenamento persistente para auditoria e billing (ex.: Kafka, Postgres).
