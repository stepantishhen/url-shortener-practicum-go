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
	UserID      string `json:"user_id,omitempty"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}

type FileStorage struct {
	mu      sync.RWMutex
	data    map[string]string   // short_id -> original_url
	byUser  map[string][]string // user_id  -> []short_id
	owners  map[string]string   // short_id -> user_id
	deleted map[string]bool     // short_id -> is_deleted
	records []record
	path    string
}

func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		data:    make(map[string]string),
		byUser:  make(map[string][]string),
		owners:  make(map[string]string),
		deleted: make(map[string]bool),
		path:    path,
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
		if r.UserID != "" {
			f.byUser[r.UserID] = append(f.byUser[r.UserID], r.ShortURL)
			f.owners[r.ShortURL] = r.UserID
		}
		if r.DeletedFlag {
			f.deleted[r.ShortURL] = true
		}
	}
	return nil
}

func (f *FileStorage) Save(userID, originalURL string) (string, error) {
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
	r := record{UUID: uuid, ShortURL: id, OriginalURL: originalURL, UserID: userID}
	f.records = append(f.records, r)
	f.data[id] = originalURL
	if userID != "" {
		f.byUser[userID] = append(f.byUser[userID], id)
		f.owners[id] = userID
	}
	if err := f.flush(); err != nil {
		f.records = f.records[:len(f.records)-1]
		delete(f.data, id)
		if userID != "" {
			ids := f.byUser[userID]
			f.byUser[userID] = ids[:len(ids)-1]
			delete(f.owners, id)
		}
		return "", err
	}
	return id, nil
}

func (f *FileStorage) Get(id string) (string, bool, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	url, ok := f.data[id]
	if !ok {
		return "", false, false
	}
	return url, true, f.deleted[id]
}

func (f *FileStorage) SaveBatch(userID string, items []BatchInput) ([]BatchOutput, error) {
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
			UserID:      userID,
		}
		f.data[pairs[i].shortID] = item.OriginalURL
		if userID != "" {
			f.byUser[userID] = append(f.byUser[userID], pairs[i].shortID)
			f.owners[pairs[i].shortID] = userID
		}
	}
	f.records = append(f.records, newRecords...)

	if err := f.flush(); err != nil {
		for _, r := range newRecords {
			delete(f.data, r.ShortURL)
			delete(f.owners, r.ShortURL)
		}
		if userID != "" {
			ids := f.byUser[userID]
			f.byUser[userID] = ids[:len(ids)-len(newRecords)]
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

func (f *FileStorage) GetByUser(userID string) ([]UserURL, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	ids := f.byUser[userID]
	result := make([]UserURL, 0, len(ids))
	for _, id := range ids {
		if url, ok := f.data[id]; ok {
			result = append(result, UserURL{ShortID: id, OriginalURL: url})
		}
	}
	return result, nil
}

func (f *FileStorage) DeleteBatch(userID string, ids []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	changed := false
	for i := range f.records {
		r := &f.records[i]
		if r.UserID == userID && contains(ids, r.ShortURL) && !r.DeletedFlag {
			r.DeletedFlag = true
			f.deleted[r.ShortURL] = true
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return f.flush()
}

func (f *FileStorage) PingContext(_ context.Context) error { return nil }

func (f *FileStorage) flush() error {
	data, err := json.Marshal(f.records)
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0644)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
