package deleter

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type BatchDeleter interface {
	DeleteBatch(userID string, shortURLs []string) error
}

type item struct {
	userID   string
	shortURL string
}

type Deleter struct {
	ch     chan item
	repo   BatchDeleter
	logger *zap.Logger
}

func NewDeleter(repo BatchDeleter, logger *zap.Logger) *Deleter {
	return &Deleter{
		ch:     make(chan item, 1024),
		repo:   repo,
		logger: logger,
	}
}

func (d *Deleter) ScheduleDeletion(userID string, shortURLs []string) {
	if len(shortURLs) == 0 {
		return
	}
	go func() {
		for _, shortURL := range shortURLs {
			d.ch <- item{userID: userID, shortURL: shortURL}
		}

	}()
}

func (d *Deleter) Run(ctx context.Context) {
	const flushThreshold = 100
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	pending := make(map[string][]string)
	pendingCount := 0

	flush := func() {
		if pendingCount == 0 {
			return
		}
		for userID, shortURLs := range pending {
			if err := d.repo.DeleteBatch(userID, shortURLs); err != nil {
				d.logger.Error("batch delete failed", zap.Error(err))
			}
		}
		pending = make(map[string][]string)
		pendingCount = 0
	}
	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case task := <-d.ch:
			pending[task.userID] = append(pending[task.userID], task.shortURL)
			pendingCount++
			if pendingCount >= flushThreshold {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
