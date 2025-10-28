package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	linewebhook "nomnomhub/app/line_webhook"
	"nomnomhub/internal/config"
	"nomnomhub/internal/database"

	appLog "nomnomhub/internal/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// load configuration
	cfg := config.Load()

	// log
	logger := appLog.New(cfg.Server.Env)
	defer logger.Sync()

	// connect to database
	db := database.Connect(cfg.DSN())
	defer db.Close()

	if strings.EqualFold(cfg.Server.Env, "prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	registerRoutes(r, logger, *cfg)

	logger.Info("server start", zap.String("port", cfg.Server.Port))
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("cannot start server", zap.Error(err))
	}

	// graceful shutdown (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("🧹 Shutting down gracefully...")
}

func registerRoutes(r *gin.Engine, logger *zap.Logger, cfg config.Config) {
	lineWebhookHandler := linewebhook.NewHandler(logger, cfg.LineWebhook)
	r.POST("/line/webhook", lineWebhookHandler.LineWebhook)
}
