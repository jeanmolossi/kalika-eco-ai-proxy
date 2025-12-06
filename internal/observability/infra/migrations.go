package infra

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
)

// Migrations returns observability migrations. Currently no migrations are required.
func Migrations(_ context.Context, _ core.Module) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
