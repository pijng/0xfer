package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("file not found")
)

type File struct {
	ID                 string
	Secret             string
	Filename           string
	ContentType        string
	Size               int64
	CreatedAt          time.Time
	ExpiresAt          time.Time
	DownloadCount      int
	DownloadsRemaining int
}

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(path string) (*DBRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &DBRepository{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		secret TEXT NOT NULL,
		filename TEXT NOT NULL,
		content_type TEXT,
		size INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		download_count INTEGER DEFAULT 0,
		downloads_remaining INTEGER DEFAULT -1
	);
	CREATE INDEX IF NOT EXISTS idx_files_expires ON files(expires_at);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

func (r *DBRepository) Create(ctx context.Context, f *File) error {
	query := `
	INSERT INTO files (id, secret, filename, content_type, size, created_at, expires_at, downloads_remaining)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, f.ID, f.Secret, f.Filename, f.ContentType, f.Size, f.CreatedAt.Unix(), f.ExpiresAt.Unix(), f.DownloadsRemaining)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	return nil
}

func (r *DBRepository) Get(ctx context.Context, id string) (*File, error) {
	query := `
	SELECT id, secret, filename, content_type, size, created_at, expires_at, download_count, downloads_remaining
	FROM files WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var f File
	var createdAt, expiresAt int64
	err := row.Scan(&f.ID, &f.Secret, &f.Filename, &f.ContentType, &f.Size, &createdAt, &expiresAt, &f.DownloadCount, &f.DownloadsRemaining)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	f.CreatedAt = time.Unix(createdAt, 0)
	f.ExpiresAt = time.Unix(expiresAt, 0)

	return &f, nil
}

func (r *DBRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM files WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *DBRepository) IncrementDownloadCount(ctx context.Context, id string) error {
	query := `UPDATE files SET download_count = download_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment download count: %w", err)
	}

	return nil
}

func (r *DBRepository) DecrementDownloadsRemaining(ctx context.Context, id string) (bool, error) {
	query := `UPDATE files SET downloads_remaining = downloads_remaining - 1 WHERE id = ? AND downloads_remaining > 0`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("decrement downloads: %w", err)
	}

	rows, _ := result.RowsAffected()

	return rows > 0, nil
}

func (r *DBRepository) GetExpired(ctx context.Context) ([]File, error) {
	query := `
	SELECT id, secret, filename, content_type, size, created_at, expires_at, download_count, downloads_remaining
	FROM files WHERE expires_at < ?
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("query expired: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var files []File
	for rows.Next() {
		var f File
		var createdAt, expiresAt int64
		if err := rows.Scan(&f.ID, &f.Secret, &f.Filename, &f.ContentType, &f.Size, &createdAt, &expiresAt, &f.DownloadCount, &f.DownloadsRemaining); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		f.CreatedAt = time.Unix(createdAt, 0)
		f.ExpiresAt = time.Unix(expiresAt, 0)
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return files, nil
}

func (r *DBRepository) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}
