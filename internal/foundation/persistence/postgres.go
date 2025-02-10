package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

const timeout = 3 * time.Second

type Database struct {
	DB *sql.DB
}

func NewDatabase(ctx context.Context, dataSourceName string) (*Database, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) IsOK(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := d.DB.PingContext(ctx)
	return err == nil
}

func (d *Database) Migrate(dbURL string) error {
	m, err := migrate.New("file:///migrations", dbURL)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			m, err = migrate.New("file://internal/business/domain/message/stores/db/postgres/migrations", dbURL)
			if err != nil {
				return fmt.Errorf("new migrate: %w", err)
			}
		}
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up: %w", err)
	}

	return nil
}
