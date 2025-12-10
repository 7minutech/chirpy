# Chirpy

Chirpy is a small Twitter-style microblogging API built in Go.  
It's focused on clean HTTP handlers, JSON APIs, and a simple data layer – perfect for learning and demonstration.

## Features

- Create and manage "chirps" (short text posts, max 140 characters)
- User authentication with JWT tokens and refresh tokens
- User management (signup, login, update profile)
- Chirpy Red premium membership via webhooks
- Admin metrics and database reset for development
- JSON-based HTTP API

## Why This Project Matters

This project demonstrates how to:

- Structure a real-world Go web service
- Work with HTTP handlers, routing, and JSON
- Implement JWT authentication and refresh tokens
- Handle configuration, errors, and database persistence
- Build RESTful APIs with proper status codes
- Use Go modules and `go run` to build and run an app

## Requirements

- Go 1.20+ (recommended)
- PostgreSQL database

## How to Run

1. **Clone the repository**
```bash
   git clone https://github.com/7minutech/chirpy.git
   cd chirpy
```

2. **Configure environment**

   Create a `.env` file in the root directory:
```env
   DB_URL="postgresql://username:password@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   SECRET="your-jwt-secret-key"
   POLKA_KEY="your-webhook-api-key"
```

3. **Run the application**
```bash
   go run .
```

   The server will start on `http://localhost:8080`

---

# API Documentation

**Base URL:** `http://localhost:8080` (development)

**Content Type:** All requests and responses use `application/json`

## Table of Contents

- [Health and Metrics](#health-and-metrics)
- [Users](#users)
- [Authentication](#authentication)
- [Chirps](#chirps)
- [Admin](#admin)
- [Webhooks](#webhooks)

## Health and Metrics

### Get Health Status
Check if the server is running.

**Endpoint:** `GET /api/healthz`

**Response:** `200 OK`
```
OK
```

### Get App Page
Returns the main HTML application page.

**Endpoint:** `GET /app/`

**Response:** `200 OK`
- Content-Type: `text/html`

### Get Metrics
View metrics for visits to `/app/`.

**Endpoint:** `GET /admin/metrics`

**Response:** `200 OK`
- Content-Type: `text/html`
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 42 times!</p>
  </body>
</html>
```

## Users

### Create User
Register a new user account.

**Endpoint:** `POST /api/users`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T10:30:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

**Error Responses:**

`400 Bad Request` - Invalid input
```json
{
  "error": "could not decode request body"
}
```

`500 Internal Server Error` - Database error
```json
{
  "error": "could not create user"
}
```

### Update User
Update user email and/or password.

**Endpoint:** `PUT /api/users`

**Headers:**
```
Authorization: Bearer {Access Token}
```

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "password": "newsecurepassword456"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T11:45:00Z",
  "email": "newemail@example.com",
  "is_chirpy_red": false
}
```

**Error Responses:**

`401 Unauthorized` - Missing or invalid token
```json
{
  "error": "could not get token"
}
```

`400 Bad Request` - Invalid input
```json
{
  "error": "could not decode request body"
}
```

`500 Internal Server Error` - Database error
```json
{
  "error": "could not update user"
}
```

## Authentication

### Login
Authenticate a user and receive JWT access token and refresh token.

**Endpoint:** `POST /api/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T10:30:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}
```

**Token Details:**
- Access Token (JWT): Valid for 1 hour
- Refresh Token: Valid for 60 days

**Error Responses:**

`401 Unauthorized` - Invalid credentials
```json
{
  "error": "Incorrect email or password"
}
```

`500 Internal Server Error` - Server error
```json
{
  "error": "could not get user"
}
```

### Refresh Token
Obtain a new JWT access token using a refresh token.

**Endpoint:** `POST /api/refresh`

**Headers:**
```
Authorization: Bearer {Refresh Token}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**

`401 Unauthorized` - Invalid or expired refresh token
```json
{
  "error": "refresh token does not exist"
}
```

`401 Unauthorized` - Token has been revoked
```json
{
  "error": "refresh token is expired"
}
```

`400 Bad Request` - Missing token
```json
{
  "error": "Couldn't get refresh token from header"
}
```

### Revoke Token
Revoke a refresh token (logout).

**Endpoint:** `POST /api/revoke`

**Headers:**
```
Authorization: Bearer {Refresh Token}
```

**Response:** `204 No Content`

**Error Responses:**

`401 Unauthorized` - Invalid refresh token
```json
{
  "error": "refresh token does not exist"
}
```

`400 Bad Request` - Missing token
```json
{
  "error": "Couldn't get refresh token from header"
}
```

## Chirps

### Chirp Resource Structure
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T10:30:00Z",
  "body": "This is my chirp!",
  "user_id": "987e6543-e21b-12d3-a456-426614174000"
}
```

### Get All Chirps
Retrieve all chirps with optional filtering and sorting.

**Endpoint:** `GET /api/chirps`

**Query Parameters:**
- `author_id` (optional) - Filter by author's user ID (UUID format)
- `sort` (optional) - Sort order: `asc` (default) or `desc`
  - Sorts by `created_at` field

**Examples:**
- `GET /api/chirps` - All chirps, ascending order
- `GET /api/chirps?sort=desc` - All chirps, newest first
- `GET /api/chirps?author_id=123e4567-e89b-12d3-a456-426614174000` - Chirps by specific author
- `GET /api/chirps?author_id=123e4567-e89b-12d3-a456-426614174000&sort=desc` - Author's chirps, newest first

**Response:** `200 OK`
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "created_at": "2024-03-15T10:30:00Z",
    "updated_at": "2024-03-15T10:30:00Z",
    "body": "This is my first chirp!",
    "user_id": "987e6543-e21b-12d3-a456-426614174000"
  },
  {
    "id": "223e4567-e89b-12d3-a456-426614174001",
    "created_at": "2024-03-15T11:00:00Z",
    "updated_at": "2024-03-15T11:00:00Z",
    "body": "Another chirp here!",
    "user_id": "987e6543-e21b-12d3-a456-426614174000"
  }
]
```

**Error Responses:**

`400 Bad Request` - Invalid author_id format
```json
{
  "error": "could not parse author id"
}
```

`500 Internal Server Error` - Database error
```json
{
  "error": "could not select all chirps"
}
```

### Get Chirp by ID
Retrieve a specific chirp.

**Endpoint:** `GET /api/chirps/{chirpID}`

**Path Parameters:**
- `chirpID` - UUID of the chirp

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T10:30:00Z",
  "body": "This is my first chirp!",
  "user_id": "987e6543-e21b-12d3-a456-426614174000"
}
```

**Error Responses:**

`404 Not Found` - Chirp doesn't exist
```json
{
  "error": "chirp does not exist"
}
```

`400 Bad Request` - Invalid chirp ID format
```json
{
  "error": "could not parse chirpID from string to uuid"
}
```

### Create Chirp
Post a new chirp.

**Endpoint:** `POST /api/chirps`

**Headers:**
```
Authorization: Bearer {Access Token}
```

**Request Body:**
```json
{
  "body": "This is my chirp! Maximum 140 characters allowed."
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-03-15T10:30:00Z",
  "updated_at": "2024-03-15T10:30:00Z",
  "body": "This is my chirp! Maximum 140 characters allowed.",
  "user_id": "987e6543-e21b-12d3-a456-426614174000"
}
```

**Validation:**
- Body must not be empty
- Body must be 140 characters or less
- Profanity is automatically filtered

**Error Responses:**

`401 Unauthorized` - Missing or invalid token
```json
{
  "error": "token was not given in headers"
}
```

`400 Bad Request` - Chirp too long
```json
{
  "error": "Chirp is too long"
}
```

`500 Internal Server Error` - Database error
```json
{
  "error": "could not create chirp"
}
```

### Delete Chirp
Delete a chirp. Users can only delete their own chirps.

**Endpoint:** `DELETE /api/chirps/{chirpID}`

**Headers:**
```
Authorization: Bearer {Access Token}
```

**Path Parameters:**
- `chirpID` - UUID of the chirp to delete

**Response:** `204 No Content`

**Error Responses:**

`401 Unauthorized` - Missing or invalid token
```json
{
  "error": "token was not given in headers"
}
```

`403 Forbidden` - User doesn't own this chirp
```json
{
  "error": "user is not creator of chirp"
}
```

`404 Not Found` - Chirp doesn't exist
```json
{
  "error": "could not get chirp"
}
```

`400 Bad Request` - Invalid chirp ID
```json
{
  "error": "could not parse chirp id"
}
```

## Admin

### Reset Database
Delete all users from the database. Only available in development environment.

**Endpoint:** `POST /admin/reset`

**Requirements:**
- `PLATFORM` environment variable must be set to `dev`

**Response:** `200 OK`

**Error Responses:**

`403 Forbidden` - Not in development mode
```json
{
  "error": "must be in dev platform"
}
```

`500 Internal Server Error` - Database error
```json
{
  "error": "could not delete users"
}
```

## Webhooks

### Polka Webhook
Handle payment webhook events from Polka. Upgrades users to Chirpy Red membership.

**Endpoint:** `POST /api/polka/webhooks`

**Headers:**
```
Authorization: ApiKey {API Key}
```

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

**Event Types:**
- `user.upgraded` - Upgrades user to Chirpy Red
- Other events are ignored (still return 204)

**Response:** `204 No Content`
- Returns 204 whether user was upgraded or event was ignored

**Error Responses:**

`401 Unauthorized` - Missing or invalid API key
```json
{
  "error": "invalid api key"
}
```

`400 Bad Request` - Missing required fields
```json
{
  "error": "could not decode requeset body"
}
```

`404 Not Found` - User doesn't exist
```json
{
  "error": "could not find user"
}
```

## Error Codes Summary

| Status Code | Description |
|-------------|-------------|
| `200 OK` | Request succeeded |
| `201 Created` | Resource created successfully |
| `204 No Content` | Request succeeded with no response body |
| `400 Bad Request` | Invalid request data or parameters |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | Authenticated but not authorized for this action |
| `404 Not Found` | Resource not found |
| `500 Internal Server Error` | Server or database error |

## Notes

- All timestamps are in ISO 8601 format (UTC)
- All UUIDs follow the standard UUID v4 format
- Profanity in chirp bodies is automatically filtered
- Refresh tokens are valid for 60 days from creation
- JWT access tokens expire after 1 hour
