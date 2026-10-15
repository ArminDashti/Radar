package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	radarauth "github.com/ArminDashti/radar-api/internal/auth"
	"github.com/ArminDashti/radar-api/internal/db"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func (s *Server) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	var passwordHash, role string
	if err := s.DB.QueryRowContext(requestContext(c),
		`SELECT password_hash, role FROM users WHERE username = ?`, input.Username,
	).Scan(&passwordHash, &role); err != nil ||
		bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	token, err := radarauth.IssueToken(s.JWTSecret, input.Username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "username": input.Username, "role": role})
}

func (s *Server) Signup(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	username := strings.TrimSpace(input.Username)
	if !usernamePattern.MatchString(username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be 3-32 characters: letters, digits, underscore"})
		return
	}
	if utf8.RuneCountInString(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}
	const role = "user"
	_, err = s.DB.ExecContext(requestContext(c),
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, string(passwordHash), role,
	)
	if err != nil {
		if db.IsUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create account"})
		return
	}
	token, err := radarauth.IssueToken(s.JWTSecret, username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "username": username, "role": role})
}

func (s *Server) lookupUserID(c *gin.Context) (int64, bool) {
	username, _ := c.Get("username")
	name, _ := username.(string)
	if name == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing username"})
		return 0, false
	}
	var userID int64
	if err := s.DB.QueryRowContext(requestContext(c),
		`SELECT id FROM users WHERE username = ?`, name,
	).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return 0, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
		return 0, false
	}
	return userID, true
}
