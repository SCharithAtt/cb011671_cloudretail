# Seller Service API Documentation

## Overview

The Seller Service handles seller-specific operations including authentication, product management (via ProductService GraphQL), and order management.

**Base URL:** `http://localhost:8081` (Development)  
**Production URL:** `https://44lkl1on22.execute-api.us-east-1.amazonaws.com`  
**Port:** 8081

## Architecture

- **Authentication:** AWS Cognito with custom seller group
- **Framework:** Go with Gin
- **Database:** Via ProductService (DynamoDB) and OrderService (PostgreSQL)
- **Authorization:** JWT validation via Cognito JWKS

---

## Endpoints

### Authentication

#### 1. Seller Login

Authenticate seller and receive JWT tokens.

**Endpoint:** `POST /sellerLogin`

**Request Body:**
```json
{
  "email": "seller@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/sellerLogin \
  -H "Content-Type: application/json" \
  -d '{"email":"seller@example.com","password":"SecurePass123!"}'
```

---

#### 2. Seller Registration

Register a new seller account.

**Endpoint:** `POST /sellerRegister`

**Request Body:**
```json
{
  "name": "John's Electronics",
  "email": "john@electronics.com",
  "password": "SecurePass123!"
}
```

**Response:** `201 Created`
```json
{
  "message": "Seller registered successfully",
  "userSub": "uuid-1234-5678-90ab-cdef"
}
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/sellerRegister \
  -H "Content-Type: application/json" \
  -d '{"name":"John Electronics","email":"john@electronics.com","password":"SecurePass123!"}'
```

---

### Product Management

#### 3. Add Product

Add a new product (requires authentication).

**Endpoint:** `POST /addProduct`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Request Body:**
```json
{
  "name": "Wireless Headphones",
  "price": 17997.0,
  "description": "Premium wireless headphones with noise cancellation",
  "stock": 100
}
```

**Response:** `201 Created`
```json
{
  "productId": "prod-uuid-1234"
}
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/addProduct \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"name":"Wireless Headphones","price":17997.0,"description":"Premium headphones","stock":100}'
```

**Note:** Prices are in LKR (Sri Lankan Rupees). Example: 59.99 USD × 300 = 17997 LKR

---

#### 4. Edit Product

Update an existing product (requires authentication).

**Endpoint:** `PUT /products/:productId`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Request Body:**
```json
{
  "price": 14997.0,
  "stock": 150
}
```

**Response:** `200 OK`
```json
{
  "message": "Product updated successfully"
}
```

---

#### 5. Delete Product

Delete a product (requires authentication).

**Endpoint:** `DELETE /products/:productId`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Response:** `200 OK`
```json
{
  "message": "Product deleted successfully"
}
```

---

#### 6. Get Seller Products

Retrieve all products for authenticated seller.

**Endpoint:** `GET /products`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Response:** `200 OK`
```json
{
  "products": [
    {
      "productId": "prod-001",
      "name": "Wireless Headphones",
      "price": 17997.0,
      "stock": 100,
      "sellerId": "seller-uuid"
    }
  ]
}
```

---

### Order Management

#### 7. Get Seller Orders

Retrieve orders for seller's products.

**Endpoint:** `GET /orders`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Response:** `200 OK`
```json
{
  "orders": [
    {
      "orderId": "order-001",
      "buyerId": "buyer-uuid",
      "items": [
        {
          "productId": "prod-001",
          "quantity": 2
        }
      ],
      "totalAmount": 35994.0,
      "status": "pending",
      "createdAt": "2026-02-07T10:30:00Z"
    }
  ]
}
```

---

#### 8. Update Order Status

Update the status of an order.

**Endpoint:** `PUT /orders/:orderId`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Request Body:**
```json
{
  "status": "shipped"
}
```

**Valid Status Values:**
- `pending` - Order placed
- `processing` - Order being prepared
- `shipped` - Order shipped
- `delivered` - Order delivered
- `cancelled` - Order cancelled

**Response:** `200 OK`
```json
{
  "message": "Order status updated successfully"
}
```

---

## Authentication

### JWT Token Structure

**Header:**
```json
{
  "Authorization": "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Token Claims:**
```json
{
  "sub": "seller-uuid",
  "email": "seller@example.com",
  "custom:role": "seller",
  "cognito:groups": ["sellers"],
  "exp": 1707324000
}
```

---

## Error Responses

### Error Format

```json
{
  "error": "Error message description"
}
```

### Common HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Resource created |
| 400 | Bad request (invalid input) |
| 401 | Unauthorized (invalid/missing token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Resource not found |
| 500 | Internal server error |

---

## Environment Variables

```bash
# Cognito Configuration
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p
COGNITO_CLIENT_ID=2tkqjdk1i7r7uefcsargsrb3tq
COGNITO_CLIENT_SECRET=your_client_secret

# Service URLs
PRODUCT_SERVICE_URL=http://product-service.cloudretail.local:8082
ORDER_SERVICE_URL=http://order-service.cloudretail.local:8083

# Server Configuration
PORT=8081
AWS_REGION=us-east-1
```

---

## Testing

```bash
# Start service
cd backend/services/seller_service
go run main.go

# Test seller registration
curl -X POST http://localhost:8081/sellerRegister \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Seller","email":"test@seller.com","password":"Test123!"}'

# Test seller login
curl -X POST http://localhost:8081/sellerLogin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@seller.com","password":"Test123!"}'

# Run unit tests
go test -v
```

---

## Integration

### ProductService GraphQL Client

The Seller Service uses GraphQL to communicate with ProductService:

```graphql
mutation AddProduct($input: AddProductInput!) {
  addProduct(input: $input) {
    productId
  }
}
```

### OrderService REST Client

HTTP calls to OrderService for order management:

```
GET /api/orders?sellerId=<seller_uuid>
PUT /api/orders/<order_id>
```
