# Sparrow

Sparrow is a Gin + GORM based backend service for a lightweight forum system.  
It provides user registration/login, post creation/query, like actions, and SSE notifications.

## Tech Stack

- Go `1.24`
- Gin (`github.com/gin-gonic/gin`)
- GORM + PostgreSQL
- JWT (`github.com/golang-jwt/jwt/v5`)
- Zap logger
- Docker / Docker Compose

## Features

- User registration and login (JWT issued on login)
- Role-based access control middleware
- Post creation and post query with optional content/edit loading
- Post like API
- SSE subscription endpoint for real-time notifications
- Auto migration on startup
- Partitioned `post_likes` table (hash partition by `post_id`, 64 partitions by default)

## Requirements

- Go `1.24+`
- PostgreSQL `15+` (or use Docker Compose)
- A valid `.env` file in project root

## Quick Start (Local)

1. Create env file:

```bash
cp example.env .env
```

2. Update `.env` values (especially `JWT_SIGNING_KEY`, minimum 32 chars).

3. Start PostgreSQL:

```bash
docker compose up -d db
```

4. Run service:

```bash
go run .
```

Service default address: `http://localhost:8025`

## Run with Docker Compose

```bash
docker compose up --build
```

- API exposed at `http://localhost:8025`
- PostgreSQL exposed at `localhost:5432`

## Environment Variables

| Name | Required | Description |
| --- | --- | --- |
| `PORT` | Yes | HTTP server port (default in example: `8025`) |
| `DB_HOST` | Yes | PostgreSQL host (`db` in Docker, usually `127.0.0.1` locally) |
| `DB_USER` | Yes | PostgreSQL username |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_NAME` | Yes | PostgreSQL database name |
| `DB_PORT` | Yes | PostgreSQL port |
| `JWT_SIGNING_KEY` | Yes | JWT signing key, must be at least 32 characters |
| `LOG_LEVEL` | Yes | Zap log level (for example: `debug`, `info`, `warn`, `error`) |

## API Overview

Base path: `/api/v1`

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `GET` | `/ping` | No | Health check, returns `"pong"` |
| `GET` | `/user/:id` | No | Get user by ID |
| `POST` | `/auth/register` | No | Register user |
| `POST` | `/auth/login` | No | Login and get JWT token |
| `GET` | `/posts/:id` | No | Get post by ID (supports query options) |
| `POST` | `/posts` | Bearer token + role | Create post |
| `POST` | `/posts/:id/like` | Bearer token + role | Like a post |
| `GET` | `/subscribe/notify` | Bearer token + role | SSE notifications |

Note: legacy endpoint `POST /posts/like` has been removed.

### Query Parameters for `GET /posts/:id`

- `includeContent` (`true/false`, default `true`)
- `includeEdits` (`true/false`, default `false`)
- `editsLimit` (`1-200`, default `20`, only effective when `includeEdits=true`)

### Protected Endpoints

Use header:

```http
Authorization: Bearer <token>
```

Roles allowed by middleware for protected endpoints: `Users`, `Admin`.

## Example Requests

Register:

```bash
curl -X POST http://localhost:8025/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "nickname":"jack",
    "realName":"Jack Sparrow",
    "email":"jack@example.com",
    "password":"strong-password",
    "birthday":946684800
  }'
```

Login:

```bash
curl -X POST http://localhost:8025/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email":"jack@example.com",
    "password":"strong-password"
  }'
```

Create post:

```bash
curl -X POST http://localhost:8025/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"content":"hello sparrow"}'
```

Like post:

```bash
curl -X POST http://localhost:8025/api/v1/posts/1/like \
  -H "Authorization: Bearer <token>"
```

Like response (`201 Created`, first time):

```json
{
  "postID": 1,
  "userID": 1001,
  "liked": true
}
```

Like response (`200 OK`, already liked):

```json
{
  "postID": 1,
  "userID": 1001,
  "liked": false,
  "message": "already liked"
}
```

Subscribe SSE:

```bash
curl -N http://localhost:8025/api/v1/subscribe/notify \
  -H "Authorization: Bearer <token>"
```

## Development

Build:

```bash
go build -o bin/sparrow
```

Format and vet:

```bash
go fmt ./...
go vet ./...
```

Run tests:

```bash
go test ./...
```

## Project Structure

```text
.
├── main.go
├── configs/
├── internal/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── router/
│   ├── service/
│   └── utils/
├── tests/
├── Dockerfile
└── docker-compose.yml
```

## Notes

- The app loads `.env` at startup; missing `.env` will fail fast.
- Logs are written to `logs/sparrow.log` and `logs/zap.log`.
- Inside Docker Compose, set `DB_HOST=db`. For local direct run, use your local DB host (commonly `127.0.0.1`).
