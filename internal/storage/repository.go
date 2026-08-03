package storage

// ConflictError is returned by Save when the original URL already exists in the store.
// ShortID contains the already existing short identifier.
type ConflictError struct {
	ShortID string
}

func (e *ConflictError) Error() string { return "url already exists: " + e.ShortID }

type BatchInput struct {
	CorrelationID string
	OriginalURL   string
}

type BatchOutput struct {
	CorrelationID string
	ShortID       string
}

type URLRepository interface {
	Save(originalURL string) (string, error)
	Get(id string) (string, bool)
	SaveBatch(items []BatchInput) ([]BatchOutput, error)
}
