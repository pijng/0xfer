# 0xfer

Minimal file sharing service. Upload files via curl, get a short URL, download the same way. Self-hosted alternative to transfer.sh or 0x0.st.

[![Go](https://img.shields.io/badge/Go-1.26-blue?style=for-the-badge&logo=go)](https://go.dev/)
![SQLite](https://img.shields.io/badge/SQLite-pure%20Go-green?style=for-the-badge&logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-ready-blue?style=for-the-badge&logo=docker)
![systemd](https://img.shields.io/badge/systemd-native-orange?style=for-the-badge&logo=linux)

## What is this

`0xfer` is a lightweight file sharing service built with Go and no external dependencies. Upload a file via curl, get a short URL for download. No auth, no registration, no complex logic. One binary, one SQLite database for metadata, files stored locally on disk.

Perfect for:

- Quick file transfers between servers
- Temporary file sharing within your infrastructure
- Self-hosted solution

## Quick Start

### Docker

```bash
docker run -d \
  --name 0xfer \
  -p 2052:2052 \
  -v 0xfer-data:/app/data \
  0xfer:latest
```

Or with docker-compose:

```bash
docker-compose up -d
```

### Binary

```bash
# Build
go build -o 0xfer ./cmd/server

# Run
./0xfer -addr :2052 -data-dir ./data
```

### systemd (for binary)

```bash
sudo cp 0xfer.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now 0xfer
```

## Usage

### Upload

```bash
# Multipart form
curl -F "file=@myfile.zip" http://localhost:2052/

# Or just POST with body
curl -X POST --data-binary @myfile.zip http://localhost:2052/

# Upload from remote URL
curl -F "url=https://example.com/file.zip" http://localhost:2052/

# Custom expiry (shorter than default)
curl -F "file=@myfile.zip" -F "expires=1h" http://localhost:2052/
curl -F "file=@myfile.zip" -F "expires=24h" http://localhost:2052/

# Limit max downloads
curl -F "file=@myfile.zip" -H "Max-Downloads: 1" http://localhost:2052/
```

Response:

```
File: myfile.zip (15.2MB)
Download: curl -JLOf http://localhost:2052/d/01ARZ3NDEKTSV4RRFFQ69G5FAV
Delete:   curl -X DELETE http://localhost:2052/d/01ARZ3NDEKTSV4RRFFQ69G5FAV/abcdefgh
Expires:  2026-04-25T18:30:00Z
```

### Download

```bash
curl -JLOf http://localhost:2052/d/01ARZ3NDEKTSV4RRFFQ69G5FAV
```

### Delete

```bash
curl -X DELETE http://localhost:2052/d/01ARZ3NDEKTSV4RRFFQ69G5FAV/abcdefgh
```

### Health check

```bash
curl http://localhost:2052/health
# OK
```

## Configuration

| Flag        | ENV              | Default         | Description                  |
| ----------- | ---------------- | --------------- | ---------------------------- |
| `-addr`     | `0XFER_ADDR`     | `:2052`         | Server address               |
| `-data-dir` | `0XFER_DATA_DIR` | `./data`        | Directory for files and DB   |
| `-max-size` | `0XFER_MAX_SIZE` | `100MB`         | Max file size                |
| `-ttl`      | `0XFER_TTL`      | `168h` (7 days) | File time-to-live            |
| `-base-url` | `0XFER_BASE_URL` | (auto)          | Base URL for generated links |

Example:

```bash
./0xfer \
  -addr :2052 \
  -data-dir /var/lib/0xfer \
  -max-size 500MB \
  -ttl 24h \
  -base-url https://files.example.com
```

## Data Structure

```
data/
├── 0xfer.db           # SQLite with metadata
├── ab/                # Files sharded by first 2 chars of ID
│   └── 01ARZ3NDEKTSV4RRFFQ69G5FAV.bin
└── xy/
    └── ...
```

Each file gets a unique ID (10 chars, ULID) and secret (8 random chars) for secure deletion.

## Features

- **Upload from URL**: fetch files from remote servers (`url=...` parameter)
- **Custom expiry**: override default TTL (`expires=1h`, `expires=24h`, etc.)
- **Max downloads**: delete after N downloads (`Max-Downloads: 1` header)
- **SSRF protection**: blocks requests to private/internal networks
- **Auto-cleanup**: expired files removed every minute
- **Sharding**: files distributed across subdirectories (first 2 chars of ID) — avoids Too Many Files issue
- **Content-Type**: auto-detected from file body
- **Graceful shutdown**: properly handles SIGTERM
- **JSON logs**: structured output via slog

## License

MIT License. See LICENSE file for details.
