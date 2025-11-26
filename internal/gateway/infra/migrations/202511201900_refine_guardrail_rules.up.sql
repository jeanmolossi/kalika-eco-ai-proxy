CREATE OR REPLACE FUNCTION apx.default_tenant_id() RETURNS uuid AS $$
  SELECT id FROM apx.tenants WHERE code = 'dev';
$$ LANGUAGE sql STABLE;

ALTER TABLE apx.guardrail_rules
    ADD COLUMN IF NOT EXISTS description TEXT DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS severity TEXT DEFAULT 'medium',
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Opcionalmente normalizar phase/kind/severity via enums
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'guardrail_severity') THEN
        CREATE TYPE apx.guardrail_severity AS ENUM ('low', 'medium', 'high', 'critical');
    END IF;
END
$$;

ALTER TABLE apx.guardrail_rules
  ALTER COLUMN severity DROP DEFAULT;

ALTER TABLE apx.guardrail_rules
    ALTER COLUMN severity TYPE apx.guardrail_severity
    USING severity::apx.guardrail_severity;

ALTER TABLE apx.guardrail_rules
  ALTER COLUMN severity SET DEFAULT 'medium'::apx.guardrail_severity;

CREATE INDEX IF NOT EXISTS guardrail_rules_tenant_phase_idx
    ON apx.guardrail_rules (tenant_id, name);

CREATE INDEX IF NOT EXISTS guardrail_rules_enabled_idx
    ON apx.guardrail_rules (is_active);

ALTER TYPE apx.guardrail_kind ADD VALUE 'regex_block';
ALTER TYPE apx.guardrail_kind ADD VALUE 'regex_rewrite';
