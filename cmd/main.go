package main

import (
	"log"
	"os"

	"messaging-system/internal/api"
	"messaging-system/internal/auth"
	"messaging-system/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database connection
	db.Initialize()
	defer db.Close()

	// Initialize token blacklist (will clean up every hour)
	_ = auth.GetTokenBlacklist()
	log.Println("Token blacklist initialized")

	// Setup Gin router
	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1"})

	// Setup routes
	api.SetupRoutes(router)

	// Start server
	router.Run(":" + os.Getenv("PORT"))
}
