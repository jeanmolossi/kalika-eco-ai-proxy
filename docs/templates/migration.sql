-- <timestamp>_<descricao>.up.sql
BEGIN;

CREATE SCHEMA IF NOT EXISTS <schema>;
-- crie apenas objetos do módulo atual

COMMIT;

-- <timestamp>_<descricao>.down.sql
BEGIN;
DROP TABLE IF EXISTS <schema>.<tabela>;
DROP TYPE IF EXISTS <schema>.<tipo>;
COMMIT;
