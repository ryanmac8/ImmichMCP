package config

import (
	"os"
	"testing"
)

func TestLoadSuccess(t *testing.T) {
	t.Setenv("IMMICH_BASE_URL", "http://immich.local:2283")
	t.Setenv("IMMICH_API_KEY", "test-key")
	t.Setenv("MCP_PORT", "8080")
	t.Setenv("MAX_PAGE_SIZE", "50")
	t.Setenv("DOWNLOAD_MODE", "inline")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "http://immich.local:2283" {
		t.Errorf("unexpected base URL: %s", cfg.BaseURL)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("unexpected API key")
	}
	if cfg.Port != 8080 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.MaxPageSize != 50 {
		t.Errorf("unexpected max page size: %d", cfg.MaxPageSize)
	}
	if cfg.DownloadMode != "inline" {
		t.Errorf("unexpected download mode: %s", cfg.DownloadMode)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("IMMICH_BASE_URL", "http://immich.local")
	t.Setenv("IMMICH_API_KEY", "key")
	os.Unsetenv("MCP_PORT")
	os.Unsetenv("MAX_PAGE_SIZE")
	os.Unsetenv("DOWNLOAD_MODE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 5000 {
		t.Errorf("expected default port 5000, got %d", cfg.Port)
	}
	if cfg.MaxPageSize != 100 {
		t.Errorf("expected default max page size 100, got %d", cfg.MaxPageSize)
	}
	if cfg.DownloadMode != "url" {
		t.Errorf("expected default download mode 'url', got %s", cfg.DownloadMode)
	}
}

func TestLoadMissingBaseURL(t *testing.T) {
	os.Unsetenv("IMMICH_BASE_URL")
	os.Unsetenv("IMMICH_URL")
	t.Setenv("IMMICH_API_KEY", "key")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing base URL")
	}
}

func TestLoadMissingAPIKey(t *testing.T) {
	t.Setenv("IMMICH_BASE_URL", "http://immich.local")
	os.Unsetenv("IMMICH_API_KEY")
	os.Unsetenv("IMMICH_TOKEN")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestLoadFallbackEnvVars(t *testing.T) {
	os.Unsetenv("IMMICH_BASE_URL")
	t.Setenv("IMMICH_URL", "http://fallback.local")
	os.Unsetenv("IMMICH_API_KEY")
	t.Setenv("IMMICH_TOKEN", "fallback-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "http://fallback.local" {
		t.Errorf("expected fallback base URL")
	}
	if cfg.APIKey != "fallback-token" {
		t.Errorf("expected fallback API key")
	}
}
