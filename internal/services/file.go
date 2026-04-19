package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"0xfer/internal/repositories"

	"github.com/oklog/ulid"
)

type FileService struct {
	db  *repositories.DBRepository
	fs  *repositories.FileRepository
	ttl time.Duration
}

func NewFileService(db *repositories.DBRepository, fs *repositories.FileRepository, ttl time.Duration) *FileService {
	return &FileService{db: db, fs: fs, ttl: ttl}
}

type UploadResult struct {
	ID       string
	Secret   string
	Filename string
	Size     int64
	Expires  time.Time
}

func (s *FileService) Upload(ctx context.Context, filename string, contentType string, size int64, body io.Reader, ttl time.Duration, maxDownloads int) (*UploadResult, error) {
	id := generateID()
	secret := generateSecret()

	now := time.Now()
	expireTime := now.Add(ttl)
	f := &repositories.File{
		ID:                 id,
		Secret:             secret,
		Filename:           filename,
		ContentType:        contentType,
		Size:               size,
		CreatedAt:          now,
		ExpiresAt:          expireTime,
		DownloadCount:      0,
		DownloadsRemaining: maxDownloads,
	}

	if err := s.fs.Save(id, body); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	if err := s.db.Create(ctx, f); err != nil {
		_ = s.fs.Delete(id)

		return nil, fmt.Errorf("create file in db: %w", err)
	}

	return &UploadResult{
		ID:       id,
		Secret:   secret,
		Filename: filename,
		Size:     size,
		Expires:  expireTime,
	}, nil
}

func (s *FileService) Get(ctx context.Context, id string) (*repositories.File, io.ReadCloser, error) {
	f, err := s.db.Get(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get file from db: %w", err)
	}

	if f.DownloadsRemaining != -1 {
		decremented, err := s.db.DecrementDownloadsRemaining(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("decrement downloads: %w", err)
		}
		if !decremented {
			return nil, nil, repositories.ErrNotFound
		}
	}

	file, err := s.fs.Get(id)
	if err != nil {
		return nil, nil, fmt.Errorf("get file from storage: %w", err)
	}

	_ = s.db.IncrementDownloadCount(ctx, id)

	return f, file, nil
}

func (s *FileService) Delete(ctx context.Context, id, secret string) error {
	f, err := s.db.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get file from db: %w", err)
	}

	if f.Secret != secret {
		return repositories.ErrNotFound
	}

	if err := s.fs.Delete(id); err != nil {
		return fmt.Errorf("delete file from storage: %w", err)
	}

	if err := s.db.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete file from db: %w", err)
	}

	return nil
}

func generateID() string {
	t := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(t), rand.Reader)

	return id.String()[:10]
}

func generateSecret() string {
	b := make([]byte, 8)
	rand.Read(b)

	return toBase62(b)
}

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func toBase62(b []byte) string {
	result := make([]byte, len(b))
	for i, v := range b {
		result[i] = base62Chars[int(v)%len(base62Chars)]
	}

	return string(result)
}
