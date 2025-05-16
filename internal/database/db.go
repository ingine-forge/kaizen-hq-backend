package database

import (
	"context"
	"fmt"
	"kaizen-hq/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	var db *pgxpool.Pool
	var err error

	maxAttempts := 10

	fmt.Println("Connecting to DB with URL:", cfg.DBURL)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = pgxpool.New(ctx, cfg.DBURL)
		if err == nil {
			// Now test the connection
			err = db.Ping(ctx)
			if err == nil {
				fmt.Println("✅ Connected to Postgres via pgxpool")
				return db, nil
			}
			// Close the pool if ping fails
			db.Close()
		}

		fmt.Printf("⏳ Attempt %d/%d: failed to connect to DB: %v\n", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)
}
