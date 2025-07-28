package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"messaging-system/internal/auth"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates JWT tokens and sets user ID in the context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in the format 'Bearer {token}'"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			// Get the secret key from environment variable
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				return nil, errors.New("JWT_SECRET environment variable not set")
			}

			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check token type - only access tokens should be allowed
			tokenType, ok := claims["type"].(string)
			if !ok || tokenType != "access" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
				c.Abort()
				return
			}

			// Check if the token's jti is blacklisted
			if jti, ok := claims["jti"].(string); ok {
				if auth.GetTokenBlacklist().IsBlacklisted(jti) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing jti claim"})
				c.Abort()
				return
			}

			// Get the user ID from the claims
			userId, ok := claims["user_id"]
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing user_id claim"})
				c.Abort()
				return
			}

			// Set the user ID in the context
			c.Set("user_id", userId)

			// Continue to the next handler
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
