// Package audit records URL shortening and redirect events for configured listeners.
package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Event represents an audit event.
type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}

// Listener is an observer that receives audit events.
type Listener interface {
	Notify(event Event)
}

// Emitter is the subject that manages observers and dispatches events.
type Emitter struct {
	mu        sync.RWMutex
	listeners []Listener
}

// NewEmitter creates a new audit emitter (subject in Observer pattern).
func NewEmitter() *Emitter {
	return &Emitter{
		listeners: make([]Listener, 0),
	}
}

// AddListener registers a new observer.
func (e *Emitter) AddListener(l Listener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, l)
}

// Emit dispatches an event to all registered listeners.
func (e *Emitter) Emit(event Event) {
	e.mu.RLock()
	listeners := e.listeners
	e.mu.RUnlock()
	for _, l := range listeners {
		l.Notify(event)
	}
}

// Record creates and emits an audit event.
func (e *Emitter) Record(action, userID, url string) {
	e.Emit(NewEvent(action, userID, url))
}

// FileListener writes audit events to a file, one JSON per line.
type FileListener struct {
	file *os.File
	enc  *json.Encoder
	mu   sync.Mutex
}

// NewFileListener creates a new file listener that appends to the given path.
func NewFileListener(path string) (*FileListener, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileListener{file: f, enc: json.NewEncoder(f)}, nil
}

// Notify writes the event as a JSON line to the file.
func (fl *FileListener) Notify(event Event) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if err := fl.enc.Encode(event); err != nil {
		log.Printf("audit: encode event: %v", err)
	}
}

// Close closes the underlying file.
func (fl *FileListener) Close() error {
	if fl.file != nil {
		return fl.file.Close()
	}
	return nil
}

// HTTPListener sends audit events via POST to a remote receiver.
type HTTPListener struct {
	client *http.Client
	url    string
	queue  chan Event
	done   chan struct{}
	mu     sync.Mutex
	closed bool
}

// NewHTTPListener creates a new HTTP listener that POSTs events to the given URL.
func NewHTTPListener(url string) *HTTPListener {
	hl := &HTTPListener{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:   url,
		queue: make(chan Event, 100),
		done:  make(chan struct{}),
	}
	go hl.run()
	return hl
}

// Notify queues an event for delivery by the listener's worker.
func (hl *HTTPListener) Notify(event Event) {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	if hl.closed {
		return
	}
	hl.queue <- event
}

// Close stops accepting events and waits for all queued events to be delivered.
func (hl *HTTPListener) Close() {
	hl.mu.Lock()
	if !hl.closed {
		hl.closed = true
		close(hl.queue)
	}
	hl.mu.Unlock()
	<-hl.done
}

func (hl *HTTPListener) run() {
	defer close(hl.done)
	for event := range hl.queue {
		hl.send(event)
	}
}

func (hl *HTTPListener) send(event Event) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("audit: marshal event: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, hl.url, bytes.NewReader(body))
	if err != nil {
		log.Printf("audit: create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hl.client.Do(req)
	if err != nil {
		log.Printf("audit: send request: %v", err)
		return
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
}

// NewEvent creates an audit event with the current timestamp.
func NewEvent(action, userID, url string) Event {
	return Event{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}
}
