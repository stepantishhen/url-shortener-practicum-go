package storage

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
