## 1. Overview

Сервис для загрузки и скачивания файлов с короткими URL. Аналог transfer.sh / 0x0.st.
**Технологии**: Go (чистый, без внешних зависимостей) + SQLite (чистый Go через `modernc.org/sqlite`)
**Деплой**: Docker + Binary с systemd

---

## 2. Architecture

### 2.1 Слои

```
┌─────────────────────────────────────────┐
│              Handlers (HTTP)            │
│   upload.go | download.go | delete.go  │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│              Services                   │
│        file.go | cleanup.go             │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│            Repositories                 │
│          db.go | file.go                │
└────────────────────┬────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
   ┌────▼────┐             ┌─────▼─────┐
   │  Disk   │             │  SQLite   │
   │ (files) │             │ (metadata)│
   └─────────┘             └───────────┘
```

---

## 3. API Specification

### 3.1 Upload (POST на корень /)

```bash
# Multipart form
curl -F "file=@myfile.zip" https://0xfer.sh/
# Или просто POST с бинарным телом
curl -X POST --data-binary @myfile.zip https://0xfer.sh/
```

**Response** (plain text):

```
File: myfile.zip (15MB)
Download: curl -JLO https://0xfer.sh/d/abc123
Delete:   curl -X DELETE https://0xfer.sh/d/abc123/xyz789
Expires:  7 days
```

### 3.2 Download (GET)

```bash
curl -JLO https://0xfer.sh/d/abc123
```

### 3.3 Delete (DELETE)

```bash
curl -X DELETE https://0xfer.sh/d/abc123/xyz789
```

### 3.4 Health (GET)

```bash
curl https://0xfer.sh/health
```

---

## 4. Configuration

| Flag        | ENV              | Default           | Description                  |
| ----------- | ---------------- | ----------------- | ---------------------------- |
| `-addr`     | `0XFER_ADDR`     | `:2052`           | Server address               |
| `-data-dir` | `0XFER_DATA_DIR` | `./data`          | Directory for files          |
| `-max-size` | `0XFER_MAX_SIZE` | `25MB`            | Max upload size              |
| `-ttl`      | `0XFER_TTL`      | `168h` (7 days)   | File TTL                     |
| `-db`       | `0XFER_DB`       | `./data/0xfer.db` | SQLite path                  |
| `-base-url` | `0XFER_BASE_URL` | (auto-detect)     | Base URL for generated links |

### 4.1 Base URL

## `0XFER_BASE_URL` используется для генерации ссылок в ответе. Если не задан — определяется автоматически из заголовков запроса (`Host` или `X-Forwarded-Host`).

## 5. Data Model

### 5.1 SQLite Schema

```sql
CREATE TABLE files (
    id          TEXT PRIMARY KEY,      -- short ID (e.g., "abc123")
    secret      TEXT NOT NULL,         -- delete token (e.g., "xyz789")
    filename    TEXT NOT NULL,         -- original filename
    content_type TEXT,                  -- MIME type
    size        INTEGER NOT NULL,       -- bytes
    created_at  INTEGER NOT NULL,       -- unix timestamp
    expires_at  INTEGER NOT NULL,      -- unix timestamp (created + TTL)
    download_count INTEGER DEFAULT 0
);
CREATE INDEX idx_files_expires ON files(expires_at);
```

### 5.2 File Storage

Файлы хранятся на диске с именем `{id}.bin` (без оригинального расширения, чтобы избежать конфликтов при одинаковых именах).

```
data/
├── 0xfer.db
├── ab/
│   └── c123abc123.bin
└── xy/
    └── z789xyz789.bin
```

## Используется двухуровневая директория (первые 2 символа ID) для избежания Too Many Files в одной директории.

## 6. Implementation Steps

### Phase 1: Core (2-3 hours)

1. [ ] **Project setup**
   ```
   cmd/server/main.go
   internal/
     config/config.go
     handlers/
       upload.go
       download.go
       delete.go
       health.go
     services/
       file.go
       cleanup.go
     repositories/
       db.go
       file.go
   ```
2. [ ] **Config** - flags + env vars
3. [ ] **Database** - SQLite connection, migrations
4. [ ] **File Service** - генерация ID, сохранение файла, запись в БД
5. [ ] **Upload Handler** - `POST /`, multipart + raw body

### Phase 2: Download & Delete (1-2 hours)

6. [ ] **Download Handler** - `GET /d/{id}`, streaming с диска
7. [ ] **Delete Handler** - `DELETE /d/{id}/{secret}`, verify + remove

### Phase 3: TTL & Cleanup (1-2 часа)

8. [ ] **Cleanup Service** - периодическая очистка просроченных файлов
   - Запускается в отдельной горутине
   - Запускается при старте (startup cleanup)
   - Проверяет `expires_at < now()` и удаляет

### Phase 4: Polish (1 hour)

9. [ ] **Logging** - Go's built-in `log/slog`
10. [ ] **Graceful shutdown** - SIGTERM handling
11. [ ] **Health check** - `GET /health`
12. [ ] **Base URL detection** - из request headers

### Phase 5: Deployment

13. [ ] **Dockerfile** - multi-stage, static binary
14. [ ] **docker-compose.yml** - dev environment с named volumes
15. [ ] **0xfer.service** - systemd unit file

---

## 7. Docker Volumes

```yaml
# docker-compose.yml
services:
  0xfer:
    build: .
    volumes:
      - 0xfer-data:/app/data
    ...
volumes:
  0xfer-data:
```

## SQLite база и файлы хранятся в named volume `0xfer-data` — сохраняется между перезапусками.

## 8. Key Design Decisions

### 8.1 ID Generation

Используем `github.com/oklog/ulid`:

- Sortable (по времени)
- Collision-resistant
- 10 символов (альфа numeric lowercase)

### 8.2 Delete Token

Генерируется отдельно от ID (8 случайных символов). Нужен чтобы нельзя было удалить чужой файл перебором.

### 8.3 File Naming

- На диске: `{id}.bin` — игнорируем оригинальное расширение
- В БД: сохраняем оригинальное имя для Content-Disposition при скачивании
- Решение проблемы "идентичных имен": каждый файл уникален по ID, поэтому конфликтов нет

### 8.4 SQLite

Используем `modernc.org/sqlite` — чистый Go, не требует CGO, компилируется в static binary.

### 8.5 TTL Cleanup

- При старте: удаляем все просроченные файлы
- В рантайме: запускаем cleanup раз в минуту
- Не блокирует основной поток — асинхронно

### 8.6 Upload on Root

## Определяем upload по HTTP методу POST (или PUT) на `/`. GET на `/` возвращает 404 или простую HTML страницу с инструкцией.

## 9. Directory Structure

```
0xfer/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   ├── upload.go
│   │   ├── download.go
│   │   ├── delete.go
│   │   └── health.go
│   ├── services/
│   │   ├── file.go
│   │   └── cleanup.go
│   ├── repositories/
│   │   ├── db.go
│   │   └── file.go
│   └── main.go
├── Dockerfile
├── docker-compose.yml
├── 0xfer.service
├── go.mod
├── go.sum
└── README.md
```

---

## 10. Future Enhancements (Later)

- [ ] Web UI (Drag & Drop)
- [ ] S3 backend
- [ ] Password protection per file
- [ ] One-time download link
- [ ] IP-based rate limiting
- [ ] Prometheus metrics
- [ ] CLI клиент

---

## 11. Estimation

| Phase             | Time     |
| ----------------- | -------- | --- |
| Core              | 2-3h     |
| Download & Delete | 1-2h     |
| TTL Cleanup       | 1-2h     |
| Polish            | 1h       |
| Deployment        | 1h       |
| **Total**         | **6-9h** | ]   |
