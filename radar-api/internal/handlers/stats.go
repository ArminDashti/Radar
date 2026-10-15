package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type adminStats struct {
	DBOK          bool       `json:"db_ok"`
	Hosts         int64      `json:"hosts"`
	ActiveHosts   int64      `json:"active_hosts"`
	Probes        int64      `json:"probes"`
	Samples       int64      `json:"samples"`
	LastSampleAt  *time.Time `json:"last_sample_at"`
	DatabaseBytes int64      `json:"database_bytes"`
}

func (s *Server) AdminStats(c *gin.Context) {
	ctx := requestContext(c)
	stats := adminStats{}

	if err := s.DB.PingContext(ctx); err != nil {
		c.JSON(http.StatusOK, stats)
		return
	}
	stats.DBOK = true

	var lastSample sql.NullInt64
	err := s.DB.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM endpoints),
			(SELECT COUNT(*) FROM endpoints WHERE active),
			(SELECT COUNT(*) FROM probes),
			(SELECT COUNT(*) FROM samples),
			(SELECT MAX(observed_at) FROM samples),
			(SELECT page_count * page_size FROM pragma_page_count, pragma_page_size)
	`).Scan(&stats.Hosts, &stats.ActiveHosts, &stats.Probes, &stats.Samples, &lastSample, &stats.DatabaseBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load admin stats"})
		return
	}
	if lastSample.Valid {
		last := time.Unix(lastSample.Int64, 0).UTC()
		stats.LastSampleAt = &last
	}
	c.JSON(http.StatusOK, stats)
}
