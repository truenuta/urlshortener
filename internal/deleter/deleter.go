package deleter

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Workers        int
	FlushThreshold int
	FlushInterval  time.Duration
	QueueCapacity  int
}

type BatchDeleter interface {
	DeleteBatch(userID string, shortURLs []string) error
}

type item struct {
	userID   string
	shortURL string
}

type Deleter struct {
	ch             chan item
	jobs           chan job
	repo           BatchDeleter
	logger         *zap.Logger
	workers        int
	flushThreshold int
	flushInterval  time.Duration
}

type job struct {
	userID    string
	shortURLs []string
}

func NewDeleter(repo BatchDeleter, logger *zap.Logger, cfg Config) *Deleter {
	return &Deleter{
		ch:             make(chan item, cfg.QueueCapacity),
		jobs:           make(chan job, cfg.QueueCapacity),
		repo:           repo,
		logger:         logger,
		workers:        cfg.Workers,
		flushThreshold: cfg.FlushThreshold,
		flushInterval:  cfg.FlushInterval,
	}
}

func (d *Deleter) ScheduleDeletion(ctx context.Context, userID string, shortURLs []string) {
	if len(shortURLs) == 0 {
		return
	}
	select {
	case d.jobs <- job{userID: userID, shortURLs: shortURLs}:
	case <-ctx.Done():
		d.logger.Warn("deletion dropped: context done", zap.String("UserID", userID))
	}
}

func (d *Deleter) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j := <-d.jobs:
					for _, shortURL := range j.shortURLs {
						select {
						case d.ch <- item{userID: j.userID, shortURL: shortURL}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}
	ticker := time.NewTicker(d.flushInterval)
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
			wg.Wait()
			flush()
			return
		case task := <-d.ch:
			pending[task.userID] = append(pending[task.userID], task.shortURL)
			pendingCount++
			if pendingCount >= d.flushThreshold {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
