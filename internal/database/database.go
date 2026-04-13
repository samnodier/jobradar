package database

import "database/sql"

type Client struct {
	db *sql.DB
}

func Connect(dsn string) (Client, error) {

}
