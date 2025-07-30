package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"messaging-system/internal/auth"
	"messaging-system/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// generateTokenID creates a unique ID for JWT tokens
func generateTokenID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// If we can't generate random bytes, use timestamp as fallback
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// getEnvDuration returns the duration from environment variable or default value
func getEnvDuration(key string, defaultDuration time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultDuration
	}

	// if env set in integer as seconds
	valueInt, err := strconv.ParseInt(valueStr, 10, 64)
	if err == nil {
		return time.Duration(valueInt) * time.Second
	}

	// if env set in string as duration (eg. "15m", "1h", etc.)
	valueDuration, err := time.ParseDuration(valueStr)
	if err == nil {
		return valueDuration
	}

	// Fall back to default if parsing fails
	log.Printf("Failed to parse duration from %s=%s, using default", key, valueStr)
	return defaultDuration
}

// RegisterRequest defines the structure for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	MobileNo string `json:"mobile_no" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest defines the structure for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterHandler handles user registration
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Insert user into database
	query := `INSERT INTO users (username, mobile_no, password) VALUES ($1, $2, $3)`
	_, err = db.GetDB().Exec(query, req.Username, req.MobileNo, string(hashedPassword))
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginHandler handles user login
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user by username
	var user struct {
		ID       int
		Username string
		Password string
	}
	query := `SELECT id, username, password FROM users WHERE username = $1`
	err := db.GetDB().QueryRow(query, req.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get token durations from environment or use defaults
	accessTokenDuration := getEnvDuration("ACCESS_TOKEN_DURATION", 15*time.Minute)
	refreshTokenDuration := getEnvDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour)

	// Generate a unique token ID (jti)
	tokenID := generateTokenID()

	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
		"type":     "access",
		"jti":      tokenID,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(refreshTokenDuration).Unix(),
		"type":     "refresh",
		"jti":      generateTokenID(), // Use a different ID for refresh token
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign tokens with secret key
	jwtSecret := os.Getenv("JWT_SECRET")
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})
}

// RefreshTokenRequest defines the structure for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenHandler handles the refresh token endpoint
func RefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate the refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verify token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verify token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		return
	}

	// Check if the token's jti is blacklisted
	if jti, ok := claims["jti"].(string); ok {
		if auth.GetTokenBlacklist().IsBlacklisted(jti) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
			return
		}

		// Blacklist the used refresh token to prevent replay attacks
		// Get the expiration time from the token
		var expiryTime time.Time
		if exp, ok := claims["exp"].(float64); ok {
			expiryTime = time.Unix(int64(exp), 0)
		} else {
			// If no expiry found, blacklist for 24 hours
			expiryTime = time.Now().Add(24 * time.Hour)
		}

		// Add the jti to the blacklist
		auth.GetTokenBlacklist().Add(jti, expiryTime)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing jti claim"})
		return
	}

	// Extract user details from claims
	userId, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Generate a new access token
	accessTokenDuration := getEnvDuration("ACCESS_TOKEN_DURATION", 15*time.Minute)
	accessClaims := jwt.MapClaims{
		"user_id":  int(userId),
		"username": username,
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
		"type":     "access",
		"jti":      generateTokenID(), // Add a unique token ID
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	// Sign the access token
	jwtSecret := os.Getenv("JWT_SECRET")
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessTokenString,
	})
}

// LogoutHandler handles user logout
func LogoutHandler(c *gin.Context) {
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Check if the header has the Bearer prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header must be in the format 'Bearer {token}'"})
		return
	}

	// Extract the token
	tokenString := parts[1]

	// Parse the token to get its expiration time and jti
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Get the token ID (jti)
	jti, ok := claims["jti"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token does not have a jti claim"})
		return
	}

	// Get expiration time
	var expiryTime time.Time
	if exp, ok := claims["exp"].(float64); ok {
		expiryTime = time.Unix(int64(exp), 0)
	} else {
		// If no expiry found, blacklist for 24 hours
		expiryTime = time.Now().Add(24 * time.Hour)
	}

	// Add the jti to the blacklist
	auth.GetTokenBlacklist().Add(jti, expiryTime)

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// MeHandler returns the current authenticated user's information
func MeHandler(c *gin.Context) {
	// Get the user_id from the context (set by the JWT middleware)
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Fetch user details from database
	var user struct {
		ID       int    `json:"user_id"`
		Username string `json:"username"`
	}

	query := `SELECT id, username FROM users WHERE id = $1`
	err := db.GetDB().QueryRow(query, userId).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
		return
	}

	c.JSON(http.StatusOK, user)
}
