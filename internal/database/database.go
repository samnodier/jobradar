package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

func NewClient(dbURL string) (*Client, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}
	// db.Ping()
	// client := Client{
	// 	Db:      db,
	// 	Queries: New(db),
	// }
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Client{
		Pool:    pool,
		Queries: New(pool),
	}, nil
}
