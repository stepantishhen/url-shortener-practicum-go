package storage

import (
	"fmt"
	"testing"
)

func BenchmarkMemoryStorage_Save(b *testing.B) {
	m := NewMemoryStorage()
	for b.Loop() {
		m.Save("user1", "https://example.com/page")
	}
}

func BenchmarkMemoryStorage_Get(b *testing.B) {
	m := NewMemoryStorage()
	id, _ := m.Save("user1", "https://example.com/page")
	for b.Loop() {
		m.Get(id)
	}
}

func BenchmarkMemoryStorage_SaveBatch(b *testing.B) {
	m := NewMemoryStorage()
	items := make([]BatchInput, 10)
	for i := range 10 {
		items[i] = BatchInput{
			CorrelationID: fmt.Sprintf("corr-%d", i),
			OriginalURL:   fmt.Sprintf("https://example.com/%d", i),
		}
	}
	for b.Loop() {
		m.SaveBatch("user1", items)
	}
}

func BenchmarkMemoryStorage_GetByUser(b *testing.B) {
	m := NewMemoryStorage()
	for i := range 100 {
		m.Save("user1", fmt.Sprintf("https://example.com/%d", i))
	}
	for b.Loop() {
		m.GetByUser("user1")
	}
}

func BenchmarkMemoryStorage_DeleteBatch(b *testing.B) {
	m := NewMemoryStorage()
	ids := make([]string, 10)
	for i := range 10 {
		id, _ := m.Save("user1", fmt.Sprintf("https://example.com/%d", i))
		ids[i] = id
	}
	for b.Loop() {
		m.DeleteBatch("user1", ids)
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for b.Loop() {
		generateID()
	}
}
