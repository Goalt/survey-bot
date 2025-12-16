package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"

	"git.ykonkov.com/ykonkov/survey-bot/internal/config"
)

//go:embed sql/*.sql
var fs embed.FS

func ConnectWithTimeout(timeout time.Duration, cnf config.DatabaseConfig) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbCh := make(chan *sqlx.DB)
	defer close(dbCh)

	errCh := make(chan error)
	defer close(errCh)

	go connect(dbCh, errCh, cnf)

	select {
	case db := <-dbCh:
		d, err := iofs.New(fs, "sql")
		if err != nil {
			return nil, fmt.Errorf("failed to create iofs source: %w", err)
		}

		m, err := migrate.NewWithSourceInstance("iofs", d, cnf.ConnectionString())
		if err != nil {
			return nil, fmt.Errorf("failed to create migrate instance: %w", err)
		}

		if cnf.MigrationsUp {
			err := m.Up()
			switch {
			case errors.Is(err, migrate.ErrNoChange):
				return db, nil
			case err != nil:
				return nil, fmt.Errorf("failed to migrate: %w", err)
			}
		}

		return db, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func connect(dbCh chan *sqlx.DB, errCh chan error, cnf config.DatabaseConfig) {
	db, err := sqlx.Open("postgres", cnf.ConnectionString())
	if err != nil {
		errCh <- err
		return
	}

	err = db.Ping()
	for err != nil {
		log.Print(err)
		time.Sleep(time.Second * 5)
		err = db.Ping()
	}

	dbCh <- db
}
