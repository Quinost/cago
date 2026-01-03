package internal

import (
	"fmt"
	"time"
	"context"
)

type CleanupWorker struct {
	cfg     *Config
	storage *Storage
}

func NewCleanupWorker(cfg *Config, storage *Storage) *CleanupWorker {
	return &CleanupWorker{
		cfg:     cfg,
		storage: storage,
	}
}

func (w *CleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := w.storage.CleanupExired()
			if count > 0 {
				fmt.Printf("Cleaned %d expired keys\n", count)
			}
		case <-ctx.Done():
			return
		}
	}
}