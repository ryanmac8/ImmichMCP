package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the ImmichMCP server.
type Config struct {
	BaseURL     string
	APIKey      string
	MaxPageSize int
	DownloadMode string
	Port        int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	baseURL := firstNonEmpty(
		os.Getenv("IMMICH_BASE_URL"),
		os.Getenv("IMMICH_URL"),
	)
	if baseURL == "" {
		return nil, fmt.Errorf("IMMICH_BASE_URL is required")
	}

	apiKey := firstNonEmpty(
		os.Getenv("IMMICH_API_KEY"),
		os.Getenv("IMMICH_TOKEN"),
	)
	if apiKey == "" {
		return nil, fmt.Errorf("IMMICH_API_KEY is required")
	}

	maxPageSize := 100
	if s := os.Getenv("MAX_PAGE_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxPageSize = n
		}
	}

	downloadMode := firstNonEmpty(os.Getenv("DOWNLOAD_MODE"), "url")

	port := 5000
	if s := os.Getenv("MCP_PORT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			port = n
		}
	}

	return &Config{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		MaxPageSize:  maxPageSize,
		DownloadMode: downloadMode,
		Port:         port,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
