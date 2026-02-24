package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/rabbitmq"
	httpHandlers "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http"
	"github.com/kuromii5/chat-bot-chat-service/internal/service"
	"github.com/kuromii5/chat-bot-chat-service/pkg/tracing"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

func main() {
	cfg := config.MustLoad()
	setupLogger(cfg.Log.Level)
	validator.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	shutdownTracer, err := tracing.InitTracer(
		context.Background(),
		"chat-service",
		cfg.Tracing.Endpoint,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to init OpenTelemetry")
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			logrus.WithError(err).Error("Failed to shutdown tracer")
		}
	}()

	pg, err := postgres.New(cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database", err)
	}

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		logrus.Fatal("Failed to connect to rabbitmq", err)
	}

	chatService := service.NewService(pg, pg, rmq)
	chatHandler := httpHandlers.NewHandler(chatService)
	notificationHandler := httpHandlers.NewNotificationHandler(rmq)

	router := httpHandlers.NewRouter(chatHandler, notificationHandler, cfg.JWT.Secret)
	httpHandlers.InitMetrics(cfg.Metrics.Port)
	server := httpHandlers.NewServer(cfg.Server.Host, cfg.Server.Port, router)

	errChan := make(chan error)
	go func() {
		logrus.Infof("server address: %s", server.Addr())
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		logrus.WithError(err).Error("Failed to start server")
		if closeErr := pg.DB.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("Database close failed")
		}
		if closeErr := rmq.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("RabbitMQ close failed")
		}
		return
	case <-ctx.Done():
		logrus.Info("Server shutdown...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("HTTP server shutdown failed, forcing close")
		}

		if err := pg.DB.Close(); err != nil {
			logrus.WithError(err).Error("Database close failed")
		}
		if err := rmq.Close(); err != nil {
			logrus.WithError(err).Error("RabbitMQ close failed")
		}
	}

	logrus.Info("Service shutdown successfully")
}

func setupLogger(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
	logrus.AddHook(&tracing.OTelHook{})
}
