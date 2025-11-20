-- 202511201830_seed_sample_data.up.sql
-- Popula um tenant e uma API key funcional para testes locais.

-- =====================================================================
-- TENANT SAMPLE
-- =====================================================================

INSERT INTO apx.tenants (
  id,
  code,
  name,
  status,
  plan_code,
  max_tokens_month,
  max_requests_minute,
  metadata
)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'dev',
  'Tenant de Desenvolvimento',
  'active',
  'free',
  500000,                 -- 500k tokens/mês
  120,                    -- 120 req/min
  '{"env":"local"}'
)
ON CONFLICT (code) DO NOTHING;

-- =====================================================================
-- API KEY SAMPLE
-- =====================================================================
-- A chave real gerada para testes será:
--   apx_dev_test_1234567890
--
-- Mas o banco armazena apenas:
--   prefix = 'apx_dev_test'
--   secret_hash = sha256('apx_dev_test_1234567890')

INSERT INTO apx.tenant_api_keys (
  id,
  tenant_id,
  name,
  prefix,
  secret_hash,
  last_four,
  status,
  metadata
)
VALUES (
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'dev-key',
  'apx_dev_test',
  digest('apx_dev_test_1234567890', 'sha256'),
  '7890',
  'active',
  '{"seed":true}'
)
ON CONFLICT (prefix) DO NOTHING;

-- =====================================================================
-- POLICY SAMPLE
-- =====================================================================

INSERT INTO apx.tenant_policies (
  id,
  tenant_id,
  version,
  is_active,
  default_model,
  enable_semantic_cache,
  cache_ttl_seconds,
  max_prompt_tokens,
  max_completion_tokens,
  config
)
VALUES (
  '33333333-3333-3333-3333-333333333333',
  '11111111-1111-1111-1111-111111111111',
  1,
  TRUE,
  'gpt-4o-mini',
  TRUE,
  3600,
  4000,
  4000,
  '{
    "models_allowed": ["gpt-4o-mini", "gpt-4o", "claude-3-haiku"],
    "routing": {
      "rules": [
        {
          "name": "cheap-classification",
          "match": {"intent": "classify"},
          "model": "qwen2-7b",
          "max_tokens": 512
        }
      ]
    },
    "guardrails": {
      "mask_pii": true,
      "blocked_terms": ["senha", "cpf", "cartão"],
      "max_chars": 16000
    }
  }'::jsonb
)
ON CONFLICT (tenant_id)
WHERE is_active = TRUE
DO NOTHING;

