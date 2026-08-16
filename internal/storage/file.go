package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	mu      sync.RWMutex
	data    map[string]string
	records []record
	path    string
}

func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		data: make(map[string]string),
		path: path,
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStorage) load() error {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || len(data) == 0 {
		return err
	}
	if err := json.Unmarshal(data, &f.records); err != nil {
		return err
	}
	for _, r := range f.records {
		f.data[r.ShortURL] = r.OriginalURL
	}
	return nil
}

func (f *FileStorage) Save(originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	uuid, err := generateID()
	if err != nil {
		return "", err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	r := record{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: originalURL,
	}
	f.records = append(f.records, r)
	f.data[id] = originalURL
	if err := f.flush(); err != nil {
		f.records = f.records[:len(f.records)-1]
		delete(f.data, id)
		return "", err
	}
	return id, nil
}

func (f *FileStorage) Get(id string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	url, ok := f.data[id]
	return url, ok
}

func (f *FileStorage) SaveBatch(items []BatchInput) ([]BatchOutput, error) {
	type idPair struct{ shortID, uuid string }
	pairs := make([]idPair, len(items))
	for i := range items {
		shortID, err := generateID()
		if err != nil {
			return nil, err
		}
		uuid, err := generateID()
		if err != nil {
			return nil, err
		}
		pairs[i] = idPair{shortID, uuid}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	newRecords := make([]record, len(items))
	for i, item := range items {
		newRecords[i] = record{
			UUID:        pairs[i].uuid,
			ShortURL:    pairs[i].shortID,
			OriginalURL: item.OriginalURL,
		}
		f.data[pairs[i].shortID] = item.OriginalURL
	}
	f.records = append(f.records, newRecords...)

	if err := f.flush(); err != nil {
		for _, r := range newRecords {
			delete(f.data, r.ShortURL)
		}
		f.records = f.records[:len(f.records)-len(newRecords)]
		return nil, err
	}

	results := make([]BatchOutput, len(items))
	for i, item := range items {
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: pairs[i].shortID}
	}
	return results, nil
}

func (f *FileStorage) PingContext(_ context.Context) error { return nil }

func (f *FileStorage) flush() error {
	data, err := json.Marshal(f.records)
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0644)
}
