package storage

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	if err := runMigrations(db); err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (p *PostgresStorage) Save(originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	res, err := p.db.Exec(
		`INSERT INTO urls (id, original_url) VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING`,
		id, originalURL,
	)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var existingID string
		err = p.db.QueryRow(`SELECT id FROM urls WHERE original_url = $1`, originalURL).Scan(&existingID)
		if err != nil {
			return "", err
		}
		return "", &ConflictError{ShortID: existingID}
	}
	return id, nil
}

func (p *PostgresStorage) Get(id string) (string, bool) {
	var originalURL string
	err := p.db.QueryRow(`SELECT original_url FROM urls WHERE id = $1`, id).Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (p *PostgresStorage) SaveBatch(items []BatchInput) ([]BatchOutput, error) {
	ids := make([]string, len(items))
	for i := range items {
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}

	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO urls (id, original_url) VALUES ($1, $2)
		ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING id`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	results := make([]BatchOutput, len(items))
	for i, item := range items {
		var returnedID string
		if err = stmt.QueryRow(ids[i], item.OriginalURL).Scan(&returnedID); err != nil {
			return nil, err
		}
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: returnedID}
	}

	return results, tx.Commit()
}
