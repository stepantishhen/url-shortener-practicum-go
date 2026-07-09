package storage

// URLRepository describes the contract for storing and retrieving shortened URLs.
type URLRepository interface {
	Save(originalURL string) (string, error)
	Get(id string) (string, bool)
}
