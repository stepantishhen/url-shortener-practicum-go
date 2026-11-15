package storage

import (
	"fmt"
	"testing"
)

func BenchmarkMemoryStorage_Save(b *testing.B) {
	m := NewMemoryStorage()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Save("user1", "https://example.com/page")
	}
}

func BenchmarkMemoryStorage_Get(b *testing.B) {
	m := NewMemoryStorage()
	id, _ := m.Save("user1", "https://example.com/page")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(id)
	}
}

func BenchmarkMemoryStorage_SaveBatch(b *testing.B) {
	m := NewMemoryStorage()
	items := make([]BatchInput, 10)
	for i := 0; i < 10; i++ {
		items[i] = BatchInput{
			CorrelationID: fmt.Sprintf("corr-%d", i),
			OriginalURL:   fmt.Sprintf("https://example.com/%d", i),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SaveBatch("user1", items)
	}
}

func BenchmarkMemoryStorage_GetByUser(b *testing.B) {
	m := NewMemoryStorage()
	for i := 0; i < 100; i++ {
		m.Save("user1", fmt.Sprintf("https://example.com/%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetByUser("user1")
	}
}

func BenchmarkMemoryStorage_DeleteBatch(b *testing.B) {
	m := NewMemoryStorage()
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		id, _ := m.Save("user1", fmt.Sprintf("https://example.com/%d", i))
		ids[i] = id
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteBatch("user1", ids)
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateID()
	}
}
