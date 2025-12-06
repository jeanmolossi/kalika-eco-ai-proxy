-- 202511201800_create_apx_guardrail_schema_and_enums.up.sql

BEGIN;

CREATE SCHEMA IF NOT EXISTS apx;

-- Replica mínima de tenants para manter referencial local no módulo de guardrails.
-- Alimentada via eventos assíncronos emitidos pelo serviço de tenants.
CREATE TABLE IF NOT EXISTS apx.tenants (
  id UUID PRIMARY KEY
);

-- Garante um tenant padrão para seeds e testes locais.
INSERT INTO apx.tenants (id)
VALUES ('11111111-1111-1111-1111-111111111111'::uuid)
ON CONFLICT (id) DO NOTHING;

CREATE TYPE apx.guardrail_kind AS ENUM (
  'regex_rewrite',
  'regex_block',
  'max_length',
  'prompt_guard',
  'allowlist',
  'denylist',
  'pii_mask'
);

COMMIT;
