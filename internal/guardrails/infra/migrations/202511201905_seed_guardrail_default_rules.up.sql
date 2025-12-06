-- Guardrail: Block prompt injection
INSERT INTO apx.guardrail_rules (
  id,
  tenant_id,
  name,
  kind,
  is_active,
  priority,
  config
) VALUES (
  uuidv7(),
  apx.default_tenant_id(),
  'Block common prompt injections',
  'regex_block',
  true,
  10,
  jsonb_build_object(
    'phase', 'input',
    'action', 'block',
    'pattern', '(?i)(ignore all previous instructions|system override|developer mode|jailbreak)',
    'severity', 'high',
    'tags', jsonb_build_array('security', 'prompt_injection')
  )
);

-- Guardrail: mascarar CPF
INSERT INTO apx.guardrail_rules (
  id,
  tenant_id,
  name,
  kind,
  is_active,
  priority,
  config
) VALUES (
  uuidv7(),
  apx.default_tenant_id(),
  'Mask CPF',
  'regex_rewrite',
  true,
  20,
  jsonb_build_object(
    'phase', 'input',
    'action', 'rewrite',
    'pattern', '\b\d{3}\.\d{3}\.\d{3}\-\d{2}\b',
    'replacement', '***-CPF-REDACTED***',
    'severity', 'medium',
    'tags', jsonb_build_array('pii')
  )
);

-- Guardrail: tamanho máximo
INSERT INTO apx.guardrail_rules (
  id,
  tenant_id,
  name,
  kind,
  is_active,
  priority,
  config
) VALUES (
  uuidv7(),
  apx.default_tenant_id(),
  'Limit prompt length',
  'max_length',
  true,
  30,
  jsonb_build_object(
    'phase', 'input',
    'action', 'block',
    'max_length', 8000,
    'severity', 'medium',
    'tags', jsonb_build_array('cost_control')
  )
);

