-- 202511201900_refine_guardrail_rules.down.sql

ALTER TABLE apx.guardrail_rules
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS severity,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

DROP TYPE IF EXISTS apx.guardrail_severity;

