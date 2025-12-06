# Template - Repository (Go)

```go
type <Entity>Repository interface {
    Find(ctx context.Context, id string) (<Entity>, error)
    Save(ctx context.Context, entity <Entity>) error
}

type pg<Entity>Repository struct {
    db *pgxpool.Pool
}

func NewPG<Entity>Repository(db *pgxpool.Pool) <Entity>Repository {
    return &pg<Entity>Repository{db: db}
}

func (r *pg<Entity>Repository) Find(ctx context.Context, id string) (<Entity>, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    const query = `SELECT ... FROM <schema>.<table> WHERE id=$1`
    row := r.db.QueryRow(ctx, query, id)

    var e <Entity>
    if err := row.Scan(&e.ID, &e.Field); err != nil {
        return <Entity>{}, mapPgError(err)
    }
    return e, nil
}

func (r *pg<Entity>Repository) Save(ctx context.Context, entity <Entity>) error {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    const stmt = `INSERT INTO <schema>.<table>(id, ...) VALUES ($1, ...) ON CONFLICT (id) DO UPDATE ...`
    if _, err := r.db.Exec(ctx, stmt, entity.ID /* ... */); err != nil {
        return fmt.Errorf("save entity: %w", err)
    }
    return nil
}
```
