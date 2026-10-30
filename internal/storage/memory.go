package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
)

// idBufPool reuses 6-byte random buffers to reduce GC pressure in generateID.
var idBufPool = sync.Pool{
	New: func() any {
		var b [6]byte
		return &b
	},
}

// MemoryStorage is an in-memory implementation of URLRepository.
type MemoryStorage struct {
	mu      sync.RWMutex
	data    map[string]string   // short_id -> original_url
	byUser  map[string][]string // user_id  -> []short_id
	owners  map[string]string   // short_id -> user_id
	deleted map[string]bool     // short_id -> is_deleted
}

// NewMemoryStorage creates an empty MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:    make(map[string]string),
		byUser:  make(map[string][]string),
		owners:  make(map[string]string),
		deleted: make(map[string]bool),
	}
}

// Save stores a new URL and returns its generated short ID.
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
		m.owners[id] = userID
	}
	return id, nil
}

// Get returns the original URL for a given short ID.
func (m *MemoryStorage) Get(id string) (string, bool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[id]
	if !ok {
		return "", false, false
	}
	return url, true, m.deleted[id]
}

// SaveBatch stores multiple URLs and returns their short IDs.
// IDs are generated before acquiring the lock to reduce contention.
func (m *MemoryStorage) SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error) {
	results := make([]BatchOutput, len(items))
	for i, item := range items {
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		results[i] = BatchOutput{CorrelationID: item.CorrelationID, ShortID: id}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, item := range items {
		m.data[results[i].ShortID] = item.OriginalURL
		if userID != "" {
			m.byUser[userID] = append(m.byUser[userID], results[i].ShortID)
			m.owners[results[i].ShortID] = userID
		}
	}
	return results, nil
}

// GetByUser returns all URLs saved by the given user.
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

// DeleteBatch soft-deletes the given short IDs owned by userID.
func (m *MemoryStorage) DeleteBatch(userID string, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		if m.owners[id] == userID {
			m.deleted[id] = true
		}
	}
	return nil
}

// PingContext implements storage.Pinger for MemoryStorage (always succeeds).
func (m *MemoryStorage) PingContext(_ context.Context) error { return nil }

// generateID returns a random 8-character URL-safe base64 string.
// The source byte buffer is pooled to reduce GC pressure.
func generateID() (string, error) {
	bp := idBufPool.Get().(*[6]byte)
	_, err := rand.Read(bp[:])
	if err != nil {
		idBufPool.Put(bp)
		return "", err
	}
	result := base64.URLEncoding.EncodeToString(bp[:])
	idBufPool.Put(bp)
	return result, nil
}
