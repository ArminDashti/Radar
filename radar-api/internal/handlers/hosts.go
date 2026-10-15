package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ArminDashti/radar-api/internal/models"
	"github.com/gin-gonic/gin"
)

type hostInput struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	HTTPEnabled bool   `json:"http_enabled"`
	ICMPEnabled bool   `json:"icmp_enabled"`
	ProbeID     *int64 `json:"probe_id"`
	Active      bool   `json:"active"`
}

func bindHostInput(c *gin.Context) (hostInput, bool) {
	var input hostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and host are required"})
		return hostInput{}, false
	}
	input.Name, input.Host = strings.TrimSpace(input.Name), strings.TrimSpace(input.Host)
	if input.Name == "" || input.Host == "" || (!input.HTTPEnabled && !input.ICMPEnabled) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, host, and at least one protocol are required"})
		return hostInput{}, false
	}
	return input, true
}

func scanHost(row interface {
	Scan(dest ...any) error
}) (models.Endpoint, error) {
	var host models.Endpoint
	err := row.Scan(&host.ID, &host.Name, &host.Host, &host.HTTPEnabled, &host.ICMPEnabled, &host.ProbeID, &host.Active, &host.LogoIcon, &host.CreatedAt)
	return host, err
}

func (s *Server) ListHosts(c *gin.Context) {
	rows, err := s.DB.QueryContext(requestContext(c), `
		SELECT id, name, host, http_enabled, icmp_enabled, probe_id, active, COALESCE(logo_icon, ''), created_at
		FROM endpoints ORDER BY lower(name), id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list hosts"})
		return
	}
	defer rows.Close()

	hosts := make([]models.Endpoint, 0)
	for rows.Next() {
		host, err := scanHost(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read hosts"})
			return
		}
		hosts = append(hosts, host)
	}
	c.JSON(http.StatusOK, hosts)
}

func (s *Server) CreateHost(c *gin.Context) {
	input, ok := bindHostInput(c)
	if !ok {
		return
	}
	host, err := scanHost(s.DB.QueryRowContext(requestContext(c), `
		INSERT INTO endpoints (name, host, http_enabled, icmp_enabled, probe_id, active)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, name, host, http_enabled, icmp_enabled, probe_id, active, COALESCE(logo_icon, ''), created_at`,
		input.Name, input.Host, input.HTTPEnabled, input.ICMPEnabled, input.ProbeID, input.Active,
	))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create host"})
		return
	}
	c.JSON(http.StatusCreated, host)
}

func (s *Server) UpdateHost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host id"})
		return
	}
	input, ok := bindHostInput(c)
	if !ok {
		return
	}
	host, err := scanHost(s.DB.QueryRowContext(requestContext(c), `
		UPDATE endpoints
		SET name = ?, host = ?, http_enabled = ?, icmp_enabled = ?, probe_id = ?, active = ?
		WHERE id = ?
		RETURNING id, name, host, http_enabled, icmp_enabled, probe_id, active, COALESCE(logo_icon, ''), created_at`,
		input.Name, input.Host, input.HTTPEnabled, input.ICMPEnabled, input.ProbeID, input.Active, id,
	))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	c.JSON(http.StatusOK, host)
}

func (s *Server) DeleteHost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host id"})
		return
	}
	result, err := s.DB.ExecContext(requestContext(c), `DELETE FROM endpoints WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete host"})
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete host"})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
