# Radar API

Radar API is the Gin/SQLite hub for collecting agent latency samples and serving WebUI grids. It uses Go 1.22, `database/sql` with the pure-Go `modernc.org/sqlite` driver, JWT WebUI authentication, SHA-256 agent tokens, and Docker Compose.

Copy `.env.example` to `.env`, then start `go run ./cmd/server`. The database is a single SQLite file set by `DATABASE_PATH`; Docker Compose stores it in the `radar-sqlite-vol` volume.
