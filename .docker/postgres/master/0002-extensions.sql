-- ============================================
-- 002-extensions.sql
-- ============================================

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pgcrypto;     -- uuid_generate_v7()
CREATE EXTENSION IF NOT EXISTS btree_gin;    -- GIN em tipos básicos
CREATE EXTENSION IF NOT EXISTS btree_gist;   -- se for usar constraints/índices avançados
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Config de busca padrão pt-br (ajuste se quiser multi-idioma por tenant)
-- (Pode deixar por sessão na app;)
SET default_text_search_config = 'pg_catalog.portuguese';


