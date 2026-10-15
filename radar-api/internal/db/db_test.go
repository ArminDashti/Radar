package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationsAndSeedCreateSamples(t *testing.T) {
	ctx := context.Background()
	conn, err := Connect(ctx, filepath.Join(t.TempDir(), "radar.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for _, path := range []string{"../../migrations/001_init.sql", "../../migrations/002_accounts.sql"} {
		if err := RunMigration(ctx, conn, path); err != nil {
			t.Fatal(err)
		}
		if err := RunMigration(ctx, conn, path); err != nil {
			t.Fatalf("rerunning %s: %v", path, err)
		}
	}
	if err := Seed(ctx, conn); err != nil {
		t.Fatal(err)
	}

	var samples int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM samples`).Scan(&samples); err != nil {
		t.Fatal(err)
	}
	if samples == 0 {
		t.Fatal("expected seeded samples, got 0")
	}
}
