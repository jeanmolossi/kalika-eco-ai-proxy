-- 202511201905_seed_guardrail_default_rules.down.sql

DELETE FROM apx.guardrail_rules
WHERE tenant_id = apx.default_tenant_id()
  AND description IN (
    'Bloqueia tentativas comuns de prompt injection.',
    'Mascaramento de CPF.',
    'Mascaramento de CNPJ.',
    'Limita tamanho do prompt para reduzir custo e evitar abusos.',
    'Mascaramento de emails.'
  );

