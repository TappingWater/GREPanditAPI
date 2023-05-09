package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	postgresdb "grepandit.com/api/internal/database"
	"grepandit.com/api/internal/handlers"
	"grepandit.com/api/internal/migrations"
	"grepandit.com/api/internal/services"
)

/**
* To run your application in production mode, you should build a binary and
* run it with the APP_ENV environment variable set to production. Make sure
* to set the appropriate production environment variables in your production
* environment.
* Create a .env.prod file  and env.dev file and configure required variables:
* APP_ENV, DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
* To run in development:
* APP_ENV=development go run cmd/server/main.go
* To run in production:
* # Build the binary
* go build -o appName cmd/server/main.go
* # Run the binary in production mode
* APP_ENV=prod ./appName
**/
func main() {
	// Connect to the database
	dbpool, err := postgresdb.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbpool.Close()
	migrations.Migrate(dbpool)

	questionService := services.NewVerbalQuestionService(dbpool)
	questionHandler := handlers.NewVerbalQuestionHandler(questionService)

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/questions", questionHandler.Create)
	e.GET("/questions/:id", questionHandler.Get)
	e.PUT("/questions/:id", questionHandler.Update)
	e.DELETE("/questions/:id", questionHandler.Delete)

	// Register your routes and handlers here

	// Start the server
	port := "8080"
	fmt.Printf("Starting server on port %s\n", port)
	e.Start(fmt.Sprintf(":%s", port))
}
