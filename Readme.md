# Collaborative Document Editor - Backend

A real-time collaborative document editing backend built with Go (Gin), PostgreSQL, Redis, and WebSockets. This backend powers a simplified Google Docs-like application with real-time editing capabilities.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Flutter Web   │    │   Go Backend    │    │   PostgreSQL    │
│  (Frontend)     │◄──►│     (Gin)       │◄──►│   Database      │
│                 │    │                 │    │                 │
│ - Access Token  │    │ - JWT Auth      │    │ - Users         │
│ - Refresh Token │    │ - REST APIs     │    │ - Documents     │
│ - WebSocket     │    │ - WebSocket     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │      Redis      │
                       │   (Caching)     │
                       │                 │
                       │ - Refresh       │
                       │   Tokens        │
                       └─────────────────┘
```

## 🚀 Features

- **User Authentication**: JWT-based authentication with access/refresh tokens
- **Document Management**: Create, read, update documents with CRUD operations
- **Real-time Collaboration**: WebSocket-based real-time editing with operational transforms
- **Auto-save**: Automatic document saving every 2 seconds
- **Token Management**: Secure token storage and refresh mechanism
- **CORS Support**: Cross-origin resource sharing for web clients

## 🛠️ Tech Stack

- **Backend Framework**: Go with Gin web framework
- **Database**: PostgreSQL (document and user storage)
- **Caching**: Redis (refresh token storage)
- **Authentication**: JWT tokens
- **Real-time**: WebSocket connections
- **Deployment**: Render (Backend)

## 📊 Database Schema
This application uses **PostgreSQL** as the primary relational database, with a simple and normalized schema designed for a collaborative text editing experience.

### 📍 Users Table

Stores registered users of the platform, with support for local and third-party authentication providers (e.g., Google).

### 📄 Documents Table

Stores documents created by users. Each document belongs to a user and contains collaborative content stored in JSON format.

**Foreign Key Constraint:**
```sql
FOREIGN KEY (author_id) REFERENCES users(id)
ON DELETE CASCADE
ON UPDATE CASCADE
```

## 🔐 Authentication Flow

```
1. User Registration/Login
   ├── POST /users/register or /users/login
   ├── Server validates credentials
   ├── Generate Access Token (15 min) + Refresh Token
   ├── Store refresh token in Redis
   └── Return tokens to client

2. Token Usage
   ├── Access token in Authorization header
   ├── Protected routes validate JWT
   └── Auto-refresh on expiry

3. Token Refresh
   ├── GET /refresh with refresh token
   ├── Validate token against Redis
   ├── Generate new token pair
   └── Update Redis storage
```

## 📡 API Documentation

### Base URL
```
Production: https://collaborative-text-editor-server-l8lp.onrender.com
Development: http://localhost:8080
```

### Authentication Endpoints

#### Register User
```http
POST /users/register
Content-Type: application/json

{
    "name": "user1",
    "email": "user1@gmail.com",
    "password": "user123"
}
```

**Response (201 Created):**
```json
{
    "access_token": "your-access-token-jwt-signed",
    "refresh_token": "your-refresh-token",
    "user": {
        "id": "675b738e-3741-4493-92de-47b18574990b",
        "email": "user1@gmail.com",
        "name": "user1",
        "provider": "local",
        "created_at": "2025-07-16T10:38:45.923909708Z",
        "updated_at": "2025-07-16T10:38:45.923909928Z"
    }
}
```

**Error (400 Bad Request):**
```json
{
    "error": "email already exists"
}
```

#### Login User
```http
POST /users/login
Content-Type: application/json

{
    "email": "user1@gmail.com",
    "password": "user123"
}
```

**Response:** Same as registration

#### Refresh Token
```http
GET /refresh
refresh_token: your-refresh-token
```

**Response (200 OK):**
```json
{
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token"
}
```

#### Logout User
```http
POST /users/logout
Header token: your-access-token
```

**Response (200 OK):**
```json
{
    "success": "logout success"
}
```

### Document Endpoints

#### Create Document
```http
POST /documents
Header token: your-access-token
```

**Response (201 Created):**
```json
{
    "id": "43140e48-7bb1-4ec4-9bb7-7b0f62d926f5",
    "author": {
        "id": "3693a8d5-7501-49cc-a0ef-c8429af66db6",
        "name": "user"
    },
    "title": "Untitled Document",
    "content": [],
    "created_at": "2025-07-16T10:47:05.230370161Z",
    "updated_at": "2025-07-16T10:47:05.230370211Z"
}
```

#### Get User Documents
```http
GET /documents/me
Header token: your-access-token
```

**Response (200 OK):**
```json
[
    {
        "id": "5db0164e-0b90-4029-b29e-4853932134ba",
        "author": {
            "id": "3693a8d5-7501-49cc-a0ef-c8429af66db6",
            "name": "user"
        },
        "title": "Hosting title",
        "content": [
            {
                "insert": "hello how are you?\n\nHola!!!"
            },
            {
                "insert": "\n",
                "attributes": {
                    "blockquote": true
                }
            }
        ],
        "created_at": "2025-07-15T11:54:18.803869Z",
        "updated_at": "2025-07-16T09:45:46.552254Z"
    }
]
```

#### Get Document by ID
```http
GET /documents/{document-id}
Header token: your-access-token
```

**Response (200 OK):** Same structure as single document

#### Update Document Title
```http
PATCH /documents/{document-id}
Header token: your-access-token
Content-Type: application/json

{
    "title": "New Title"
}
```

**Response (200 OK):**
```json
{
    "new_title": "New Title",
    "success": "ok"
}
```

## 🔌 WebSocket Integration

### Connection
```javascript
const ws = new WebSocket('ws://https://collaborative-text-editor-server-l8lp.onrender.com/ws/{document-id}');
```

### Message Types

#### Join Room
```json
{
    "event": "join",
    "room": "document-id",
    "data": {}
}
```

#### Typing (Real-time Edits)
```json
{
    "event": "typing",
    "room": "document-id",
    "data": {
        "ops": [
            {"retain": 4},
            {"insert": "Hello "},
            {"delete": 1}
        ]
    }
}
```

#### Save Document
```json
{
    "event": "save",
    "room": "document-id",
    "data": [
        {"insert": "Final document content"}
    ]
}
```

### Server Broadcast
```json
{
    "event": "changes",
    "data": {
        "ops": [
            {"retain": 4},
            {"insert": "text"}
        ]
    }
}
```

## 🏃‍♂️ Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Git

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-username/collaborative-editor-backend.git
cd collaborative-editor-backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=collaborative_editor
DB_SSLMODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_SECRET=your-jwt-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key

PORT=8080
```

4. **Database Setup**
```bash
# Create database
createdb collaborative_editor

# Run migrations (if you have migration files)
go run migrate.go
```

5. **Run the application**
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## 🔧 Project Structure

```
.
├── controllers/          # HTTP request handlers
├── models/              # Database models
├── middleware/          # Authentication middleware
├── routes/              # Route definitions
├── ws/                  # WebSocket handlers
├── database/            # Database connection
├── utils/               # Utility functions
├── main.go              # Application entry point
├── go.mod               # Go module file
└── README.md
```

## 🧪 Testing

### Manual Testing with curl

#### Register a user
```bash
curl -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Create a document
```bash
curl -X POST http://localhost:8080/documents \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### WebSocket Testing
Use a WebSocket client like Postman or write a simple HTML page:
```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws/your-document-id');
        
        ws.onopen = function() {
            console.log('Connected to WebSocket');
            ws.send(JSON.stringify({
                event: 'join',
                room: 'your-document-id',
                data: {}
            }));
        };
        
        ws.onmessage = function(event) {
            console.log('Received:', JSON.parse(event.data));
        };
    </script>
</body>
</html>
```

## 🚀 Deployment

### Render Deployment

1. **Connect your GitHub repository to Render**
2. **Set environment variables in Render dashboard**
3. **Build Command:** `go build -o main main.go`
4. **Start Command:** `./main`

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Update documentation
- Ensure WebSocket connections are properly handled
- Validate all API inputs

## 📝 Common Issues & Solutions

### WebSocket Connection Issues
- Ensure CORS is properly configured
- Check if the document ID exists
- Verify JWT token is valid

### Database Connection Problems
- Check PostgreSQL service status
- Verify database credentials
- Ensure database exists

### Redis Connection Issues
- Check Redis service status
- Verify Redis host and port
- Check if Redis requires authentication

## 🔒 Security Considerations

- All passwords are hashed using bcrypt
- JWT tokens have short expiry times
- Refresh tokens are stored securely in Redis
- Input validation on all endpoints
- CORS configured for specific origins
- Rate limiting recommended for production

## 📚 API Error Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Internal Server Error |

## 🔄 Real-time Collaboration Flow

```
1. User opens document → WebSocket connection to /ws/{docId}
2. Send "join" event → Server adds client to room
3. User types → Send "typing" event with delta
4. Server broadcasts to all other clients in room
5. Every 2 seconds → Send "save" event to persist changes
6. Server saves to PostgreSQL → Updates document content
```

## 📊 Performance Considerations

- Connection pooling for database
- Redis for fast token lookups
- WebSocket connection management
- Automatic cleanup of disconnected clients
- Document autosave batching




**Happy Coding! 🚀**