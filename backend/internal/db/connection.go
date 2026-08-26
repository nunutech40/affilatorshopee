package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(2 * time.Second)
	}
	_ = db.Close()
	return nil, fmt.Errorf("database is not ready")
}

func RunMigrations(databaseURL, migrationsDir string) error {
	dir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return err
	}
	migration, err := migrate.New("file://"+filepath.ToSlash(dir), databaseURL)
	if err != nil {
		return err
	}
	defer func() { _, _ = migration.Close() }()
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
