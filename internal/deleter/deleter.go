package deleter

import (
	"time"

	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/storage"
)

type item struct {
	userID string
	id     string
}

type Service struct {
	in   chan item
	repo storage.URLRepository
	log  *zap.Logger
}

func New(repo storage.URLRepository, log *zap.Logger) *Service {
	s := &Service{
		in:   make(chan item, 1024),
		repo: repo,
		log:  log,
	}
	go s.run()
	return s
}

func (s *Service) Submit(userID string, ids []string) {
	for _, id := range ids {
		s.in <- item{userID: userID, id: id}
	}
}

func (s *Service) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var buf []item

	for {
		select {
		case it := <-s.in:
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

func (s *Service) flush(buf []item) {
	byUser := make(map[string][]string)
	for _, it := range buf {
		byUser[it.userID] = append(byUser[it.userID], it.id)
	}
	for userID, ids := range byUser {
		if err := s.repo.DeleteBatch(userID, ids); err != nil {
			s.log.Error("delete batch failed", zap.String("user_id", userID), zap.Error(err))
		}
	}
}
