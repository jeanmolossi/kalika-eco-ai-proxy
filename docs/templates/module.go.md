# Template - Module (Go)

```go
package <module>

type module struct{}

const ModuleName = "<module>"

func NewModule() core.Module { return &module{} }

func (m *module) Name() string   { return ModuleName }
func (m *module) Weight() int    { return 5 } // ajuste conforme dependências

func (m *module) Provide(ctx context.Context, c *core.Container) error {
    // registre adapters (repos, engines)
    return nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
    return infra.Migrations(ctx, m)
}

func (m *module) MigrationDB(_ context.Context, c *core.Container) (*sql.DB, error) {
    conn := core.MustGet[*pg.DB](c, database.<ConnKey>)
    return conn.SQL(), nil
}

func (m *module) Routes(g *echo.Group, c *core.Container) error {
    // registre handlers
    return nil
}

func (m *module) Start(ctx context.Context, c *core.Container) (func(context.Context) error, error) {
    // background workers opcionais
    return nil, nil
}
```
