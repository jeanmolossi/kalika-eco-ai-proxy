# Template - Use Case (Go)

```go
package <module>

type <UseCaseName> struct {
    deps Dependencies // defina struct com ports necessárias
}

func New<UseCaseName>(deps Dependencies) *<UseCaseName> {
    return &<UseCaseName>{deps: deps}
}

func (uc *<UseCaseName>) Execute(ctx context.Context, input <InputDTO>) (<OutputDTO>, error) {
    ctx, cancel := context.WithTimeout(ctx, uc.deps.Config.Timeouts.UseCase)
    defer cancel()

    // 1) validar input
    if err := input.Validate(); err != nil {
        return <OutputDTO>{}, fmt.Errorf("validate input: %w", err)
    }

    // 2) orquestrar portas
    entity, err := uc.deps.Store.Find(ctx, input.ID)
    if err != nil {
        return <OutputDTO>{}, fmt.Errorf("find entity: %w", err)
    }

    // 3) aplicar regra
    entity.Apply(input)

    // 4) persistir/emitir eventos
    if err := uc.deps.Store.Save(ctx, entity); err != nil {
        return <OutputDTO>{}, fmt.Errorf("save entity: %w", err)
    }

    return MapEntityToDTO(entity), nil
}
```
