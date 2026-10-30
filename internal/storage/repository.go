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

// BatchInput carries one item for a batch-shorten request.
type BatchInput struct {
	CorrelationID string
	OriginalURL   string
}

// BatchOutput carries the result for one item in a batch-shorten response.
type BatchOutput struct {
	CorrelationID string
	ShortID       string
}

// UserURL is a short-ID / original-URL pair owned by a single user.
type UserURL struct {
	ShortID     string
	OriginalURL string
}

// URLRepository is the primary storage abstraction for the URL shortener.
// All methods are safe for concurrent use.
type URLRepository interface {
	// Get looks up a short ID and returns (originalURL, found, deleted).
	Get(id string) (string, bool, bool)
	// Save creates a new short ID for originalURL on behalf of userID.
	// Returns ConflictError when the URL already exists.
	Save(userID, originalURL string) (string, error)
	// SaveBatch creates short IDs for multiple URLs in a single operation.
	SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error)
	// GetByUser returns all URLs saved by the given user.
	GetByUser(userID string) ([]UserURL, error)
	// DeleteBatch soft-deletes the given short IDs owned by userID.
	DeleteBatch(userID string, ids []string) error
}

// Pinger checks that the underlying storage is reachable.
type Pinger interface {
	PingContext(ctx context.Context) error
}
