package storage

type URLRepository interface {
	Save(originalURL string) (string, error)
	Get(id string) (string, bool)
}
