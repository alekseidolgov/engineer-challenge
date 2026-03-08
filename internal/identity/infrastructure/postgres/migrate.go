package postgres

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) (int, error) {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return 0, fmt.Errorf("create migration source: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return 0, fmt.Errorf("create migrator: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return 0, fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	if dirty {
		return int(version), fmt.Errorf("migration version %d is dirty", version)
	}

	return int(version), nil
}
