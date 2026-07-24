package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// ConnectPostgres establishes the connection pool to Postgres using
// DATABASE_URL. Call this once at startup, same as ConnectMongo was.
func ConnectPostgres() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("postgres config error: %v", err)
	}
	config.MaxConns = 10

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("postgres connect error: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("postgres ping error: %v", err)
	}

	Pool = pool
	log.Println("connected to Postgres database")
}