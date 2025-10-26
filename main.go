package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	linewebhook "nomnomhub/app/line_webhook"
	"nomnomhub/internal/config"
	"nomnomhub/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// load configuration
	cfg := config.Load()

	// connect to database
	db := database.Connect(cfg.DSN())
	defer db.Close()

	// example: print app info
	log.Printf("🚀 %s is running in %s mode", cfg.Server.Name, cfg.Server.Env)

	r := gin.Default()
	registerRoutes(r, *cfg)

	log.Printf("🚀 Server running on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)

	// graceful shutdown (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("🧹 Shutting down gracefully...")
}

func registerRoutes(r *gin.Engine, cfg config.Config) {
	lineWebhookHandler := linewebhook.NewHandler(cfg.LineWebhook)
	r.POST("/webhook/line", lineWebhookHandler.LineWebhook)
}
