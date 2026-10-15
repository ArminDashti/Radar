package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type hostRequestRow struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Host         string     `json:"host"`
	HTTPEnabled  bool       `json:"http_enabled"`
	ICMPEnabled  bool       `json:"icmp_enabled"`
	Status       string     `json:"status"`
	ReviewerNote string     `json:"reviewer_note"`
	CreatedAt    time.Time  `json:"created_at"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
}

func (s *Server) CreateHostRequest(c *gin.Context) {
	userID, ok := s.lookupUserID(c)
	if !ok {
		return
	}
	var input struct {
		Name        string `json:"name" binding:"required"`
		Host        string `json:"host" binding:"required"`
		HTTPEnabled *bool  `json:"http_enabled"`
		ICMPEnabled *bool  `json:"icmp_enabled"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and host are required"})
		return
	}
	name := strings.TrimSpace(input.Name)
	host := strings.TrimSpace(input.Host)
	if name == "" || host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and host are required"})
		return
	}
	httpEnabled := true
	icmpEnabled := true
	if input.HTTPEnabled != nil {
		httpEnabled = *input.HTTPEnabled
	}
	if input.ICMPEnabled != nil {
		icmpEnabled = *input.ICMPEnabled
	}
	if !httpEnabled && !icmpEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one protocol must be enabled"})
		return
	}
	var row hostRequestRow
	err := s.DB.QueryRowContext(requestContext(c), `
		INSERT INTO host_requests (user_id, name, host, http_enabled, icmp_enabled)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, user_id, name, host, http_enabled, icmp_enabled, status, reviewer_note, created_at, reviewed_at`,
		userID, name, host, httpEnabled, icmpEnabled,
	).Scan(
		&row.ID, &row.UserID, &row.Name, &row.Host, &row.HTTPEnabled, &row.ICMPEnabled,
		&row.Status, &row.ReviewerNote, &row.CreatedAt, &row.ReviewedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create host request"})
		return
	}
	username, _ := c.Get("username")
	row.Username, _ = username.(string)
	c.JSON(http.StatusCreated, row)
}

func (s *Server) ListHostRequests(c *gin.Context) {
	userID, ok := s.lookupUserID(c)
	if !ok {
		return
	}
	role, _ := c.Get("role")
	status := strings.TrimSpace(c.Query("status"))
	args := []any{}
	where := []string{}
	if role != "admin" {
		args = append(args, userID)
		where = append(where, "r.user_id = ?")
	}
	if status != "" {
		args = append(args, status)
		where = append(where, "r.status = ?")
	}
	query := `
		SELECT r.id, r.user_id, u.username, r.name, r.host, r.http_enabled, r.icmp_enabled,
		       r.status, r.reviewer_note, r.created_at, r.reviewed_at
		FROM host_requests r
		JOIN users u ON u.id = r.user_id`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY r.created_at DESC"
	rows, err := s.DB.QueryContext(requestContext(c), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list host requests"})
		return
	}
	defer rows.Close()
	list := make([]hostRequestRow, 0)
	for rows.Next() {
		var row hostRequestRow
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.Username, &row.Name, &row.Host, &row.HTTPEnabled, &row.ICMPEnabled,
			&row.Status, &row.ReviewerNote, &row.CreatedAt, &row.ReviewedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read host requests"})
			return
		}
		list = append(list, row)
	}
	c.JSON(http.StatusOK, list)
}

func (s *Server) ApproveHostRequest(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}
	var note struct {
		ReviewerNote string `json:"reviewer_note"`
	}
	_ = c.ShouldBindJSON(&note)

	tx, err := s.DB.BeginTx(requestContext(c), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start transaction"})
		return
	}
	defer tx.Rollback()

	var req hostRequestRow
	err = tx.QueryRowContext(requestContext(c), `
		SELECT id, user_id, name, host, http_enabled, icmp_enabled, status
		FROM host_requests WHERE id = ?`, id,
	).Scan(&req.ID, &req.UserID, &req.Name, &req.Host, &req.HTTPEnabled, &req.ICMPEnabled, &req.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "host request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load host request"})
		return
	}
	if req.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "host request is not pending"})
		return
	}

	var endpointID int64
	err = tx.QueryRowContext(requestContext(c), `
		INSERT INTO endpoints (name, host, http_enabled, icmp_enabled, active)
		VALUES (?, ?, ?, ?, true)
		RETURNING id`,
		req.Name, req.Host, req.HTTPEnabled, req.ICMPEnabled,
	).Scan(&endpointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create host"})
		return
	}

	var reviewedAt time.Time
	err = tx.QueryRowContext(requestContext(c), `
		UPDATE host_requests
		SET status = 'approved', reviewer_note = ?, reviewed_at = ?
		WHERE id = ?
		RETURNING reviewed_at`,
		strings.TrimSpace(note.ReviewerNote), time.Now().UTC(), id,
	).Scan(&reviewedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not approve host request"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not commit approval"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          id,
		"status":      "approved",
		"endpoint_id": endpointID,
		"reviewed_at": reviewedAt,
	})
}

func (s *Server) RejectHostRequest(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}
	var note struct {
		ReviewerNote string `json:"reviewer_note"`
	}
	_ = c.ShouldBindJSON(&note)

	result, err := s.DB.ExecContext(requestContext(c), `
		UPDATE host_requests
		SET status = 'rejected', reviewer_note = ?, reviewed_at = ?
		WHERE id = ? AND status = 'pending'`,
		strings.TrimSpace(note.ReviewerNote), time.Now().UTC(), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reject host request"})
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reject host request"})
		return
	}
	if affected == 0 {
		var exists bool
		_ = s.DB.QueryRowContext(requestContext(c), `SELECT EXISTS(SELECT 1 FROM host_requests WHERE id = ?)`, id).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "host request not found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "host request is not pending"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "status": "rejected"})
}
