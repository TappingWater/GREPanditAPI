package postgresdb

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

/**
* Checks for environment variables to be loaded correctly.
**/
func init() {
	// Load environment variables from .env file during development
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	// Load environment variables from the appropriate .env file
	envFile := fmt.Sprintf(".env.%s", appEnv)
	if err := godotenv.Load(envFile); err != nil {
		fmt.Printf("No %s file found\n", envFile)
	}
}

/**
* Connects to the POSTGreSQL Db instance and sets up a connection pool to be
* used. Requires environment variables to be passed correctly.
**/
func ConnectDB() (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	dbpool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return dbpool, nil
}
