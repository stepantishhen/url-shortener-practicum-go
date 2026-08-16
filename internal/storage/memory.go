package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type MemoryStorage struct {
	mu     sync.RWMutex
	data   map[string]string   // short_id -> original_url
	byUser map[string][]string // user_id  -> []short_id
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:   make(map[string]string),
		byUser: make(map[string][]string),
	}
}

func (m *MemoryStorage) Save(userID, originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = originalURL
	if userID != "" {
		m.byUser[userID] = append(m.byUser[userID], id)
	}
	return id, nil
}

func (m *MemoryStorage) Get(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[id]
	return url, ok
}

func (m *MemoryStorage) SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error) {
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
		if userID != "" {
			m.byUser[userID] = append(m.byUser[userID], ids[i])
		}
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: ids[i]}
	}
	return results, nil
}

func (m *MemoryStorage) GetByUser(userID string) ([]UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := m.byUser[userID]
	result := make([]UserURL, 0, len(ids))
	for _, id := range ids {
		if url, ok := m.data[id]; ok {
			result = append(result, UserURL{ShortID: id, OriginalURL: url})
		}
	}
	return result, nil
}

func (m *MemoryStorage) PingContext(_ context.Context) error { return nil }

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}
