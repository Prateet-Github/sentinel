package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRateLimitConfig(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "config.yaml")

	data := []byte(`
server:
  port: 8080

rate_limit:
  enabled: true
  capacity: 100
  refill_rate: 100

routes:
  - method: GET
    path: /users
    backend: users-service
`)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.RateLimit.Enabled {
		t.Fatal("expected rate limiting to be enabled")
	}

	if cfg.RateLimit.Capacity != 100 {
		t.Fatalf(
			"capacity = %v, want 100",
			cfg.RateLimit.Capacity,
		)
	}

	if cfg.RateLimit.RefillRate != 100 {
		t.Fatalf(
			"refill rate = %v, want 100",
			cfg.RateLimit.RefillRate,
		)
	}
}
