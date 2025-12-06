-- 202511201900_refine_guardrail_rules.up.sql

CREATE OR REPLACE FUNCTION apx.default_tenant_id() RETURNS uuid AS $$
  SELECT '11111111-1111-1111-1111-111111111111'::uuid;
$$ LANGUAGE sql IMMUTABLE;

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
