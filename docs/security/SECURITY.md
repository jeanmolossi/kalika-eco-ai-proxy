# Segurança

## Práticas Obrigatórias
- **Dados sensíveis:** nunca logar segredos, tokens ou payloads completos. Mascarar IDs e hashes.
- **Autenticação:** requisições ao gateway devem validar API keys via `tenant` store. Sempre usar comparação constante para hashes.
- **Hashing:** API keys armazenadas como `sha256` + prefixo; nunca guardar segredos em texto puro.
- **Criptografia em trânsito:** habilite TLS nos bancos e provedores externos quando disponível.
- **Input Validation:** sanitize entradas de usuário e normalize encoding; use guardrails para PII e prompt injection.
- **Least privilege:** conexões DB com usuário dedicado por módulo, apenas permissões necessárias ao schema.
- **Secrets management:** carregar segredos de env/secret manager, nunca versionar.
- **Rate limit:** obrigatório em rotas públicas para reduzir abuso.

## Resposta a Incidentes
- Logs estruturados com `request_id` e `tenant_id` em todas as rotas.
- Em vazamento de chave: revogar via `tenant.RevokeExpired` e rodar rotina de rotação.
- Registrar incidentes e RCA em `docs/decisions/` quando pertinente.

## Checklist antes de deploy
- Variáveis `*_POSTGRES_DSN` apontando para bancos isolados.
- Guardrails habilitados (`GUARDRAILS_ENABLED=true`) em produção.
- Observabilidade ativa (traces + métricas) para detecção precoce.
- `golangci-lint` sem warnings de segurança (gosec, errcheck, sqlclosecheck).
