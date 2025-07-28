# Messaging System API Documentation

## Overview
This messaging system supports both peer-to-peer (DM) and group messaging with JWT-based authentication.

## Authentication
All endpoints require a valid JWT access token in the Authorization header:
```
Authorization: Bearer <your_access_token>
```

## Messaging Endpoints

### 1. Send Message
**POST** `/api/message/send`

Send a direct message or group message.

**Request Body:**
```json
// For Direct Message
{
    "receiver_id": 2,
    "content": "Hello, how are you?"
}

// For Group Message  
{
    "group_id": 1,
    "content": "Hello everyone!"
}
```

**Response:**
```json
{
    "message": "Direct message sent successfully"
}
```

**Notes:**
- Either `receiver_id` or `group_id` must be provided, but not both
- For group messages, sender must be a member of the group
- Group constraints are enforced (max 25 members, max 2 admins)

### 2. Get All Messages
**GET** `/api/messages`

Retrieve all messages for the authenticated user (both sent/received DMs and group messages).

**Response:**
```json
{
    "messages": [
        {
            "id": 1,
            "sender_id": 1,
            "receiver_id": 2,
            "group_id": null,
            "content": "Hello!",
            "created_at": "2025-01-01T12:00:00Z",
            "is_group": false
        },
        {
            "id": 2,
            "sender_id": 1,
            "group_id": 1,
            "receiver_id": null,
            "content": "Group message",
            "created_at": "2025-01-01T12:05:00Z",
            "is_group": true
        }
    ]
}
```

### 3. Get Conversation
**GET** `/api/conversation/:user_id`

Get direct message conversation between authenticated user and another user.

**Response:**
```json
{
    "messages": [
        {
            "id": 1,
            "sender_id": 1,
            "receiver_id": 2,
            "group_id": null,
            "content": "Hello!",
            "created_at": "2025-01-01T12:00:00Z",
            "is_group": false
        }
    ]
}
```

### 4. Get Group Messages
**GET** `/api/group/:group_id/messages`

Get all messages for a specific group (requires membership).

**Response:**
```json
{
    "messages": [
        {
            "id": 2,
            "sender_id": 1,
            "group_id": 1,
            "receiver_id": null,
            "content": "Group message",
            "created_at": "2025-01-01T12:05:00Z",
            "is_group": true
        }
    ]
}
```

## Group Management Endpoints

### 1. Create Group
**POST** `/api/group/create`

Create a new group (creator becomes admin automatically).

**Request Body:**
```json
{
    "group_name": "My Awesome Group"
}
```

**Response:**
```json
{
    "message": "Group created successfully",
    "group_id": 1
}
```

### 2. Add Member to Group
**POST** `/api/group/:group_id/add-member`

Add a member to a group (admin only).

**Request Body:**
```json
{
    "member_id": 3,
    "is_admin": false
}
```

**Response:**
```json
{
    "message": "Member added to group successfully"
}
```

### 3. Get User Groups
**GET** `/api/groups`

Get all groups the authenticated user is a member of.

**Response:**
```json
{
    "groups": [
        {
            "id": 1,
            "group_name": "My Awesome Group",
            "creator_id": 1,
            "created_at": "2025-01-01T11:00:00Z"
        }
    ]
}
```

### 4. Get Group Members
**GET** `/api/group/:group_id/members`

Get all members of a specific group (requires membership).

**Response:**
```json
{
    "members": [
        {
            "id": 1,
            "group_id": 1,
            "member_id": 1,
            "username": "john_doe",
            "is_admin": true,
            "joined_at": "2025-01-01T11:00:00Z"
        },
        {
            "id": 2,
            "group_id": 1,
            "member_id": 2,
            "username": "jane_smith",
            "is_admin": false,
            "joined_at": "2025-01-01T11:30:00Z"
        }
    ]
}
```

## Database Schema

### Messages Table
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    receiver_id INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Constraint: Either receiver_id OR group_id must be set, but not both
ALTER TABLE messages ADD CONSTRAINT messages_receiver_or_group_check 
    CHECK (
        (receiver_id IS NOT NULL AND group_id IS NULL) OR 
        (receiver_id IS NULL AND group_id IS NOT NULL)
    );

-- Conditional indexes for better performance
CREATE INDEX idx_messages_receiver_id ON messages (receiver_id) WHERE receiver_id IS NOT NULL;
CREATE INDEX idx_messages_group_id ON messages (group_id) WHERE group_id IS NOT NULL;
```

### Groups Table
```sql
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    group_name VARCHAR(100) NOT NULL,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Group Members Table
```sql
CREATE TABLE group_members (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id),
    member_id INTEGER NOT NULL REFERENCES users(id),
    is_admin BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (group_id, member_id)
);
```

## Error Handling

The API returns appropriate HTTP status codes:
- `200 OK` - Success
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required or invalid token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

## Business Rules

### Group Constraints
- Maximum 25 members per group
- Maximum 2 admins per group
- Only group members can send messages to the group
- Only group admins can add new members

### Messaging Rules
- Users cannot send messages to themselves
- Receiver must exist for direct messages
- Group must exist and sender must be a member for group messages
- All messages include timestamp and sender information
- Database constraint ensures either `receiver_id` OR `group_id` is set, but not both
- Message type (DM vs Group) is determined by which field is populated

## Example Usage

### 1. Send a Direct Message
```bash
curl -X POST http://localhost:8080/api/message/send \
  -H "Authorization: Bearer your_access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_id": 2,
    "content": "Hello there!"
  }'
```

### 2. Create a Group and Send a Message
```bash
# Create group
curl -X POST http://localhost:8080/api/group/create \
  -H "Authorization: Bearer your_access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Development Team"
  }'

# Add member to group
curl -X POST http://localhost:8080/api/group/1/add-member \
  -H "Authorization: Bearer your_access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "member_id": 3,
    "is_admin": false
  }'

# Send group message
curl -X POST http://localhost:8080/api/message/send \
  -H "Authorization: Bearer your_access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": 1,
    "content": "Welcome to the team!"
  }'
```
