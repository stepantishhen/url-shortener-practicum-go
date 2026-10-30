package audit

import (
	"testing"
)

func BenchmarkEmitter_Emit_NoListeners(b *testing.B) {
	emitter := NewEmitter()
	event := Now("shorten", "user1", "https://example.com/page")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.Emit(event)
	}
}

func BenchmarkEmitter_Emit_WithListeners(b *testing.B) {
	emitter := NewEmitter()
	for j := 0; j < 3; j++ {
		emitter.AddListener(&mockListener{notify: func(Event) {}})
	}
	event := Now("shorten", "user1", "https://example.com/page")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.Emit(event)
	}
}

func BenchmarkFileListener_Notify(b *testing.B) {
	fl, err := NewFileListener(b.TempDir() + "/audit.log")
	if err != nil {
		b.Fatal(err)
	}
	defer fl.Close()
	event := Now("shorten", "user1", "https://example.com/page")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fl.Notify(event)
	}
}

func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Now("shorten", "user1", "https://example.com/page")
	}
}
