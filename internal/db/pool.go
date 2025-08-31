package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustPool(connStr string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic(err)
	}
	return pool
}
