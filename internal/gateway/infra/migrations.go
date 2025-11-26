package infra

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrations(ctx context.Context, m core.Module) ([]core.MigrationFile, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read %s migrations dir: %w", m.Name(), err)
	}

	files := make([]core.MigrationFile, 0, 100)

	type pair struct {
		up   string
		down string
	}

	tmp := make(map[string]*pair)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name() // e.g.: 202511201800_create_apx_schema.up.sql

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", name))
		if err != nil {
			return nil, fmt.Errorf("read %s migration file %s: %w", m.Name(), name, err)
		}

		base := name
		isUp := strings.HasSuffix(name, ".up.sql")
		isDown := strings.HasSuffix(name, ".down.sql")

		key := strings.ReplaceAll(strings.ReplaceAll(base, ".up.sql", ""), ".down.sql", "")

		p, ok := tmp[key]
		if !ok {
			p = &pair{}
			tmp[key] = p
		}

		if isUp {
			p.up = string(content)
		} else if isDown {
			p.down = string(content)
		}
	}

	for key, p := range tmp {
		// key is still something like 202511201800_create_apx_schema
		files = append(files, core.MigrationFile{
			Name:   key, // important: starts by timestamp/sequence global
			Up:     p.up,
			Down:   p.down,
			Module: m.Name(),
		})
	}

	return files, nil
}
