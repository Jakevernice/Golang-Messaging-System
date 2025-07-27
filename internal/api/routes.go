package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes for the application
func SetupRoutes(router *gin.Engine) {
	// Public routes
	public := router.Group("/api")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)
		public.POST("/logout", LogoutHandler)
	}
}
