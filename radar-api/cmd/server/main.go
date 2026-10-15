package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ArminDashti/radar-api/internal/config"
	"github.com/ArminDashti/radar-api/internal/db"
	"github.com/ArminDashti/radar-api/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.RunMigration(ctx, database, "migrations/001_init.sql"); err != nil {
		log.Fatal(err)
	}
	if err := db.RunMigration(ctx, database, "migrations/002_accounts.sql"); err != nil {
		log.Fatal(err)
	}
	if err := db.Seed(ctx, database); err != nil {
		log.Fatal(err)
	}

	server := &handlers.Server{DB: database, JWTSecret: cfg.JWTSecret, LogoDir: cfg.LogoDir}
	router := gin.New()
	// Trust private/loopback proxies so ClientIP uses X-Forwarded-For behind HAProxy.
	_ = router.SetTrustedProxies([]string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1"})
	router.Use(gin.Logger(), gin.Recovery(), cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	api.POST("/auth/login", server.Login)
	api.POST("/auth/signup", server.Signup)
	api.GET("/logos/:filename", server.ServeLogo)
	api.GET("/probes", server.ListProbes)
	api.GET("/grid/hosts", server.HostGrid)
	api.GET("/grid/probes", server.ProbeGrid)
	web := api.Group("")
	web.Use(server.WebAuth())
	web.GET("/me/host-prefs", server.GetHostPrefs)
	web.PUT("/me/host-prefs", server.PutHostPrefs)
	web.GET("/host-requests", server.ListHostRequests)
	web.POST("/host-requests", server.CreateHostRequest)
	admin := web.Group("")
	admin.Use(server.AdminOnly())
	admin.PUT("/probes/:id", server.UpdateProbe)
	admin.GET("/hosts", server.ListHosts)
	admin.POST("/hosts", server.CreateHost)
	admin.PUT("/hosts/:id", server.UpdateHost)
	admin.DELETE("/hosts/:id", server.DeleteHost)
	admin.POST("/hosts/:id/logo", server.UploadHostLogo)
	admin.POST("/hosts/:id/test", server.TestHost)
	admin.GET("/admin/stats", server.AdminStats)
	admin.POST("/host-requests/:id/approve", server.ApproveHostRequest)
	admin.POST("/host-requests/:id/reject", server.RejectHostRequest)
	agent := api.Group("/agent")
	agent.Use(server.AgentAuth())
	agent.GET("/targets", server.AgentTargets)
	agent.POST("/samples", server.AgentSamples)

	log.Printf("radar API listening on :%s", cfg.Port)
	if err := router.Run("0.0.0.0:" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
