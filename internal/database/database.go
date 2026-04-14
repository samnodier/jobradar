package database

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	Db      *sql.DB
	Queries *Queries
}

func NewClient(dbURL string) (*Client, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}
	db.Ping()
	client := Client{
		Db:      db,
		Queries: New(db),
	}
	return &client, nil
}
