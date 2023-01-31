package postgresql

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

func Execute(db *sql.DB, dbName string) error {
	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("sourceInstance error: %v", err)
	}

	targetInstance, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return fmt.Errorf("targetInstance error: %v", err)
	}

	m, err := migrate.NewWithInstance("migrations", sourceInstance, dbName, targetInstance)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance error: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to latest version: %v", err)
	}

	return sourceInstance.Close()
}
