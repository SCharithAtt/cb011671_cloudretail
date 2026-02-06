# User Service API Documentation

## Overview

The User Service handles user authentication using AWS Cognito with OAuth2/OIDC. It provides endpoints for login, callback handling, and user management.

**Base URL:** `http://localhost:8080` (Development)  
**Production URL:** `https://44lkl1on22.execute-api.us-east-1.amazonaws.com`  
**Port:** 8080

## Architecture

- **Authentication:** AWS Cognito OAuth2/OIDC
- **Framework:** Go with net/http
- **Authorization:** JWT tokens (ID token, Access token, Refresh token)

---

## Endpoints

### 1. Initiate Login

Redirects user to AWS Cognito Hosted UI for authentication.

**Endpoint:** `GET /login`

**Query Parameters:** None

**Response:**
- **302 Found** - Redirects to Cognito login page

**Example Request:**
```bash
curl -I https://44lkl1on22.execute-api.us-east-1.amazonaws.com/login
```

**Flow:**
1. User clicks "Login"
2. Redirected to Cognito Hosted UI
3. User enters credentials
4. Cognito redirects to `/callback` with authorization code

---

### 2. OAuth Callback

Handles the OAuth2 callback from Cognito, exchanges authorization code for tokens.

**Endpoint:** `GET /callback`

**Query Parameters:**
- `code` (required) - Authorization code from Cognito
- `state` (optional) - CSRF protection state parameter

**Response:**
- **302 Found** - Redirects to frontend with tokens in URL fragment
- **400 Bad Request** - Missing or invalid authorization code
- **500 Internal Server Error** - Token exchange failed

**Example Success Response:**
```
HTTP/1.1 302 Found
Location: https://main.d1zj0qo7tbzc0o.amplifyapp.com/#id_token=eyJ...&access_token=eyJ...
```

**Token Response Fields:**
- `id_token` - JWT containing user identity claims (email, sub, name)
- `access_token` - Access token for API authorization
- `refresh_token` - Token to obtain new access tokens

**ID Token Claims:**
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "email_verified": true,
  "cognito:username": "user@example.com",
  "exp": 1707324000
}
```

---

### 3. Health Check

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "user-service"
}
```

---

## Authentication Flow

```
┌─────┐                                  ┌─────────┐
│User │                                  │ Cognito │
└──┬──┘                                  └────┬────┘
   │ 1. GET /login                            │
   ├─────────────────────────────────────────►│
   │                                           │
   │ 2. 302 → Cognito Hosted UI               │
   ◄─────────────────────────────────────────┤
   │                                           │
   │ 3. User enters credentials                │
   ├─────────────────────────────────────────►│
   │                                           │
   │ 4. 302 → /callback?code=xxx               │
   ◄─────────────────────────────────────────┤
   │                                           │
   │ 5. Exchange code for tokens               │
   ├─────────────────────────────────────────►│
   │                                           │
   │ 6. Returns JWT tokens                     │
   ◄─────────────────────────────────────────┤
   │                                           │
   │ 7. 302 → Frontend with tokens             │
   │                                           │
```

---

## Environment Variables

```bash
# Required
COGNITO_CLIENT_ID=2tkqjdk1i7r7uefcsargsrb3tq
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p

# Optional
COGNITO_CLIENT_SECRET=your_client_secret
COGNITO_REDIRECT_URL=https://44lkl1on22.execute-api.us-east-1.amazonaws.com/callback
COGNITO_LOGOUT_URL=https://main.d1zj0qo7tbzc0o.amplifyapp.com
FRONTEND_URL=https://main.d1zj0qo7tbzc0o.amplifyapp.com
PORT=8080
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "error_code",
  "error_description": "Human readable error message"
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `invalid_request` | Missing required parameter | 400 |
| `invalid_grant` | Invalid or expired authorization code | 401 |
| `server_error` | Internal server error | 500 |

---

## Security Considerations

1. **HTTPS Only:** Always use HTTPS in production
2. **State Parameter:** CSRF protection via state parameter
3. **Token Storage:** Store tokens securely (httpOnly cookies recommended)
4. **Token Expiration:** ID tokens expire after 1 hour
5. **Refresh Tokens:** Use refresh tokens to obtain new access tokens

---

## Testing

```bash
# Start service
cd backend/services/user_service
go run main.go

# Test login flow
curl -I http://localhost:8080/login

# Run unit tests
go test -v
```

---

## Integration with Other Services

- **Seller Service:** Validates JWT tokens from same Cognito User Pool
- **Product Service:** Validates JWT tokens for authenticated mutations
- **Order Service:** Validates JWT tokens for order operations
