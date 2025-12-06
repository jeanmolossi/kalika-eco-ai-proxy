package core

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
)

type UnwrapConn interface {
	SQL() *sql.DB
}

type migrationData struct {
	Up   []byte
	Down []byte
}

type memorySource struct {
	migrations *source.Migrations
	files      map[uint]migrationData
}

type MigrationFile struct {
	// Full name of the file (e.g., 202511201830_init_ai_proxy.up.sql)
	// We will use the numeric/timestamp prefix to sort.
	Name string

	// SQL content of the migration UP.
	Up string

	// SQL content of the migration DOWN (can be empty if you don't want to provide rollback for this MVP)
	Down string

	// (Optional) Reference of the owning module, just for logging.
	Module string
}

func RunAllMigrations(ctx context.Context, db *sql.DB, modules []Module) error {
	var all []MigrationFile

	for _, m := range modules {
		migs, err := m.Migrations(ctx, nil)
		if err != nil {
			return err
		}

		if len(migs) == 0 {
			continue
		}

		all = append(all, migs...)
	}

	if len(all) == 0 {
		return nil
	}

	src, err := newMemorySource(all)
	if err != nil {
		return fmt.Errorf("build memory source: %w", err)
	}

	// Create source in memory for go-migrate

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("mem", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

func newMemorySource(files []MigrationFile) (source.Driver, error) {
	migs := source.NewMigrations()
	data := make(map[uint]migrationData)

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	for _, f := range files {
		v, err := parseVersion(f.Name)
		if err != nil {
			return nil, err
		}

		if f.Up != "" {
			if ok := migs.Append(&source.Migration{
				Version:    v,
				Identifier: f.Name,
				Direction:  source.Up,
			}); !ok {
				return nil, fmt.Errorf("duplicate migration version %d (up)", v)
			}
		}

		if f.Down != "" {
			if ok := migs.Append(&source.Migration{
				Version:    v,
				Identifier: f.Name,
				Direction:  source.Down,
			}); !ok {
				return nil, fmt.Errorf("duplicate migration version %d (down)", v)
			}
		}

		data[v] = migrationData{
			Up:   []byte(f.Up),
			Down: []byte(f.Down),
		}
	}

	return &memorySource{
		migrations: migs,
		files:      data,
	}, nil
}

func parseVersion(name string) (uint, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid migration name: %s", name)
	}

	n, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version in name %s: %w", name, err)
	}

	return uint(n), nil
}

// ---------- implementação de source.Driver ----------

func (s *memorySource) Open(url string) (source.Driver, error) {
	// Não vamos abrir por URL, só via NewWithInstance.
	return nil, fmt.Errorf("memorySource does not support Open(url)")
}

func (s *memorySource) Close() error {
	return nil
}

func (s *memorySource) First() (uint, error) {
	v, ok := s.migrations.First()
	if !ok {
		return 0, fs.ErrNotExist
	}

	return v, nil
}

func (s *memorySource) Prev(version uint) (uint, error) {
	v, ok := s.migrations.Prev(version)
	if !ok {
		return 0, fs.ErrNotExist
	}

	return v, nil
}

func (s *memorySource) Next(version uint) (uint, error) {
	v, ok := s.migrations.Next(version)
	if !ok {
		return 0, fs.ErrNotExist
	}

	return v, nil
}

func (s *memorySource) ReadUp(version uint) (io.ReadCloser, string, error) {
	m, ok := s.migrations.Up(version)
	if !ok {
		return nil, "", fs.ErrNotExist
	}

	d, ok := s.files[version]
	if !ok || len(d.Up) == 0 {
		return nil, "", fs.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(d.Up)), m.Identifier, nil
}

func (s *memorySource) ReadDown(version uint) (io.ReadCloser, string, error) {
	m, ok := s.migrations.Down(version)
	if !ok {
		return nil, "", fs.ErrNotExist
	}

	d, ok := s.files[version]
	if !ok || len(d.Down) == 0 {
		return nil, "", fs.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(d.Down)), m.Identifier, nil
}
