package deleter

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/storage"
)

type item struct {
	userID string
	id     string
}

type Service struct {
	in     chan item
	repo   storage.URLRepository
	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a Service and starts a background goroutine that processes
// delete requests. Call Shutdown(timeout) to stop it and flush pending items.
func New(repo storage.URLRepository, log *zap.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		in:     make(chan item, 1024),
		repo:   repo,
		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *Service) Submit(userID string, ids []string) {
	for _, id := range ids {
		select {
		case s.in <- item{userID: userID, id: id}:
		default:
			s.log.Warn("delete queue is full, skipping item",
				zap.String("user_id", userID),
				zap.String("id", id),
			)
		}
	}
}

func (s *Service) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var buf []item

	for {
		select {
		case <-s.ctx.Done():
			s.drainAndFlush(buf)
			return

		case it, ok := <-s.in:
			if !ok {
				if len(buf) > 0 {
					s.flush(buf)
				}
				return
			}
			buf = append(buf, it)
			if len(buf) >= 100 {
				s.flush(buf)
				buf = buf[:0]
			}

		case <-ticker.C:
			if len(buf) > 0 {
				s.flush(buf)
				buf = buf[:0]
			}
		}
	}
}

func (s *Service) drainAndFlush(buf []item) {
	s.log.Info("shutting down deleter service, flushing remaining items",
		zap.Int("remaining", len(s.in)),
	)
	for {
		select {
		case it, ok := <-s.in:
			if !ok {
				return
			}
			buf = append(buf, it)
			if len(buf) >= 100 {
				s.flush(buf)
				buf = buf[:0]
			}
		default:
			if len(buf) > 0 {
				s.flush(buf)
			}
			return
		}
	}
}

func (s *Service) flush(buf []item) {
	if len(buf) == 0 {
		return
	}

	byUser := make(map[string][]string)
	for _, it := range buf {
		byUser[it.userID] = append(byUser[it.userID], it.id)
	}

	for userID, ids := range byUser {
		if err := s.repo.DeleteBatch(userID, ids); err != nil {
			s.log.Error("delete batch failed",
				zap.String("user_id", userID),
				zap.Int("count", len(ids)),
				zap.Error(err),
			)
		} else {
			s.log.Debug("delete batch successful",
				zap.String("user_id", userID),
				zap.Int("count", len(ids)),
			)
		}
	}
}

func (s *Service) Shutdown(timeout time.Duration) error {
	s.log.Info("shutting down deleter service")
	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("deleter service stopped gracefully")
		return nil
	case <-time.After(timeout):
		return errors.New("deleter shutdown timeout")
	}
}
