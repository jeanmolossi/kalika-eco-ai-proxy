-- 202511201820_create_apx_guardrail_rules.up.sql

BEGIN;

CREATE TABLE apx.guardrail_rules (
  id             UUID PRIMARY KEY DEFAULT uuidv7(),
  tenant_id      UUID NOT NULL,
  name           TEXT NOT NULL,
  kind           apx.guardrail_kind NOT NULL,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  priority       INT NOT NULL DEFAULT 100,
  config         JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_guardrail_rules_tenant
  ON apx.guardrail_rules (tenant_id);

CREATE INDEX idx_guardrail_rules_active
  ON apx.guardrail_rules (tenant_id, is_active);

COMMIT;
