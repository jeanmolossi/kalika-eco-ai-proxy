
-- 202511201800_create_apx_schema_and_enums.up.sql

BEGIN;

CREATE SCHEMA IF NOT EXISTS apx;

CREATE TYPE apx.tenant_status AS ENUM (
  'trialing',
  'active',
  'suspended',
  'canceled'
);

CREATE TYPE apx.api_key_status AS ENUM (
  'active',
  'revoked'
);

CREATE TYPE apx.guardrail_kind AS ENUM (
  'block_term',
  'mask_pii',
  'max_length',
  'custom_lua',
  'classification_block'
);

COMMIT;
