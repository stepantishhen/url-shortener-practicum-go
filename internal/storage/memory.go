package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (m *MemoryStorage) Save(originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = originalURL
	return id, nil
}

func (m *MemoryStorage) Get(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[id]
	return url, ok
}

func (m *MemoryStorage) SaveBatch(items []BatchInput) ([]BatchOutput, error) {
	ids := make([]string, len(items))
	for i := range items {
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	results := make([]BatchOutput, len(items))
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, item := range items {
		m.data[ids[i]] = item.OriginalURL
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: ids[i]}
	}
	return results, nil
}

func (m *MemoryStorage) PingContext(_ context.Context) error { return nil }

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}
