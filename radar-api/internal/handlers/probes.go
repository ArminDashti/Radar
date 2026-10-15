package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ArminDashti/radar-api/internal/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) ListProbes(c *gin.Context) {
	rows, err := s.DB.QueryContext(requestContext(c), `
		SELECT p.id, p.code, p.name, p.flag_icon, p.created_at,
			COALESCE((
				SELECT a.public_ip
				FROM agents a
				WHERE a.probe_id = p.id AND a.public_ip <> ''
				ORDER BY a.last_seen_at DESC NULLS LAST
				LIMIT 1
			), '') AS public_ip
		FROM probes p
		ORDER BY p.id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list probes"})
		return
	}
	defer rows.Close()

	probes := make([]models.Probe, 0)
	for rows.Next() {
		var probe models.Probe
		if err := rows.Scan(&probe.ID, &probe.Code, &probe.Name, &probe.FlagIcon, &probe.CreatedAt, &probe.PublicIP); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read probes"})
			return
		}
		probes = append(probes, probe)
	}
	c.JSON(http.StatusOK, probes)
}

func (s *Server) UpdateProbe(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe id"})
		return
	}
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var probe models.Probe
	err = s.DB.QueryRowContext(requestContext(c), `
		UPDATE probes SET name = ? WHERE id = ?
		RETURNING id, code, name, flag_icon, created_at`,
		input.Name, id,
	).Scan(&probe.ID, &probe.Code, &probe.Name, &probe.FlagIcon, &probe.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
		return
	}
	_ = s.DB.QueryRowContext(requestContext(c), `
		SELECT COALESCE((
			SELECT a.public_ip
			FROM agents a
			WHERE a.probe_id = ? AND a.public_ip <> ''
			ORDER BY a.last_seen_at DESC NULLS LAST
			LIMIT 1
		), '')`, id).Scan(&probe.PublicIP)
	c.JSON(http.StatusOK, probe)
}
