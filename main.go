package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lestrrat-go/jwx/jwk"

	"grepandit.com/api/internal/database"
	"grepandit.com/api/internal/handlers"
	customMiddleware "grepandit.com/api/internal/middleware"
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
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	database.Migrate(db)

	autoRefresh := jwk.NewAutoRefresh(context.Background())
	// Configure the AutoRefresh  to refresh every 15 minutes
	jwksURL := "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_twk93JaQC/.well-known/jwks.json"
	autoRefresh.Configure(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))
	set, err := autoRefresh.Fetch(context.Background(), jwksURL)
	if err != nil {
		log.Fatalf("Failed to fetch JWK set: %v", err)
	}

	// Create services
	verbalQuestionService := services.NewVerbalQuestionService(db)
	wordService := services.NewWordService(db)
	userService := services.NewUserService(db)
	userVerbalStatsService := services.NewUserVerbalStatsService(db)

	// Create handlers
	verbalQuestionHandler := handlers.NewVerbalQuestionHandler(verbalQuestionService)
	wordHandler := handlers.NewWordHandler(wordService)
	userHandler := handlers.NewUserHandler(userService)
	userVerbalStatsHandler := handlers.NewUserVerbalStatHandler(userVerbalStatsService)

	// Start the Echo server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(customMiddleware.JWTAuthMiddleware(set))

	// Register routes
	registerRoutes(e, verbalQuestionHandler, wordHandler, userHandler, userVerbalStatsHandler)

	// Start the server
	port := "8080"
	fmt.Printf("Starting server on port %s\n", port)
	e.Start(fmt.Sprintf(":%s", port))
}

func registerRoutes(e *echo.Echo,
	verbalQuestionHandler *handlers.VerbalQuestionHandler,
	wordHandler *handlers.WordHandler,
	userHandler *handlers.UserHandler,
	userVerbalStatHandler *handlers.UserVerbalStatHandler) {
	// VerbalQuestion routes
	vqGroup := e.Group("/vbquestions")
	vqGroup.POST("", verbalQuestionHandler.Create)
	vqGroup.GET("/:id", verbalQuestionHandler.Get)
	vqGroup.GET("/adaptive", verbalQuestionHandler.GetAdaptiveQuestions)
	vqGroup.POST("/random", verbalQuestionHandler.GetRandomQuestions)
	vqGroup.GET("", verbalQuestionHandler.GetAll)

	// Word routes
	wGroup := e.Group("/words")
	wGroup.POST("", wordHandler.Create)
	wGroup.PATCH("/marked", wordHandler.MarkWords)
	wGroup.GET("/marked", wordHandler.GetMarkedWords)
	wGroup.GET("/:id", wordHandler.GetByID)
	wGroup.GET("/word/:word", wordHandler.GetByWord)

	// User routes
	uGroup := e.Group("/users")
	uGroup.POST("", userHandler.Create)
	uGroup.GET("", userHandler.Get)
	uGroup.POST("/marked-words", userHandler.AddMarkedWords)
	uGroup.POST("/marked-questions", userHandler.AddMarkedQuestions)
	uGroup.DELETE("/marked-words", userHandler.RemoveMarkedWords)
	uGroup.DELETE("/marked-questions", userHandler.RemoveMarkedQuestions)
	uGroup.GET("/marked-words", userHandler.GetMarkedWordsByUserToken)
	uGroup.GET("/marked-questions", userHandler.GetMarkedVerbalQuestionsByUserToken)
	uGroup.GET("/problematic-words", userHandler.GetProblematicWordsByUserToken)

	// UserVerbalStat routes
	uvsGroup := e.Group("/verbal-stats")
	uvsGroup.POST("", userVerbalStatHandler.Create)
	uvsGroup.GET("", userVerbalStatHandler.GetVerbalStatsByUserToken)
}
