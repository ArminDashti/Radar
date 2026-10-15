package handlers

import (
	"net/http"
	"time"

	"github.com/ArminDashti/radar-api/internal/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) AgentTargets(c *gin.Context) {
	rows, err := s.DB.QueryContext(requestContext(c), `
		SELECT id, name, host, http_enabled, icmp_enabled
		FROM endpoints WHERE active ORDER BY lower(name), id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list targets"})
		return
	}
	defer rows.Close()

	targets := make([]models.Target, 0)
	for rows.Next() {
		var target models.Target
		var httpEnabled, icmpEnabled bool
		if err := rows.Scan(&target.ID, &target.Name, &target.Host, &httpEnabled, &icmpEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read targets"})
			return
		}
		if httpEnabled {
			target.Protocols = append(target.Protocols, "http")
		}
		if icmpEnabled {
			target.Protocols = append(target.Protocols, "icmp")
		}
		targets = append(targets, target)
	}
	c.JSON(http.StatusOK, targets)
}

func (s *Server) AgentSamples(c *gin.Context) {
	var input struct {
		Samples []models.SampleInput `json:"samples" binding:"required,min=1,max=5000"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "samples must contain 1 to 5000 items"})
		return
	}
	agentID := c.GetInt64(agentIDKey)
	tx, err := s.DB.BeginTx(requestContext(c), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not begin sample batch"})
		return
	}
	defer tx.Rollback()

	accepted := 0
	skipped := 0
	for _, sample := range input.Samples {
		if sample.Protocol != "http" && sample.Protocol != "icmp" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "protocol must be http or icmp"})
			return
		}
		observedAt := sample.ObservedAt.UTC().Truncate(time.Minute).Unix()
		if !sample.OK {
			sample.LatencyMS = nil
		}
		result, err := tx.ExecContext(requestContext(c), `
			INSERT INTO samples (agent_id, endpoint_id, protocol, observed_at, latency_ms, ok)
			SELECT ?, e.id, ?, ?, ?, ?
			FROM endpoints e
			WHERE e.id = ? AND e.active
			  AND ((? = 'http' AND e.http_enabled) OR (? = 'icmp' AND e.icmp_enabled))
			ON CONFLICT (agent_id, endpoint_id, protocol, observed_at)
			DO UPDATE SET latency_ms = excluded.latency_ms, ok = excluded.ok`,
			agentID, sample.Protocol, observedAt, sample.LatencyMS, sample.OK,
			sample.EndpointID, sample.Protocol, sample.Protocol)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not store sample"})
			return
		}
		affected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store sample"})
			return
		}
		if affected == 0 {
			// Host deleted/deactivated or protocol disabled mid-cycle — keep the rest of the batch.
			skipped++
			continue
		}
		accepted++
	}
	if _, err := tx.ExecContext(requestContext(c),
		`UPDATE agents SET last_seen_at = ?, public_ip = ? WHERE id = ?`,
		time.Now().UTC(), c.ClientIP(), agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update agent"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not commit sample batch"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"accepted": accepted, "skipped": skipped})
}
