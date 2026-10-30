package audit

import (
	"testing"
)

func BenchmarkEmitter_Emit_NoListeners(b *testing.B) {
	emitter := NewEmitter()
	event := Now("shorten", "user1", "https://example.com/page")
	for b.Loop() {
		emitter.Emit(event)
	}
}

func BenchmarkEmitter_Emit_WithListeners(b *testing.B) {
	emitter := NewEmitter()
	for range 3 {
		emitter.AddListener(&mockListener{notify: func(Event) {}})
	}
	event := Now("shorten", "user1", "https://example.com/page")
	for b.Loop() {
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
	for b.Loop() {
		fl.Notify(event)
	}
}

func BenchmarkNow(b *testing.B) {
	for b.Loop() {
		Now("shorten", "user1", "https://example.com/page")
	}
}
