package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ArminDashti/radar-api/internal/probe"
	"github.com/gin-gonic/gin"
)

type hostTestResult struct {
	Protocol  string   `json:"protocol"`
	OK        bool     `json:"ok"`
	LatencyMS *float64 `json:"latency_ms"`
	Error     string   `json:"error"`
}

type hostTestResponse struct {
	ID      int64            `json:"id"`
	Name    string           `json:"name"`
	Host    string           `json:"host"`
	Results []hostTestResult `json:"results"`
}

func (s *Server) TestHost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host id"})
		return
	}

	var name, host string
	var httpEnabled, icmpEnabled bool
	err = s.DB.QueryRowContext(requestContext(c), `
		SELECT name, host, http_enabled, icmp_enabled
		FROM endpoints WHERE id = ?`, id,
	).Scan(&name, &host, &httpEnabled, &icmpEnabled)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	if !httpEnabled && !icmpEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no protocols enabled"})
		return
	}

	ctx := requestContext(c)
	results := make([]hostTestResult, 0, 2)
	var mu sync.Mutex
	var wg sync.WaitGroup
	setResult := func(index int, r hostTestResult) {
		mu.Lock()
		for len(results) <= index {
			results = append(results, hostTestResult{})
		}
		results[index] = r
		mu.Unlock()
	}

	next := 0
	if httpEnabled {
		index := next
		next++
		wg.Add(1)
		go func() {
			defer wg.Done()
			latency, ok, errMsg := probe.HTTP(ctx, host, 5*time.Second)
			setResult(index, hostTestResult{Protocol: "http", OK: ok, LatencyMS: latency, Error: errMsg})
		}()
	}
	if icmpEnabled {
		index := next
		next++
		wg.Add(1)
		go func() {
			defer wg.Done()
			latency, ok, errMsg := probe.ICMP(ctx, host, 3*time.Second)
			setResult(index, hostTestResult{Protocol: "icmp", OK: ok, LatencyMS: latency, Error: errMsg})
		}()
	}
	wg.Wait()

	c.JSON(http.StatusOK, hostTestResponse{
		ID: id, Name: name, Host: host, Results: results,
	})
}
