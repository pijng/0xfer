package services

import (
	"context"
	"log/slog"
	"time"

	"0xfer/internal/repositories"
)

type CleanupService struct {
	db       *repositories.DBRepository
	fs       *repositories.FileRepository
	interval time.Duration
	logger   *slog.Logger
}

func NewCleanupService(db *repositories.DBRepository, fs *repositories.FileRepository, interval time.Duration, logger *slog.Logger) *CleanupService {
	return &CleanupService{
		db:       db,
		fs:       fs,
		interval: interval,
		logger:   logger,
	}
}

func (s *CleanupService) Start(ctx context.Context) {
	s.cleanup(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

func (s *CleanupService) cleanup(ctx context.Context) {
	files, err := s.db.GetExpired(ctx)
	if err != nil {
		s.logger.Error("failed to get expired files", "err", err)

		return
	}

	for _, f := range files {
		if err := s.fs.Delete(f.ID); err != nil {
			s.logger.Warn("failed to delete file", "id", f.ID, "err", err)
		}
		if err := s.db.Delete(ctx, f.ID); err != nil {
			s.logger.Warn("failed to delete from db", "id", f.ID, "err", err)
		}
		s.logger.Info("deleted expired file", "id", f.ID, "filename", f.Filename)
	}
}
