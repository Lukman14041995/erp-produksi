// Package migrate runs the embedded SQL migrations against the target
// database on server startup, so a fresh container (or a container behind a
// freshly provisioned database) comes up schema-ready without a separate
// migration step in the deployment pipeline.
package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ranji/clothing-erp/migrations"
)

// Run applies every pending migration. It is safe to call from multiple
// concurrently starting replicas: golang-migrate takes a Postgres advisory
// lock for the duration of the run, so only one replica actually migrates
// while the others block until it finishes, then see ErrNoChange.
func Run(databaseURL string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
