package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
	"github.com/kuromii5/chat-bot-chat-service/config"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres/message"
	httpHandlers "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http"
	"github.com/kuromii5/chat-bot-chat-service/internal/service"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

func main() {
	cfg := config.MustLoad()
	setupLogger(cfg.Log.Level)
	validator.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := postgres.New(&cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	messageRepo := message.NewRepository(db)
	chatService := service.NewService(messageRepo)
	chatHandler := httpHandlers.NewHandler(chatService)

	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpiry, cfg.JWT.RefreshTokenExpiry)
	router := httpHandlers.NewRouter(chatHandler, jwtManager)
	server := httpHandlers.NewServer(cfg.Server.Host, cfg.Server.Port, router)
	if err := server.Start(); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}

	server.WaitAndShutdown(ctx)
}

func setupLogger(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}
