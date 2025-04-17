package database

import (
	"context"
	"fmt"
	"kaizen-hq/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}
