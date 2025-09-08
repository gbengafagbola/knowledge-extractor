package db

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() {
    url := os.Getenv("DATABASE_URL")
    if url == "" {
        log.Fatal("DATABASE_URL is not set")
    }

    pool, err := pgxpool.New(context.Background(), url)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }

    // Test connection
    err = pool.Ping(context.Background())
    if err != nil {
        log.Fatalf("Unable to ping database: %v", err)
    }

    Pool = pool
    fmt.Println(" Connected to Postgres!")
}
