package api

import (
	"log"
	"net/http"
	"time"

	"messaging-system/pkg/db"

	"github.com/gin-gonic/gin"
)

// SendMessageRequest defines the structure for sending messages
type SendMessageRequest struct {
	ReceiverID *int   `json:"receiver_id,omitempty"` // For DM
	GroupID    *int   `json:"group_id,omitempty"`    // For group message
	Content    string `json:"content" binding:"required"`
}

// Message represents a message in the system
type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID *int      `json:"receiver_id,omitempty"`
	GroupID    *int      `json:"group_id,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsGroup    bool      `json:"is_group"` // Computed field based on GroupID != nil
}

// SendMessageHandler handles sending messages (both DM and group)
func SendMessageHandler(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get sender ID from context (set by JWT middleware)
	senderID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert senderID to int (it comes as float64 from JWT claims)
	senderIDInt, ok := senderID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Validate that exactly one of receiver_id or group_id is provided
	if (req.ReceiverID == nil && req.GroupID == nil) || (req.ReceiverID != nil && req.GroupID != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either receiver_id or group_id must be provided, but not both"})
		return
	}

	// Handle DM (Direct Message)
	if req.ReceiverID != nil {
		if err := sendDirectMessage(int(senderIDInt), *req.ReceiverID, req.Content); err != nil {
			log.Printf("Error sending direct message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Direct message sent successfully"})
		return
	}

	// Handle Group Message
	if req.GroupID != nil {
		if err := sendGroupMessage(int(senderIDInt), *req.GroupID, req.Content); err != nil {
			log.Printf("Error sending group message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Group message sent successfully"})
		return
	}
}

// sendDirectMessage handles sending a direct message between two users
func sendDirectMessage(senderID, receiverID int, content string) error {
	// Validate that receiver exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := db.GetDB().QueryRow(query, receiverID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return &ValidationError{"Receiver user does not exist"}
	}

	// Prevent sending message to oneself
	if senderID == receiverID {
		return &ValidationError{"Cannot send message to yourself"}
	}

	// Insert the message
	query = `INSERT INTO messages (sender_id, receiver_id, content) VALUES ($1, $2, $3)`
	_, err = db.GetDB().Exec(query, senderID, receiverID, content)
	return err
}

// sendGroupMessage handles sending a message to a group
func sendGroupMessage(senderID, groupID int, content string) error {
	// Check if group exists
	var groupExists bool
	query := `SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1)`
	err := db.GetDB().QueryRow(query, groupID).Scan(&groupExists)
	if err != nil {
		return err
	}
	if !groupExists {
		return &ValidationError{"Group does not exist"}
	}

	// Check if sender is a member of the group
	var isMember bool
	query = `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND member_id = $2)`
	err = db.GetDB().QueryRow(query, groupID, senderID).Scan(&isMember)
	if err != nil {
		return err
	}
	if !isMember {
		return &ValidationError{"You are not a member of this group"}
	}

	// Check group constraints
	if err := validateGroupConstraints(groupID); err != nil {
		return err
	}

	// Insert the group message
	query = `INSERT INTO messages (sender_id, group_id, content) VALUES ($1, $2, $3)`
	_, err = db.GetDB().Exec(query, senderID, groupID, content)
	return err
}

// validateGroupConstraints checks if group meets the requirements
func validateGroupConstraints(groupID int) error {
	// Check member count (max 25 members)
	var memberCount int
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = $1`
	err := db.GetDB().QueryRow(query, groupID).Scan(&memberCount)
	if err != nil {
		return err
	}
	if memberCount > 25 {
		return &ValidationError{"Group has exceeded maximum member limit of 25"}
	}

	// Check admin count (max 2 admins)
	var adminCount int
	query = `SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND is_admin = true`
	err = db.GetDB().QueryRow(query, groupID).Scan(&adminCount)
	if err != nil {
		return err
	}
	if adminCount > 2 {
		return &ValidationError{"Group has exceeded maximum admin limit of 2"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// GetMessagesHandler retrieves messages for a user (both DMs and group messages)
func GetMessagesHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Query to get all messages where user is sender or receiver
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.group_id, m.content, m.created_at
		FROM messages m
		WHERE m.sender_id = $1 
		   OR m.receiver_id = $1 
		   OR (m.group_id IS NOT NULL AND m.group_id IN (
		       SELECT group_id FROM group_members WHERE member_id = $1
		   ))
		ORDER BY m.created_at DESC
		LIMIT 10`

	rows, err := db.GetDB().Query(query, int(userIDInt))
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.GroupID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		// Compute is_group field based on whether group_id is set
		msg.IsGroup = msg.GroupID != nil
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetConversationHandler retrieves messages between two users (DM conversation)
func GetConversationHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get other user ID from URL parameter
	otherUserID := c.Param("user_id")
	if otherUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID parameter is required"})
		return
	}

	// Query to get conversation between two users
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.group_id, m.content, m.created_at
		FROM messages m
		WHERE m.group_id IS NULL AND (
		    (m.sender_id = $1 AND m.receiver_id = $2) OR
		    (m.sender_id = $2 AND m.receiver_id = $1)
		)
		ORDER BY m.created_at ASC
		LIMIT 10`

	rows, err := db.GetDB().Query(query, int(userIDInt), otherUserID)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve conversation"})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.GroupID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		// For conversation, is_group is always false
		msg.IsGroup = false
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetGroupMessagesHandler retrieves messages for a specific group
func GetGroupMessagesHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get group ID from URL parameter
	groupID := c.Param("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID parameter is required"})
		return
	}

	// Check if user is a member of the group
	var isMember bool
	query := `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND member_id = $2)`
	err := db.GetDB().QueryRow(query, groupID, int(userIDInt)).Scan(&isMember)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify group membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this group"})
		return
	}

	// Query to get group messages
	query = `
		SELECT m.id, m.sender_id, m.receiver_id, m.group_id, m.content, m.created_at
		FROM messages m
		WHERE m.group_id = $1
		ORDER BY m.created_at DESC
		LIMIT 10`

	rows, err := db.GetDB().Query(query, groupID)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group messages"})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.GroupID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		// For group messages, is_group is always true
		msg.IsGroup = true
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
