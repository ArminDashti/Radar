package config

import "testing"

func TestLoadAppliesDefaults(t *testing.T) {
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("PORT", "")
	t.Setenv("CORS_ORIGINS", "")

	cfg := Load()
	if cfg.DatabasePath != "/data/radar.db" {
		t.Fatalf("DatabasePath = %q, want /data/radar.db", cfg.DatabasePath)
	}
	if cfg.Port != "8088" {
		t.Fatalf("Port = %q, want 8088", cfg.Port)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("CORSOrigins length = %d, want 2", len(cfg.CORSOrigins))
	}
}
