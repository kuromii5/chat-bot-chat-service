package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/kuromii5/chat-bot-chat-service/config"
)

type postgres struct {
	DB *sqlx.DB
}

func New(cfg config.DatabaseConfig) (*postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "pgx", DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &postgres{DB: db}, nil
}

func DSN(c config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetAllTags fetches all available tag names from the database.
// Used at startup to seed the BadgerDB tag cache.
func (r *postgres) GetAllTags(ctx context.Context) ([]string, error) {
	var names []string
	if err := r.DB.SelectContext(ctx, &names, getTagsQuery); err != nil {
		return nil, fmt.Errorf("get all tags: %w", err)
	}
	return names, nil
}
