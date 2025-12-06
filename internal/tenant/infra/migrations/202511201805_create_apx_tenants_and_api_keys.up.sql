
-- 202511201805_create_apx_tenants_and_api_keys.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS apx.tenants (
  id                  UUID PRIMARY KEY DEFAULT uuidv7(),
  code                TEXT NOT NULL UNIQUE,         -- ex: "empresa_x"
  name                TEXT NOT NULL,                -- nome legível
  status              apx.tenant_status NOT NULL DEFAULT 'trialing',
  plan_code           TEXT NOT NULL DEFAULT 'free', -- ex: "free", "pro", "enterprise"
  max_tokens_month    BIGINT NOT NULL DEFAULT 0,    -- 0 = ilimitado ou só controlado por plano
  max_requests_minute INT NOT NULL DEFAULT 60,
  metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tenants_status
  ON apx.tenants (status);

CREATE INDEX idx_tenants_plan_code
  ON apx.tenants (plan_code);

CREATE TABLE IF NOT EXISTS apx.tenant_api_keys (
  id             UUID PRIMARY KEY DEFAULT uuidv7(),
  tenant_id      UUID NOT NULL REFERENCES apx.tenants(id) ON DELETE CASCADE,
  name           TEXT NOT NULL,             -- ex: "backend-prod"
  prefix         TEXT NOT NULL,             -- ex: "apx_live_abc123"
  secret_hash    BYTEA NOT NULL,            -- hash da key inteira (ex: SHA256)
  last_four      TEXT,                      -- opcional, para exibir na UI
  status         apx.api_key_status NOT NULL DEFAULT 'active',
  expires_at     TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at     TIMESTAMPTZ,
  metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
  CONSTRAINT tenant_api_keys_uq_prefix UNIQUE (prefix)
);

CREATE INDEX idx_tenant_api_keys_tenant_id
  ON apx.tenant_api_keys (tenant_id);

CREATE INDEX idx_tenant_api_keys_status
  ON apx.tenant_api_keys (status);

COMMIT;
