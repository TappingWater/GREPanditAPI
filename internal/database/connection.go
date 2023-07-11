package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBSecrets struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getDBCredentials() (DBSecrets, error) {
	if os.Getenv("APP_ENV") == "dev" {
		return DBSecrets{os.Getenv("DB_HOST"), os.Getenv("DB_PASSWORD")}, nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return DBSecrets{}, fmt.Errorf("unable to load SDK config, %v", err)
	}
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(os.Getenv("SECRET_NAME")),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return DBSecrets{}, err
	}
	var secretData DBSecrets
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return DBSecrets{}, err
	}
	return secretData, nil
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
		os.Getenv("DB_HOST"),
		secrets.Username,
		secrets.Password,
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
