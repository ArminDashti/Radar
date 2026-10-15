package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	radarauth "github.com/ArminDashti/radar-api/internal/auth"
	"github.com/gin-gonic/gin"
)

const (
	agentIDKey      = "agent_id"
	agentProbeIDKey = "agent_probe_id"
)

type Server struct {
	DB        *sql.DB
	JWTSecret string
	LogoDir   string
}

func (s *Server) WebAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := radarauth.ParseToken(s.JWTSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		role := claims.Role
		if role == "" {
			role = "user"
		}
		c.Set("username", claims.Username)
		c.Set("role", role)
		c.Next()
	}
}

func (s *Server) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		c.Next()
	}
}

func (s *Server) AgentAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing agent bearer token"})
			return
		}
		hash := radarauth.HashAgentToken(token)
		var storedHash string
		var agentID, probeID int64
		err := s.DB.QueryRowContext(c.Request.Context(),
			`SELECT id, probe_id, token_hash FROM agents WHERE token_hash = ?`, hash,
		).Scan(&agentID, &probeID, &storedHash)
		if err != nil || !radarauth.EqualTokenHash(storedHash, hash) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token"})
			return
		}
		c.Set(agentIDKey, agentID)
		c.Set(agentProbeIDKey, probeID)
		c.Next()
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	scheme, token, found := strings.Cut(c.GetHeader("Authorization"), " ")
	return token, found && strings.EqualFold(scheme, "Bearer") && token != ""
}

func requestContext(c *gin.Context) context.Context {
	return c.Request.Context()
}
