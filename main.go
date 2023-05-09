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
	db, err := postgresdb.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	migrations.Migrate(db)

	// Create services
	verbalQuestionService := services.NewVerbalQuestionService(db)
	paragraphService := services.NewParagraphService(db)
	wordService := services.NewWordService(db)

	// Create handlers
	verbalQuestionHandler := handlers.NewVerbalQuestionHandler(verbalQuestionService)
	paragraphHandler := handlers.NewParagraphHandler(paragraphService)
	wordHandler := handlers.NewWordHandler(wordService)

	// Start the Echo server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	registerRoutes(e, verbalQuestionHandler, paragraphHandler, wordHandler)

	// Start the server
	port := "8080"
	fmt.Printf("Starting server on port %s\n", port)
	e.Start(fmt.Sprintf(":%s", port))
}

func registerRoutes(e *echo.Echo, verbalQuestionHandler *handlers.VerbalQuestionHandler, paragraphHandler *handlers.ParagraphHandler, wordHandler *handlers.WordHandler) {
	// VerbalQuestion routes
	vqGroup := e.Group("/vbquestion")
	vqGroup.POST("", verbalQuestionHandler.Create)
	vqGroup.GET("/:id", verbalQuestionHandler.Get)
	vqGroup.GET("/random", verbalQuestionHandler.GetRandomQuestions)
	vqGroup.GET("/count", verbalQuestionHandler.Count)

	// Paragraph routes
	pGroup := e.Group("/paragraph")
	pGroup.POST("", paragraphHandler.Create)
	pGroup.GET("/:id", paragraphHandler.Get)

	// Word routes
	wGroup := e.Group("/word")
	wGroup.POST("", wordHandler.Create)
	wGroup.GET("/:id", wordHandler.GetByID)
	wGroup.GET("/:word", wordHandler.GetByWord)
	wGroup.PUT("/:id", wordHandler.Update)
	wGroup.DELETE("/:id", wordHandler.Delete)
}
