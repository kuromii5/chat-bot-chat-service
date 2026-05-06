package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
	badgercache "github.com/kuromii5/chat-bot-chat-service/internal/adapters/badger"
	kafkaproducer "github.com/kuromii5/chat-bot-chat-service/internal/adapters/kafka"
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
	"github.com/kuromii5/chat-bot-shared/tracing"
)

type App struct {
	closer     Closer
	httpServer *httpserver.Server
	relay      *outboxrelay.Relay
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	var a App

	shutdownTracer, err := tracing.InitTracer(
		ctx,
		"chat-service",
		cfg.Tracing.Endpoint,
		cfg.Tracing.Sampler,
	)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %w", err)
	}
	a.closer.Add(shutdownTracer)

	pg, err := postgres.New(cfg.Database) //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	a.closer.Add(func(_ context.Context) error { return pg.DB.Close() })

	cache, err := badgercache.New()
	if err != nil {
		return nil, fmt.Errorf("open BadgerDB: %w", err)
	}
	a.closer.Add(func(_ context.Context) error { return cache.Close() })

	allTags, err := pg.GetAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch tags for cache: %w", err)
	}
	if err := cache.LoadTags(ctx, allTags); err != nil {
		return nil, fmt.Errorf("load tags into BadgerDB: %w", err)
	}

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}
	a.closer.Add(func(_ context.Context) error { return rmq.Close() })

	kafkaProd := kafkaproducer.NewProducer(cfg.Kafka.Brokers)
	a.closer.Add(func(_ context.Context) error { return kafkaProd.Close() })

	tracingPG := tracingadapter.NewRepo(pg)
	tracingBroker := tracingadapter.NewBroker(rmq)
	tracingKafka := tracingadapter.NewKafkaProd(kafkaProd)

	a.relay = outboxrelay.NewRelay(
		tracingPG, tracingBroker, tracingBroker, tracingBroker, tracingKafka,
		2*time.Second,
	)

	msgSvc := msgservice.NewService(tracingPG, tracingPG)
	tagSvc := tagservice.NewService(tracingPG, cache)
	roomSvc := roomservice.NewService(tracingPG)

	router := httpserver.NewRouter(
		msghandler.NewHandler(tracingsvc.NewMsgService(msgSvc)),
		taghandler.NewHandler(tracingsvc.NewTagService(tagSvc)),
		wshandler.NewHandler(rmq),
		roomhandler.NewHandler(tracingsvc.NewRoomService(roomSvc)),
		cfg.JWT.Secret,
	)

	httpserver.InitMetrics(ctx, cfg.Metrics.Port)
	a.httpServer = httpserver.NewServer(cfg.Server.Host, cfg.Server.Port, router)
	a.closer.Add(a.httpServer.Shutdown)

	return &a, nil
}

func (a *App) Run(ctx context.Context) error {
	go a.relay.Run(ctx)

	errChan := make(chan error, 1)
	go func() {
		logrus.Infof("server address: %s", a.httpServer.Addr())
		if err := a.httpServer.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (a *App) Close(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	a.closer.Close(shutdownCtx)
}
