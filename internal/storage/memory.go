package storage

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

// MemoryStorage is an in-memory implementation of URLRepository.
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
	m.data[id] = originalURL
	m.mu.Unlock()
	return id, nil
}

func (m *MemoryStorage) Get(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[id]
	return url, ok
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}
