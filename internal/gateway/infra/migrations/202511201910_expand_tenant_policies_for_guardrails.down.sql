-- 202511201910_expand_tenant_policies_for_guardrails.down.sql

ALTER TABLE apx.tenant_policies
    DROP COLUMN IF EXISTS enable_guardrails,
    DROP COLUMN IF EXISTS enable_pii_filtering,
    DROP COLUMN IF EXISTS enable_prompt_injection_protection,
    DROP COLUMN IF EXISTS max_prompt_length;

