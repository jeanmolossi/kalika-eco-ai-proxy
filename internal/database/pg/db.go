package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
)

type DB struct {
	pool   *pgxpool.Pool
	cfg    config.Postgres
	closed bool
}

func Open(ctx context.Context, cfg config.Postgres) (*DB, error) {
	dsn := cfg.DSN

	if dsn == "" {
		if cfg.SSLMode == "" {
			cfg.SSLMode = "disable"
		}

		if cfg.Port == 0 {
			cfg.Port = 5432
		}

		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Database.Database, cfg.SSLMode)
	}

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.AppName != "" {
		conf.ConnConfig.RuntimeParams["application_name"] = cfg.AppName
	}

	if cfg.MaxConns > 0 {
		conf.MaxConns = cfg.MaxConns
	}

	if cfg.MinConns > 0 {
		conf.MinConns = cfg.MinConns
	}

	if cfg.MaxConnLifetime > 0 {
		conf.MaxConnLifetime = cfg.MaxConnLifetime
	}

	if cfg.MaxConnIdletime > 0 {
		conf.MaxConnIdleTime = cfg.MaxConnIdletime
	}

	if cfg.HealthcheckFreq > 0 {
		conf.HealthCheckPeriod = cfg.HealthcheckFreq
	}

	// Connect timeout
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	cctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(cctx, conf)
	if err != nil {
		return nil, err
	}

	db := &DB{pool: pool, cfg: cfg}
	if err := db.Ping(cctx); err != nil {
		pool.Close()
		return nil, err
	}

	return db, nil
}

func (d *DB) Close() {
	if d.closed {
		return
	}

	d.pool.Close()
	d.closed = true
}

func (d *DB) Pool() *pgxpool.Pool { return d.pool }

func (d *DB) Ping(ctx context.Context) error {
	tmo := d.cfg.ConnectTimeout
	if tmo <= 0 {
		tmo = 2 * time.Second
	}

	cctx, cancel := context.WithTimeout(ctx, tmo)
	defer cancel()

	return d.pool.Ping(cctx)
}

func (d *DB) SQL() *sql.DB {
	return sql.OpenDB(stdlib.GetPoolConnector(d.pool))
}
