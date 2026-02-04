package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/rabbitmq"
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

	pg, err := postgres.New(cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database", err)
	}
	defer pg.DB.Close()

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		logrus.Fatal("Failed to connect to rabbitmq", err)
	}
	defer rmq.Close()

	chatService := service.NewService(pg, pg, rmq)
	chatHandler := httpHandlers.NewHandler(chatService)
	notificationHandler := httpHandlers.NewNotificationHandler(rmq)

	router := httpHandlers.NewRouter(chatHandler, notificationHandler, cfg.JWT.Secret)
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
