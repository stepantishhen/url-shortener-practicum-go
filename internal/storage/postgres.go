package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

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

// nullableStr converts an empty string to nil so that SQL drivers map it to NULL
// rather than attempting to cast "" to a typed column (e.g. UUID).
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (p *PostgresStorage) Save(userID, originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	res, err := p.db.Exec(
		`INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING`,
		id, originalURL, nullableStr(userID),
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

func (p *PostgresStorage) Get(id string) (string, bool, bool) {
	var originalURL string
	var isDeleted bool
	err := p.db.QueryRow(
		`SELECT original_url, is_deleted FROM urls WHERE id = $1`, id,
	).Scan(&originalURL, &isDeleted)
	if err != nil {
		return "", false, false
	}
	return originalURL, true, isDeleted
}

func (p *PostgresStorage) PingContext(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error) {
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

	insertStmt, err := tx.Prepare(`INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING`)
	if err != nil {
		return nil, err
	}
	defer insertStmt.Close()

	selectStmt, err := tx.Prepare(`SELECT id FROM urls WHERE original_url = $1`)
	if err != nil {
		return nil, err
	}
	defer selectStmt.Close()

	results := make([]BatchOutput, len(items))
	for i, item := range items {
		if _, err = insertStmt.Exec(ids[i], item.OriginalURL, nullableStr(userID)); err != nil {
			return nil, err
		}
		var returnedID string
		if err = selectStmt.QueryRow(item.OriginalURL).Scan(&returnedID); err != nil {
			return nil, err
		}
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: returnedID}
	}

	return results, tx.Commit()
}

func (p *PostgresStorage) GetByUser(userID string) ([]UserURL, error) {
	rows, err := p.db.Query(`SELECT id, original_url FROM urls WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []UserURL
	for rows.Next() {
		var u UserURL
		if err := rows.Scan(&u.ShortID, &u.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}

func (p *PostgresStorage) DeleteBatch(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, 0, len(ids)+1)
	args = append(args, userID)
	ph := make([]string, len(ids))
	for i, id := range ids {
		args = append(args, id)
		ph[i] = fmt.Sprintf("$%d", i+2)
	}
	query := "UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND id IN (" +
		strings.Join(ph, ",") + ")"
	_, err := p.db.Exec(query, args...)
	return err
}
