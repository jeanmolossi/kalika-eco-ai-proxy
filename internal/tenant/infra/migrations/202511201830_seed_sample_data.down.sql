-- 202511201830_seed_sample_data.down.sql
-- Remove o seed criado para ambiente de dev.

DELETE FROM apx.tenant_policies
WHERE id = '33333333-3333-3333-3333-333333333333';

DELETE FROM apx.tenant_api_keys
WHERE id = '22222222-2222-2222-2222-222222222222';

DELETE FROM apx.tenants
WHERE id = '11111111-1111-1111-1111-111111111111';
