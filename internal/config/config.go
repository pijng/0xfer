package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Addr    string
	DataDir string
	MaxSize int64
	TTL     time.Duration
	DBPath  string
	BaseURL string
}

const defaultMaxSize = 100 << 20
const defaultTTL = 168 * time.Hour

func Load() *Config {
	c := &Config{
		Addr:    getEnvOrDefault("0XFER_ADDR", ":2052"),
		DataDir: getEnvOrDefault("0XFER_DATA_DIR", "./data"),
		MaxSize: parseSize(getEnvOrDefault("0XFER_MAX_SIZE", "")),
		TTL:     parseDur(getEnvOrDefault("0XFER_TTL", "")),
		DBPath:  os.Getenv("0XFER_DB"),
		BaseURL: getEnvOrDefault("0XFER_BASE_URL", ""),
	}

	flag.StringVar(&c.Addr, "addr", c.Addr, "Server address")
	flag.StringVar(&c.DataDir, "data-dir", c.DataDir, "Directory for files")
	flag.StringVar(&c.DBPath, "db", c.DBPath, "SQLite path")
	flag.StringVar(&c.BaseURL, "base-url", c.BaseURL, "Base URL for links")
	flag.Parse()

	if c.DBPath == "" {
		c.DBPath = c.DataDir + "/0xfer.db"
	}

	return c
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func parseSize(s string) int64 {
	if s == "" {
		return defaultMaxSize
	}

	var value int64
	var unit string
	if _, err := fmt.Sscanf(s, "%d%s", &value, &unit); err != nil {
		if _, err := fmt.Sscanf(s, "%d", &value); err != nil {
			return 0
		}

		return value
	}

	switch unit {
	case "KB", "kb":
		return value * 1024
	case "MB", "mb":
		return value * 1024 * 1024
	case "GB", "gb":
		return value * 1024 * 1024 * 1024
	}

	return value
}

func parseDur(s string) time.Duration {
	if s == "" {
		return defaultTTL
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}

	return defaultTTL
}
