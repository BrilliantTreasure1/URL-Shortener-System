package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDatabase() (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		getEnv("PGUSER", "postgres"),
		getEnv("PGPASSWORD", "postgres"),
		getEnv("PGHOST", "localhost"),
		getEnv("PGPORT", "5432"),
		getEnv("PGDATABASE", "test_api"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	return db, nil
}

func getEnv(key string, fallback string) string {

	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}