BEGIN;

ALTER TABLE apx.tenant_policies
    ADD COLUMN IF NOT EXISTS enable_guardrails BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS enable_pii_filtering BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS enable_prompt_injection_protection BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS max_prompt_length INT DEFAULT 8000;

END;
