package storage

import (
	"context"
	"fmt"
)

// ConflictError is returned by Save when the original URL already exists in the store.
// ShortID contains the already existing short identifier.
type ConflictError struct {
	ShortID string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("url already exists: %s", e.ShortID)
}

type BatchInput struct {
	CorrelationID string
	OriginalURL   string
}

type BatchOutput struct {
	CorrelationID string
	ShortID       string
}

type UserURL struct {
	ShortID     string
	OriginalURL string
}

type URLRepository interface {
	// Get returns originalURL, found, deleted.
	Get(id string) (string, bool, bool)
	Save(userID, originalURL string) (string, error)
	SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error)
	GetByUser(userID string) ([]UserURL, error)
	DeleteBatch(userID string, ids []string) error
}

type Pinger interface {
	PingContext(ctx context.Context) error
}
