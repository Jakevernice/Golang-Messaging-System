package api

import (
	"messaging-system/internal/middleware"

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
		public.POST("/refresh", RefreshTokenHandler)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/me", MeHandler)

		// Messaging endpoints
		protected.POST("/message/send", SendMessageHandler)
		protected.GET("/messages", GetMessagesHandler)
		protected.GET("/conversation/:user_id", GetConversationHandler)
		protected.GET("/group/:group_id/messages", GetGroupMessagesHandler)

		// Group management endpoints
		protected.POST("/group/create", CreateGroupHandler)
		protected.POST("/group/:group_id/add-member", AddMemberToGroupHandler)
		protected.GET("/groups", GetUserGroupsHandler)
		protected.GET("/group/:group_id/members", GetGroupMembersHandler)
	}
}
