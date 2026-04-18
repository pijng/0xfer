package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"0xfer/pkg/shortcut"
)

func TestDBRepository(t *testing.T) {
	tmp := t.TempDir() + "/test.db"
	defer func() { _ = os.Remove(tmp) }() //nolint:errcheck

	repo, err := NewDBRepository(tmp)
	shortcut.FatalIfErr(err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	now := time.Now()

	t.Run("create and get", func(t *testing.T) {
		f := &File{ //nolint:exhaustruct //nolint:exhaustruct
			ID:          "test123456",
			Secret:      "secret12",
			Filename:    "test.txt",
			ContentType: "text/plain",
			Size:        100,
			CreatedAt:   now,
			ExpiresAt:   now.Add(time.Hour),
		}

		err := repo.Create(ctx, f)
		shortcut.FatalIfErr(err)

		got, err := repo.Get(ctx, "test123456")
		shortcut.FatalIfErr(err)

		if got.Filename != "test.txt" {
			t.Errorf("filename = %q; want %q", got.Filename, "test.txt")
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := repo.Get(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("delete", func(t *testing.T) {
		f := &File{ //nolint:exhaustruct //nolint:exhaustruct
			ID:          "del123456",
			Secret:      "secret12",
			Filename:    "delete.txt",
			ContentType: "text/plain",
			Size:        100,
			CreatedAt:   now,
			ExpiresAt:   now.Add(time.Hour),
		}

		err := repo.Create(ctx, f)
		shortcut.FatalIfErr(err)

		err = repo.Delete(ctx, "del123456")
		shortcut.FatalIfErr(err)

		_, err = repo.Get(ctx, "del123456")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("increment download count", func(t *testing.T) {
		f := &File{ //nolint:exhaustruct //nolint:exhaustruct
			ID:          "cnt123456",
			Secret:      "secret12",
			Filename:    "count.txt",
			ContentType: "text/plain",
			Size:        100,
			CreatedAt:   now,
			ExpiresAt:   now.Add(time.Hour),
		}

		err := repo.Create(ctx, f)
		shortcut.FatalIfErr(err)

		err = repo.IncrementDownloadCount(ctx, "cnt123456")
		shortcut.FatalIfErr(err)

		got, err := repo.Get(ctx, "cnt123456")
		shortcut.FatalIfErr(err)

		if got.DownloadCount != 1 {
			t.Errorf("download count = %d; want 1", got.DownloadCount)
		}
	})

	t.Run("get expired", func(t *testing.T) {
		f := &File{ //nolint:exhaustruct
			ID:          "exp123456",
			Secret:      "secret12",
			Filename:    "expired.txt",
			ContentType: "text/plain",
			Size:        100,
			CreatedAt:   now,
			ExpiresAt:   now.Add(-time.Hour),
		}

		err := repo.Create(ctx, f)
		shortcut.FatalIfErr(err)

		expired, err := repo.GetExpired(ctx)
		shortcut.FatalIfErr(err)

		if len(expired) == 0 {
			t.Error("expected expired files, got none")
		}
	})
}
