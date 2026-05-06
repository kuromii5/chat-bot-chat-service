package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
	"github.com/kuromii5/chat-bot-chat-service/internal/app"
	apperrors "github.com/kuromii5/chat-bot-chat-service/internal/errors"
	"github.com/kuromii5/chat-bot-shared/tracing"
	"github.com/kuromii5/chat-bot-shared/validator"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

func main() {
	cfg := config.MustLoad()
	setupLogger(cfg.Log.Level)
	validator.Init()
	wrapper.RegisterErrors(apperrors.Registry)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.New(ctx, cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize app")
	}

	if err := a.Run(ctx); err != nil {
		logrus.WithError(err).Error("App error")
	} else {
		logrus.Info("Shutting down...")
	}

	a.Close(context.Background())
	logrus.Info("Service stopped")
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
