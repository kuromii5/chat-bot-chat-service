package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
	badgercache "github.com/kuromii5/chat-bot-chat-service/internal/adapters/badger"
	outboxrelay "github.com/kuromii5/chat-bot-chat-service/internal/adapters/outbox"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/rabbitmq"
	tracingadapter "github.com/kuromii5/chat-bot-chat-service/internal/adapters/tracing"
	httpserver "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http"
	msghandler "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/msg"
	roomhandler "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/room"
	taghandler "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/tag"
	wshandler "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/ws"
	msgservice "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
	roomservice "github.com/kuromii5/chat-bot-chat-service/internal/service/room"
	tagservice "github.com/kuromii5/chat-bot-chat-service/internal/service/tag"
	tracingsvc "github.com/kuromii5/chat-bot-chat-service/internal/service/tracing"
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
		cfg.Tracing.Sampler,
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

	cache, err := badgercache.New()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open BadgerDB")
	}

	allTags, err := pg.GetAllTags(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to fetch tags for cache")
	}
	if err := cache.LoadTags(ctx, allTags); err != nil {
		logrus.WithError(err).Fatal("Failed to load tags into BadgerDB")
	}

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		logrus.Fatal("Failed to connect to rabbitmq", err)
	}

	tracingPG := tracingadapter.NewRepo(pg)

	relay := outboxrelay.NewRelay(tracingPG, rmq, rmq, 2*time.Second)
	go relay.Run(ctx)

	msgSvc := msgservice.NewService(tracingPG, tracingPG)
	tagSvc := tagservice.NewService(tracingPG, cache)
	roomSvc := roomservice.NewService(tracingPG, rmq)

	router := httpserver.NewRouter(
		msghandler.NewHandler(tracingsvc.NewMsgService(msgSvc)),
		taghandler.NewHandler(tracingsvc.NewTagService(tagSvc)),
		wshandler.NewHandler(rmq),
		roomhandler.NewHandler(tracingsvc.NewRoomService(roomSvc)),
		cfg.JWT.Secret,
	)

	httpserver.InitMetrics(ctx, cfg.Metrics.Port)
	server := httpserver.NewServer(cfg.Server.Host, cfg.Server.Port, router)

	errChan := make(chan error, 1)
	go func() {
		logrus.Infof("server address: %s", server.Addr())
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		logrus.WithError(err).Error("Failed to start server")
		if closeErr := cache.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("BadgerDB close failed")
		}
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

		if err := cache.Close(); err != nil {
			logrus.WithError(err).Error("BadgerDB close failed")
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
