-- 202511201815_create_apx_usage_tables.up.sql

BEGIN;

CREATE SCHEMA IF NOT EXISTS apx;

-- Replica mínima de tenants para manter referencial local no gateway.
-- Alimentada via eventos assíncronos (ex.: Kafka) pelo serviço de tenants.
CREATE TABLE IF NOT EXISTS apx.tenants (
  id UUID PRIMARY KEY
);

-- Garante um tenant padrão para seeds e testes locais.
INSERT INTO apx.tenants (id)
VALUES ('11111111-1111-1111-1111-111111111111'::uuid)
ON CONFLICT (id) DO NOTHING;


CREATE TABLE apx.usage_events (
  id                 BIGSERIAL PRIMARY KEY,
  occurred_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  request_id         UUID NOT NULL,
  tenant_id          UUID NOT NULL REFERENCES apx.tenants(id) ON DELETE CASCADE,
  user_id            TEXT,
  model              TEXT NOT NULL,
  provider           TEXT NOT NULL,           -- "openai", "anthropic", "local"
  prompt_tokens      INT NOT NULL,
  completion_tokens  INT NOT NULL,
  total_tokens       INT NOT NULL,
  cost_usd           NUMERIC(18, 6) NOT NULL,
  source             TEXT,                    -- "http-api", "batch-job", etc.
  metadata           JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_usage_events_tenant_time
  ON apx.usage_events (tenant_id, occurred_at DESC);

CREATE INDEX idx_usage_events_model_time
  ON apx.usage_events (model, occurred_at DESC);

CREATE TABLE apx.usage_daily (
  id                 BIGSERIAL PRIMARY KEY,
  day                DATE NOT NULL,
  tenant_id          UUID NOT NULL REFERENCES apx.tenants(id) ON DELETE CASCADE,
  model              TEXT NOT NULL,
  provider           TEXT NOT NULL,
  total_requests     BIGINT NOT NULL DEFAULT 0,
  prompt_tokens      BIGINT NOT NULL DEFAULT 0,
  completion_tokens  BIGINT NOT NULL DEFAULT 0,
  total_tokens       BIGINT NOT NULL DEFAULT 0,
  cost_usd           NUMERIC(18, 6) NOT NULL DEFAULT 0,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT usage_daily_uq UNIQUE (day, tenant_id, model, provider)
);

CREATE INDEX idx_usage_daily_tenant_day
  ON apx.usage_daily (tenant_id, day);

CREATE INDEX idx_usage_daily_day
  ON apx.usage_daily (day);

END;
