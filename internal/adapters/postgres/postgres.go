package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kuromii5/chat-bot-chat-service/config"
)

type Postgres struct {
	DB       *sqlx.DB
	tagCache map[string]struct{}
	tagMu    sync.RWMutex
}

func New(cfg config.DatabaseConfig) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "pgx", DSN(cfg))
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	pg := &Postgres{DB: db, tagCache: make(map[string]struct{}), tagMu: sync.RWMutex{}}
	if err := pg.initTagCache(ctx); err != nil {
		return nil, err
	}

	return pg, nil
}

func DSN(c config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func (r *Postgres) initTagCache(ctx context.Context) error {
	var names []string
	if err := r.DB.SelectContext(ctx, &names, getTagsQuery); err != nil {
		return err
	}

	r.tagCache = make(map[string]struct{}, len(names))
	for _, name := range names {
		r.tagCache[name] = struct{}{}
	}

	return nil
}
