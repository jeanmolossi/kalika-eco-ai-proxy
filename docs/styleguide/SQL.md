# SQL Styleguide

- Cada módulo tem schema e banco próprios. Nunca crie FK para tabelas de outro módulo.
- Nomes em snake_case; prefixe schema (`apx.<tabela>`).
- Migrations usam timestamps + descrição (`202511201815_create_apx_usage_tables`).
- Use `IF NOT EXISTS` para objetos idempotentes; envolva em `BEGIN/COMMIT` quando alterar múltiplos objetos.
- Seeds devem ser opcionais e limpos no down quando fizer sentido.
- Evite funções mágicas sem documentação; prefira `COMMENT ON TABLE/COLUMN` quando útil.
- Índices: sempre nomeados `idx_<tabela>_<colunas>` e apenas quando necessário.
- Tipos enumerados: mantenha em schema do módulo, com down correspondente.
