package main

import (
	"log"
	"net/http" // Added for CORS middleware
	"os"
	"ultra-chat-backend/config"
	"ultra-chat-backend/handlers"
	"ultra-chat-backend/repositories"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 2. Load the .env file at the very beginning.
	err := godotenv.Load()
	if err != nil {
		// This is not a fatal error for production, where env vars are set directly.
		// For local development, it's a sign that the .env file is missing.
		log.Println("Warning: Could not load .env file. Using system environment variables.")
	}

	// The rest of your main function remains the same.
	// It will now correctly pick up the variables loaded from the .env file.
	ddbClient := config.ConnectDB()

	userRepo := repositories.NewUserRepository(ddbClient)
	summaryRepo := repositories.NewSummaryRepository(ddbClient)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	authHandler := handlers.NewAuthHandler(userRepo)
	summaryHandler := handlers.NewSummaryHandler(summaryRepo)

	// Auth Routes
	e.GET("/login", authHandler.Login)
	e.GET("/callback", authHandler.Callback)
	e.GET("/profile", authHandler.Profile)

	// Summary Routes
	e.POST("/create-summary", summaryHandler.CreateSummary)
	e.GET("/summarizer", summaryHandler.GetSummaries)
	e.PUT("/update-summary", summaryHandler.UpdateSummary)
	e.DELETE("/delete-summary", summaryHandler.DeleteSummary)
	e.GET("/is_authenticated", summaryHandler.IsAuthenticated)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5001"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(e.Start(":" + port))
}
