package api

import (
	"log"
	"net/http"
	"time"

	"messaging-system/pkg/db"

	"github.com/gin-gonic/gin"
)

// CreateGroupRequest defines the structure for creating a group
type CreateGroupRequest struct {
	GroupName string `json:"group_name" binding:"required"`
}

// AddMemberRequest defines the structure for adding a member to a group
type AddMemberRequest struct {
	MemberID int  `json:"member_id" binding:"required"`
	IsAdmin  bool `json:"is_admin,omitempty"`
}

// Group represents a group in the system
type Group struct {
	ID        int       `json:"id"`
	GroupName string    `json:"group_name"`
	CreatorID int       `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupMember represents a group member
type GroupMember struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	MemberID int       `json:"member_id"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

// CreateGroupHandler handles creating a new group
func CreateGroupHandler(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get creator ID from context
	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	creatorIDInt, ok := creatorID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Begin transaction
	tx, err := db.GetDB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}
	defer tx.Rollback()

	// Create the group
	var groupID int
	query := `INSERT INTO groups (group_name, creator_id) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRow(query, req.GroupName, int(creatorIDInt)).Scan(&groupID)
	if err != nil {
		log.Printf("Error creating group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Add creator as admin member
	query = `INSERT INTO group_members (group_id, member_id, is_admin) VALUES ($1, $2, true)`
	_, err = tx.Exec(query, groupID, int(creatorIDInt))
	if err != nil {
		log.Printf("Error adding creator as admin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group created successfully",
		"group_id": groupID,
	})
}

// AddMemberToGroupHandler handles adding a member to a group
func AddMemberToGroupHandler(c *gin.Context) {
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get requester ID from context
	requesterID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requesterIDInt, ok := requesterID.(float64)
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

	// Check if requester is an admin of the group
	var isAdmin bool
	query := `SELECT is_admin FROM group_members WHERE group_id = $1 AND member_id = $2`
	err := db.GetDB().QueryRow(query, groupID, int(requesterIDInt)).Scan(&isAdmin)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this group or do not have permission"})
		return
	}
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add members to the group"})
		return
	}

	// Check if user to be added exists
	var userExists bool
	query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err = db.GetDB().QueryRow(query, req.MemberID).Scan(&userExists)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	// Check if user is already a member
	var alreadyMember bool
	query = `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND member_id = $2)`
	err = db.GetDB().QueryRow(query, groupID, req.MemberID).Scan(&alreadyMember)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}
	if alreadyMember {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this group"})
		return
	}

	// Check group constraints before adding
	var memberCount int
	query = `SELECT COUNT(*) FROM group_members WHERE group_id = $1`
	err = db.GetDB().QueryRow(query, groupID).Scan(&memberCount)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check group constraints"})
		return
	}
	if memberCount >= 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group has reached maximum member limit of 25"})
		return
	}

	// If trying to add as admin, check admin count
	if req.IsAdmin {
		var adminCount int
		query = `SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND is_admin = true`
		err = db.GetDB().QueryRow(query, groupID).Scan(&adminCount)
		if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check admin constraints"})
			return
		}
		if adminCount >= 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group has reached maximum admin limit of 2"})
			return
		}
	}

	// Add the member
	query = `INSERT INTO group_members (group_id, member_id, is_admin) VALUES ($1, $2, $3)`
	_, err = db.GetDB().Exec(query, groupID, req.MemberID, req.IsAdmin)
	if err != nil {
		log.Printf("Error adding member to group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added to group successfully"})
}

// GetUserGroupsHandler retrieves all groups that a user is a member of
func GetUserGroupsHandler(c *gin.Context) {
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

	// Query to get all groups user is a member of
	query := `
		SELECT g.id, g.group_name, g.creator_id, g.created_at
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.member_id = $1
		ORDER BY g.created_at DESC`

	rows, err := db.GetDB().Query(query, int(userIDInt))
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve groups"})
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		err := rows.Scan(&group.ID, &group.GroupName, &group.CreatorID, &group.CreatedAt)
		if err != nil {
			log.Printf("Error scanning group: %v", err)
			continue
		}
		groups = append(groups, group)
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// GetGroupMembersHandler retrieves all members of a specific group
func GetGroupMembersHandler(c *gin.Context) {
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

	// Query to get all members of the group
	query = `
		SELECT gm.id, gm.group_id, gm.member_id, gm.is_admin, gm.joined_at, u.username
		FROM group_members gm
		INNER JOIN users u ON gm.member_id = u.id
		WHERE gm.group_id = $1
		ORDER BY gm.is_admin DESC, gm.joined_at ASC`

	rows, err := db.GetDB().Query(query, groupID)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group members"})
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var member GroupMember
		var username string
		err := rows.Scan(&member.ID, &member.GroupID, &member.MemberID, &member.IsAdmin, &member.JoinedAt, &username)
		if err != nil {
			log.Printf("Error scanning member: %v", err)
			continue
		}

		memberData := map[string]interface{}{
			"id":        member.ID,
			"group_id":  member.GroupID,
			"member_id": member.MemberID,
			"username":  username,
			"is_admin":  member.IsAdmin,
			"joined_at": member.JoinedAt,
		}
		members = append(members, memberData)
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}
