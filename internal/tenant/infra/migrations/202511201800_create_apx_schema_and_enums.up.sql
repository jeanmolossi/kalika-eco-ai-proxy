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

COMMIT;
