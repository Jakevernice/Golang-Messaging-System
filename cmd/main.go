package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	r := gin.Default()

	// Public routes
	public := r.Group("/api")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)
		public.POST("/logout", LogoutHandler)
	}

	// Start server
	r.Run(":" + os.Getenv("PORT"))
}

// --- Placeholder handler implementations below ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	MobileNo string `json:"mobile_no" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: Hash password, insert user into DB
	c.JSON(http.StatusCreated, gin.H{"message": "User registered"})
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: Validate credentials, issue JWT token
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": "<jwt_token>"})
}

func LogoutHandler(c *gin.Context) {
	// TODO: Invalidate JWT if using token blacklist (optional)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
