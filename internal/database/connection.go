package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type DBSecrets struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     string `json:"port"`
	SSLMode  string `json:"sslmode"`
}

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

func getDBCredentials() (DBSecrets, error) {
	if os.Getenv("APP_ENV") == "prod" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
		}))
		svc := secretsmanager.New(sess)
		secretName := "rds!db-d7a1a557-6da2-4b05-8b7d-65ea491f3bd7" // Replace with your secret name
		input := &secretsmanager.GetSecretValueInput{
			SecretId: &secretName,
		}
		result, err := svc.GetSecretValue(input)
		if err != nil {
			return DBSecrets{}, err
		}
		var secretData DBSecrets
		err = json.Unmarshal([]byte(*result.SecretString), &secretData)
		if err != nil {
			return DBSecrets{}, err
		}
		return secretData, nil
	} else {
		return DBSecrets{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			Port:     os.Getenv("DB_PORT"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		}, nil
	}
}

/**
* Connects to the POSTGreSQL Db instance and sets up a connection pool to be
* used. Requires secrets to be accessible from AWS Secrets Manager.
**/
func ConnectDB() (*pgxpool.Pool, error) {
	secrets, err := getDBCredentials()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve DB secrets: %v", err)
	}
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		secrets.Host,
		secrets.Username,
		secrets.Password,
		secrets.DBName,
		secrets.Port,
		secrets.SSLMode,
	)
	dbpool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return dbpool, nil
}
