# Order Service API Documentation

## Overview

The Order Service handles order creation, payment simulation, and order management using PostgreSQL with GORM. It integrates with ProductService for stock validation and publishes events to EventBridge.

**Base URL:** `http://localhost:8083` (Development)  
**Production URL:** `https://44lkl1on22.execute-api.us-east-1.amazonaws.com`  
**Port:** 8083

## Architecture

- **API Type:** REST
- **Framework:** Go with Gin
- **Database:** AWS RDS PostgreSQL with GORM
- **Events:** AWS EventBridge
- **Authorization:** JWT validation via Cognito JWKS

---

## Endpoints

### Order Management

#### 1. Create Order

Create a new order and process payment.

**Endpoint:** `POST /createOrder`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "userId": "user-uuid",
  "items": [
    {
      "productId": "prod-001",
      "quantity": 2
    },
    {
      "productId": "prod-002",
      "quantity": 1
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "orderId": "order-uuid-1234",
  "buyerId": "user-uuid",
  "sellerId": "seller-uuid",
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
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/createOrder \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-123","items":[{"productId":"prod-001","quantity":2}]}'
```

**Order Flow:**
1. Validates product availability via ProductService GraphQL
2. Calculates total amount (prices in LKR)
3. Simulates payment processing
4. Creates order in database
5. Publishes `OrderPlaced` event to EventBridge
6. Returns order details

---

#### 2. Get User Orders

Retrieve all orders for a specific user.

**Endpoint:** `GET /getOrders`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `userId` (required) - User UUID

**Response:** `200 OK`
```json
{
  "orders": [
    {
      "orderId": "order-001",
      "buyerId": "user-uuid",
      "sellerId": "seller-uuid",
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

**Example:**
```bash
curl "https://44lkl1on22.execute-api.us-east-1.amazonaws.com/getOrders?userId=user-123" \
  -H "Authorization: Bearer eyJhbGc..."
```

---

#### 3. Get All Orders

Admin endpoint to retrieve all orders.

**Endpoint:** `GET /api/orders`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>` (Admin role required)

**Query Parameters:**
- `limit` (optional) - Number of orders to return (default: 100)
- `offset` (optional) - Pagination offset (default: 0)
- `status` (optional) - Filter by status

**Response:** `200 OK`
```json
{
  "orders": [
    {
      "orderId": "order-001",
      "buyerId": "user-uuid",
      "sellerId": "seller-uuid",
      "items": [...],
      "totalAmount": 35994.0,
      "status": "pending",
      "createdAt": "2026-02-07T10:30:00Z"
    }
  ],
  "total": 150,
  "limit": 100,
  "offset": 0
}
```

---

#### 4. Get Order by ID

Retrieve details of a specific order.

**Endpoint:** `GET /api/orders/:orderId`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Response:** `200 OK`
```json
{
  "orderId": "order-001",
  "buyerId": "user-uuid",
  "sellerId": "seller-uuid",
  "items": [
    {
      "productId": "prod-001",
      "quantity": 2
    }
  ],
  "totalAmount": 35994.0,
  "status": "pending",
  "paymentStatus": "completed",
  "createdAt": "2026-02-07T10:30:00Z",
  "updatedAt": "2026-02-07T10:30:00Z"
}
```

---

#### 5. Update Order Status

Update the status of an order (seller/admin only).

**Endpoint:** `PUT /api/orders/:orderId`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "status": "shipped"
}
```

**Valid Status Values:**
- `pending` - Order placed, awaiting processing
- `processing` - Order is being prepared
- `shipped` - Order has been shipped
- `delivered` - Order delivered to customer
- `cancelled` - Order cancelled

**Response:** `200 OK`
```json
{
  "message": "Order status updated successfully",
  "order": {
    "orderId": "order-001",
    "status": "shipped",
    "updatedAt": "2026-02-07T11:00:00Z"
  }
}
```

**Example:**
```bash
curl -X PUT https://44lkl1on22.execute-api.us-east-1.amazonaws.com/api/orders/order-001 \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"status":"shipped"}'
```

---

### Health Check

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "order-service",
  "database": "connected"
}
```

---

## Authentication

### JWT Token Format

**Header:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Token Claims:**
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "custom:role": "buyer",
  "exp": 1707324000
}
```

---

## Error Responses

### Error Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### Common Error Codes

| HTTP Code | Error Code | Description |
|-----------|------------|-------------|
| 400 | `INVALID_INPUT` | Invalid request body or parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT token |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `ORDER_NOT_FOUND` | Order does not exist |
| 409 | `INSUFFICIENT_STOCK` | Product out of stock |
| 500 | `INTERNAL_ERROR` | Server error |

---

## Events

### OrderPlaced Event

Published to EventBridge when order is created.

**Source:** `order.service`  
**Detail Type:** `OrderPlaced`

**Event Detail:**
```json
{
  "orderId": "order-uuid-1234",
  "buyerId": "user-uuid",
  "sellerId": "seller-uuid",
  "items": [
    {
      "productId": "prod-001",
      "quantity": 2
    }
  ],
  "totalAmount": 35994.0,
  "timestamp": "2026-02-07T10:30:00Z"
}
```

**Consumer:** Stock updater Lambda function (automatically reduces product stock)

---

## Database Schema

### Orders Table (PostgreSQL)

```sql
CREATE TABLE orders (
  order_id UUID PRIMARY KEY,
  buyer_id VARCHAR(255) NOT NULL,
  seller_id VARCHAR(255) NOT NULL,
  items JSONB NOT NULL,
  total_amount DECIMAL(10, 2) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',
  payment_status VARCHAR(50) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes:**
- `idx_orders_buyer_id` on `buyer_id`
- `idx_orders_seller_id` on `seller_id`
- `idx_orders_status` on `status`
- `idx_orders_created_at` on `created_at`

---

## Environment Variables

```bash
# Database Configuration
DB_HOST=cloudretail-db.xxxx.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=cb011671_cr_pass_senura
DB_NAME=cloudretail
DB_SSL_MODE=require

# AWS Configuration
AWS_REGION=us-east-1
EVENTBRIDGE_BUS_ARN=arn:aws:events:us-east-1:xxx:event-bus/cloudretail-events

# Service URLs
PRODUCT_GRAPHQL_URL=http://product-service.cloudretail.local:8082/graphql

# Cognito Configuration
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p

# Server Configuration
PORT=8083
```

---

## Testing

```bash
# Start service
cd backend/services/order_service
go run main.go

# Test health endpoint
curl http://localhost:8083/health

# Create an order
curl -X POST http://localhost:8083/createOrder \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-123","items":[{"productId":"prod-001","quantity":1}]}'

# Get user orders
curl "http://localhost:8083/getOrders?userId=user-123" \
  -H "Authorization: Bearer <token>"

# Run unit tests
go test -v
```

---

## Integration

### ProductService GraphQL Integration

Order Service queries ProductService to validate stock and get prices:

```graphql
query GetProduct($id: ID!) {
  product(id: $id) {
    productId
    name
    price
    stock
    sellerId
  }
}
```

### EventBridge Integration

Publishes events for order lifecycle:

```go
eventBridgeClient.PutEvents(ctx, &eventbridge.PutEventsInput{
    Entries: []types.PutEventsRequestEntry{
        {
            Source:     aws.String("order.service"),
            DetailType: aws.String("OrderPlaced"),
            Detail:     aws.String(eventJSON),
            EventBusName: aws.String(eventBusArn),
        },
    },
})
```

---

## Payment Processing

### Payment Flow

1. **Validation:** Validate order exists and is in `pending` status
2. **Amount Verification:** Verify payment amount matches order total
3. **Payment Gateway:** Simulate payment processing (mock implementation)
4. **Update Status:** Update order and payment status
5. **Stock Update:** EventBridge triggers Lambda to reduce product stock

### Payment States

- `pending` - Payment not yet processed
- `processing` - Payment being processed
- `completed` - Payment successful
- `failed` - Payment failed
- `refunded` - Payment refunded

---

## Currency Note

All amounts are in **LKR (Sri Lankan Rupees)**.

Example order totals:
- 2 × Wireless Headphones (17,997 LKR each) = 35,994 LKR
- 1 × USB-C Charger (8,997 LKR) = 8,997 LKR

---

## Security Considerations

1. **Authentication:** All endpoints require valid JWT token
2. **Authorization:** Users can only access their own orders
3. **Input Validation:** All inputs validated and sanitized
4. **SQL Injection:** Protected by GORM parameter binding
5. **Rate Limiting:** Implement in production

---

## Future Enhancements

- [ ] Add order cancellation workflow
- [ ] Implement refund processing
- [ ] Add order tracking system
- [ ] Support multiple payment gateways
- [ ] Add order notifications (email/SMS)
- [ ] Implement order analytics dashboard
- [ ] Add invoice generation
- [ ] Support split payments
