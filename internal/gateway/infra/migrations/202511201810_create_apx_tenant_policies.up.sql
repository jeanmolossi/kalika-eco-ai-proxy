-- 202511201810_create_apx_tenant_policies.up.sql

CREATE TABLE apx.tenant_policies (
  id                     UUID PRIMARY KEY DEFAULT uuidv7(),
  tenant_id              UUID NOT NULL REFERENCES apx.tenants(id) ON DELETE CASCADE,
  version                INT NOT NULL DEFAULT 1,
  is_active              BOOLEAN NOT NULL DEFAULT TRUE,

  -- campos mais acessados
  default_model          TEXT NOT NULL DEFAULT 'gpt-4o-mini',
  enable_semantic_cache  BOOLEAN NOT NULL DEFAULT TRUE,
  cache_ttl_seconds      INT NOT NULL DEFAULT 3600,
  max_prompt_tokens      INT NOT NULL DEFAULT 4000,
  max_completion_tokens  INT NOT NULL DEFAULT 4000,

  -- JSONB com configs detalhadas (routing, guardrails, etc.)
  config                 JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- garante uma policy ativa por tenant
CREATE UNIQUE INDEX uq_tenant_policies_active
  ON apx.tenant_policies (tenant_id)
  WHERE is_active = TRUE;

CREATE INDEX idx_tenant_policies_tenant_id
  ON apx.tenant_policies (tenant_id);

CREATE INDEX idx_tenant_policies_created_at
  ON apx.tenant_policies (created_at);

