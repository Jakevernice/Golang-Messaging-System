
# 💬 Messaging App Backend (Go + Postgres + Docker)

This is a backend messaging system built with **Golang**, **PostgreSQL**, and **Docker Compose** as part of an intern onboarding task. The project includes user authentication, direct and group messaging, and is designed for LLM integration and modular API expansion.

---

## 🚀 How to Run the Project

### 📦 Prerequisites
- Docker
- Docker Compose

### 🔧 Setup

1. **Clone the repo**
```bash
git clone https://github.com/jakevernice/Golang-Messaging-System.git
cd Golang-Messaging-System
```

2. **Copy environment config**
```bash
cp .env.example .env
```

3. **Start all services**
```bash
docker-compose up --build
```

4. **Run migrations manually**
```bash
docker-compose run --rm migrate
```

5. **Access API**
- Backend will be running at: `http://localhost:8080`
- Use Postman or curl to test routes

---

## 📚 API Endpoints

### 🔐 Auth
| Method | Endpoint       | Description                     |
|--------|----------------|---------------------------------|
| POST   | `/api/register`| Register a new user             |
| POST   | `/api/login`   | Login and receive JWT           |
| POST   | `/api/logout`  | Logout (placeholder)            |
| POST   | `/api/refresh` | Refresh access token            |

---

### 📩 Messaging

| Type       | Endpoint                         | Description                          |
|------------|----------------------------------|--------------------------------------|
| Direct     | POST `/api/message/send`         | Send message to another user         |
| Group      | POST `/api/message/send`         | Send message to a group              |
| Thread     | GET `/api/conversation/:user_id` | Recent messages from specific DM.    |
| DM Preview | GET `/api/messages`              | Recent messages from chat and DMs    |
| Group View | GET `/api/groups`                | List of Groups associated with user  |

---

### 🧠 Group LLM Summary (Planned)
| Method | Endpoint                | Description                                 |
|--------|-------------------------|---------------------------------------------|
| GET    | `/group/:id/summary`    | Returns LLM summary of group conversation   |

---

## 🧱 Database Schema (Simplified)

- `users(id, username, password)`
- `groups(id, group_name, creator_id)`
- `group_members(group_id, member_id, is_admin)`
- `messages(id, sender_id, receiver_id?, group_id?, content, created_at)`

> ✅ A CHECK constraint ensures `receiver_id` XOR `group_id` is present in messages.

---

## ✅ Features Implemented

### 🔧 Infrastructure
- [x] Dockerized app with Go + Postgres
- [x] Live reload via Air
- [x] Migrations managed with `golang-migrate`

### 🔐 Authentication
- [x] Register/Login with bcrypt + JWT
- [x] Refresh tokens
- [x] Server-side JWT blacklist support

### 💬 Messaging
- [x] Direct messages (DMs)
- [x] Group messages
- [x] Middleware-based group limits (max 25 members, 2 admins)
- [x] SQL CHECK constraint to validate message type
- [x] Conditional indexes for performance

### 📥 Message Retrieval
- [x] Fetch latest messages in a thread (chat or group)
- [x] Previews for DMs and groups (partially implemented, route naming pending)
- [x] Test SQL scripts and manual validation done

---

## 🔜 Features To Be Completed

- [ ] Route `/api/chats`: return 10 distinct recent DMs (with preview)
- [ ] Ensure `/api/conversation/:user_id` returns latest 10 messages correctly
- [ ] Route `/group/:group_id/summary`: LLM integration (OpenAI/Cohere)
- [ ] Final cleanup & code review with proper testing.

---

## 🧪 Manual SQL Testing Snippets

```sql
-- Send a DM
INSERT INTO messages (sender_id, receiver_id, content)
VALUES (7, 8, 'Hello user 8');

-- Send a group message
INSERT INTO messages (sender_id, group_id, content)
VALUES (7, 2, 'Hello group 2');

-- View latest 10 messages in group
SELECT * FROM messages WHERE group_id = 2 ORDER BY created_at DESC LIMIT 10;

-- Check members in a group
SELECT * FROM group_members WHERE group_id = 2;

-- Enforce admin count in app:
SELECT COUNT(*) FROM group_members WHERE group_id = 2 AND is_admin = true;
```

---

## 👨‍💻 Author

Built by Vishal, as part of intern onboarding.

---
