package deleter_test

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/deleter"
	"url-shortener-practicum-go/internal/storage"
)

type mockRepo struct {
	mu      sync.Mutex
	calls   map[string][]string
	flushed chan struct{}
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		calls:   make(map[string][]string),
		flushed: make(chan struct{}, 2048),
	}
}

func (m *mockRepo) DeleteBatch(userID string, ids []string) error {
	m.mu.Lock()
	m.calls[userID] = append(m.calls[userID], ids...)
	m.mu.Unlock()
	m.flushed <- struct{}{}
	return nil
}

func (m *mockRepo) get(userID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.calls[userID]))
	copy(cp, m.calls[userID])
	return cp
}

func (m *mockRepo) Save(_, _ string) (string, error)                          { return "", nil }
func (m *mockRepo) Get(_ string) (string, bool, bool)                         { return "", false, false }
func (m *mockRepo) SaveBatch(_ string, _ []storage.BatchInput) ([]storage.BatchOutput, error) {
	return nil, nil
}
func (m *mockRepo) GetByUser(_ string) ([]storage.UserURL, error) { return nil, nil }

func waitFlushed(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for DeleteBatch to be called")
	}
}

func TestDeleter_Submit_SizeBasedFlush(t *testing.T) {
	repo := newMockRepo()
	svc := deleter.New(repo, zap.NewNop())

	const userID = "alice"
	ids := make([]string, 100)
	for i := range ids {
		ids[i] = "id"
	}
	svc.Submit(userID, ids)

	waitFlushed(t, repo.flushed, 2*time.Second)

	got := repo.get(userID)
	if len(got) != 100 {
		t.Errorf("expected 100 ids deleted, got %d", len(got))
	}
}

func TestDeleter_Submit_GroupsByUser(t *testing.T) {
	repo := newMockRepo()
	svc := deleter.New(repo, zap.NewNop())

	for i := 0; i < 50; i++ {
		svc.Submit("alice", []string{"a"})
		svc.Submit("bob", []string{"b"})
	}

	waitFlushed(t, repo.flushed, 2*time.Second)

	gotAlice := repo.get("alice")
	gotBob := repo.get("bob")
	if len(gotAlice) == 0 {
		t.Error("expected alice's IDs to be deleted")
	}
	if len(gotBob) == 0 {
		t.Error("expected bob's IDs to be deleted")
	}
}

func TestDeleter_Shutdown_FlushesRemaining(t *testing.T) {
	repo := newMockRepo()
	svc := deleter.New(repo, zap.NewNop())

	const userID = "alice"
	ids := make([]string, 50)
	for i := range ids {
		ids[i] = "id"
	}
	svc.Submit(userID, ids)

	// Буфер канала 1024 — все 50 поместились, run() может ещё не обработал
	// Shutdown вызывает drainAndFlush, который вычитает остаток
	if err := svc.Shutdown(2 * time.Second); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	got := repo.get(userID)
	if len(got) != 50 {
		t.Errorf("expected 50 ids deleted after shutdown, got %d", len(got))
	}
}

type slowMockRepo struct {
	mu      sync.Mutex
	calls   map[string][]string
	delay   chan struct{}
	flushed chan struct{}
}

func newSlowMockRepo() *slowMockRepo {
	return &slowMockRepo{
		calls:   make(map[string][]string),
		delay:   make(chan struct{}),
		flushed: make(chan struct{}, 2048),
	}
}

func (m *slowMockRepo) DeleteBatch(userID string, ids []string) error {
	m.mu.Lock()
	m.calls[userID] = append(m.calls[userID], ids...)
	m.mu.Unlock()
	m.flushed <- struct{}{}
	<-m.delay // блокируем навсегда, пока не закроют канал
	return nil
}

func (m *slowMockRepo) Save(_, _ string) (string, error)                          { return "", nil }
func (m *slowMockRepo) Get(_ string) (string, bool, bool)                         { return "", false, false }
func (m *slowMockRepo) SaveBatch(_ string, _ []storage.BatchInput) ([]storage.BatchOutput, error) {
	return nil, nil
}
func (m *slowMockRepo) GetByUser(_ string) ([]storage.UserURL, error) { return nil, nil }

func TestDeleter_Shutdown_TimesOut(t *testing.T) {
	repo := newSlowMockRepo()
	svc := deleter.New(repo, zap.NewNop())

	ids := make([]string, 100)
	for i := range ids {
		ids[i] = "id"
	}
	svc.Submit("alice", ids)

	err := svc.Shutdown(100 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	close(repo.delay)
}

func TestDeleter_Submit_Nonblocking_WhenFull(t *testing.T) {
	repo := newMockRepo()
	svc := deleter.New(repo, zap.NewNop())

	// Канал имеет буфер 1024 — Submit должен вернуть мгновенно,
	// т.к. select с default не блокируется при переполнении
	done := make(chan struct{})
	go func() {
		svc.Submit("alice", make([]string, 1100))
		close(done)
	}()

	select {
	case <-done:
		// OK — Submit завершился
	case <-time.After(1 * time.Second):
		t.Fatal("Submit blocked — queue should be non-blocking")
	}

	if err := svc.Shutdown(2 * time.Second); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}
