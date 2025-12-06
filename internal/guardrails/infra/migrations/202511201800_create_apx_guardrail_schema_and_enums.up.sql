-- 202511201800_create_apx_guardrail_schema_and_enums.up.sql

BEGIN;

CREATE SCHEMA IF NOT EXISTS apx;

CREATE TYPE apx.guardrail_kind AS ENUM (
  'regex_block',
  'prompt_guard',
  'allowlist',
  'denylist',
  'pii_mask'
);

COMMIT;
