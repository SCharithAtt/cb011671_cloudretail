# Order Service

A production-ready order management microservice for the CloudRetail e-commerce platform. Handles order creation, payment simulation, and order status management with PostgreSQL/RDS, EventBridge integration, and GraphQL queries to ProductService for stock validation.

## Features

- **Order Management** with PostgreSQL (GORM)
- **JWT Authentication** with AWS Cognito (JWKS validation)
- **GraphQL Integration** with ProductService for stock checks
- **EventBridge** event publishing on order creation
- **Payment Simulation** with checkbox redirect flow
- **Seller/Buyer Order Filtering** based on JWT role
- **Comprehensive Tests** with table-driven test patterns

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Order Service                      │
│                     (Port 8083)                     │
├─────────────────────────────────────────────────────┤
│  REST API (Gin)                                    │
│  ├─ Create Order (with stock check)               │
│  ├─ Payment Simulation                            │
│  ├─ Get Orders (buyer/seller filtering)           │
│  └─ Update Status (seller only)                   │
├─────────────────────────────────────────────────────┤
│  JWT Middleware (Cognito JWKS Validation)         │
├─────────────────────────────────────────────────────┤
│  PostgreSQL (GORM)      GraphQL Client            │
│  ├─ Orders table        └─ ProductService         │
│  └─ JSONB items             stock checks          │
├─────────────────────────────────────────────────────┤
│  EventBridge publisher (order-placed events)      │
└─────────────────────────────────────────────────────┘
```

## Database Schema

### Orders Table
```sql
CREATE TABLE orders (
  order_id UUID PRIMARY KEY,
  buyer_id VARCHAR NOT NULL,
  seller_id VARCHAR NOT NULL,
  items JSONB NOT NULL,  -- [{"productId": "...", "quantity": 2}]
  status VARCHAR NOT NULL DEFAULT 'pending',  -- pending|paid|shipped|delivered
  total_price DECIMAL NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

## Endpoints

### Public Endpoints

#### Health Check
```http
GET /health
```
Returns service health status.

#### Payment Simulation
```http
GET /simulatePayment/:orderId
```
Returns payment simulation page data.

**Response:**
```json
{
  "message": "Check box to mark paid",
  "orderId": "order-uuid",
  "totalPrice": 99.99,
  "status": "pending"
}
```

#### Mark Payment Done
```http
POST /markPaymentDone/:orderId
Content-Type: application/json

{
  "paid": true
}
```
Marks order as paid if checkbox confirmed.

**Response:**
```json
{
  "redirect": "/orderConfirmed/order-uuid"
}
```

#### Order Confirmed
```http
GET /orderConfirmed/:orderId
```
Returns confirmed order details.

### Protected Endpoints (Require JWT)

#### Create Order
```http
POST /createOrder
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "items": [
    {"productId": "prod-1", "quantity": 2},
    {"productId": "prod-2", "quantity": 1}
  ]
}
```

**Flow:**
1. Extract `buyerId` from JWT
2. For each item, query ProductService GraphQL: `getProductById(id)` → check stock
3. If stock insufficient: return 400 error
4. Calculate `totalPrice` based on ProductService prices
5. Create order in RDS with status "pending"
6. Fire EventBridge "order-placed" event
7. Return `orderId` and `paymentUrl`

**Response:**
```json
{
  "orderId": "order-uuid",
  "paymentUrl": "/simulatePayment/order-uuid"
}
```

#### Get Orders
```http
GET /getOrders?sellerId=<optional>
Authorization: Bearer <JWT>
```

**Filtering:**
- **Buyer role**: Returns orders where `buyer_id = JWT.sub`
- **Seller role**: Returns orders where `seller_id = JWT.sub` (or specified `sellerId` if matches)

**Response:**
```json
[
  {
    "orderId": "order-uuid",
    "buyerId": "buyer-id",
    "sellerId": "seller-id",
    "items": [{"productId": "prod-1", "quantity": 2}],
    "status": "paid",
    "totalPrice": 99.99,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
]
```

#### Update Order Status (Seller Only)
```http
PUT /updateStatus/:orderId
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "status": "shipped"  // shipped|delivered|cancelled
}
```

**Authorization:**
- Verifies `custom:role = "seller"` in JWT
- Verifies order's `seller_id` matches JWT `sub` (ownership check)

**Response:**
```json
{
  "message": "Order status updated successfully",
  "orderId": "order-uuid",
  "status": "shipped"
}
```

## Environment Variables

Create a `.env` file:

```env
# AWS Configuration
AWS_REGION=us-east-1

# Cognito Configuration
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p

# RDS Configuration
RDS_DSN=host=localhost user=postgres password=postgres dbname=cloudretail port=5432 sslmode=disable

# EventBridge Configuration
EVENTBRIDGE_BUS_ARN=arn:aws:events:us-east-1:111546515511:event-bus/cloud-retail-bus

# Service URLs
PRODUCT_GRAPHQL_URL=http://product-service:8082/graphql

# Server Configuration
PORT=8083
```

## Dependencies

- **GORM** v1.31.1 - PostgreSQL ORM
- **gorm.io/driver/postgres** v1.6.0 - PostgreSQL driver
- **hasura/go-graphql-client** v0.15.1 - GraphQL client
- **AWS SDK v2** - EventBridge
- **Gin** v1.11.0 - HTTP server
- **golang-jwt/jwt** v5 - JWT validation
- **google/uuid** v1.6.0 - UUID generation
- **godotenv** v1.5.1 - Environment loading

## Local Development

### Database Setup (PostgreSQL)
```bash
# Using Docker
docker run --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=cloudretail \
  -p 5432:5432 \
  -d postgres:15

# GORM auto-migrates the schema on startup
```

### Run the Service
```bash
cd backend/services/order_service
go run .
```

### Run Tests
```bash
go test -v
```

**Test Coverage:**
- ✅ Health endpoint
- ✅ Create order validation
- ✅ Mark payment validation
- ✅ Update status validation
- ✅ JWT middleware rejection
- ✅ JWT validation function
- ✅ Order JSONB marshal/unmarshal
- ✅ Order model structure
- ✅ EventBridge payload structure
- ✅ GraphQL query structure

**Total: 14 test suites, all passing**

### Build
```bash
go build .
```

## Docker Build

```bash
docker build -t order-service:latest .
```

## Kubernetes Deployment

### Deploy to EKS
```bash
# Build and push to ECR
docker build -t 111546515511.dkr.ecr.us-east-1.amazonaws.com/order-service:latest .
docker push 111546515511.dkr.ecr.us-east-1.amazonaws.com/order-service:latest

# Apply Kubernetes manifests
kubectl apply -f deployment.yaml
```

### Verify Deployment
```bash
kubectl get pods -l app=order-service
kubectl logs -l app=order-service
```

## Integration with Other Services

### ProductService (GraphQL)
OrderService queries ProductService before creating orders:

```graphql
query GetProduct($id: ID!) {
  getProductById(id: $id) {
    productId
    name
    price
    stock
    sellerId
  }
}
```

**Usage in CreateOrder:**
1. For each item in order, query ProductService
2. Check `stock >= quantity`
3. Use `price` to calculate `totalPrice`
4. Use `sellerId` as order's `seller_id`

### SellerService (REST)
SellerService calls OrderService REST endpoints:

```bash
# Get seller's orders
GET http://order-service:8083/getOrders
Authorization: Bearer <seller-jwt>

# Update order status
PUT http://order-service:8083/updateStatus/:orderId
Authorization: Bearer <seller-jwt>
{
  "status": "shipped"
}
```

### ProductService (EventBridge)
OrderService fires events that ProductService listens to:

**Event Structure:**
```json
{
  "Source": "order-service",
  "DetailType": "order-placed",
  "Detail": {
    "orderId": "order-uuid",
    "items": [
      {"productId": "prod-1", "quantity": 2}
    ]
  },
  "EventBusName": "arn:aws:events:us-east-1:111546515511:event-bus/cloud-retail-bus"
}
```

ProductService listener updates stock: `stock -= quantity`

## Authentication Flow

1. **User logs in** via user_service OR seller logs in via seller_service → Receives JWT
2. **User makes request** with `Authorization: Bearer <JWT>`
3. **Order service validates JWT**:
   - Extracts `kid` from token header
   - Fetches JWKS from Cognito (cached 1hr)
   - Verifies RSA signature
4. **Extracts claims**:
   - `sub` → `userId` (buyerId or sellerId)
   - `email` → userEmail
   - `custom:role` → customRole (seller|buyer)
5. **Authorization checks**:
   - `createOrder`: Any authenticated user (uses `sub` as buyerId)
   - `getOrders`: Filters by `sub` based on role
   - `updateStatus`: Requires `role=seller` AND order ownership

## Payment Simulation Flow

```
1. POST /createOrder → Creates order with status "pending"
   ↓
2. Returns paymentUrl: "/simulatePayment/{orderId}"
   ↓
3. GET /simulatePayment/{orderId} → Shows payment page
   ↓
4. Frontend displays checkbox: "I have paid"
   ↓
5. POST /markPaymentDone/{orderId} {"paid": true}
   ↓
6. Updates status to "paid"
   ↓
7. Returns redirect: "/orderConfirmed/{orderId}"
   ↓
8. GET /orderConfirmed/{orderId} → Shows order details
```

## Error Handling

| Status Code | Scenario |
|-------------|----------|
| 400 | Invalid request body, insufficient stock, invalid status |
| 401 | Missing or invalid JWT token |
| 403 | Seller trying to update order they don't own |
| 404 | Order not found |
| 500 | Database error, EventBridge error, GraphQL query failure |

## Production Considerations

1. **Database Connection Pooling**: GORM auto-configures connection pooling
2. **Transaction Management**: `CreateOrder` uses GORM transactions
3. **JWKS Caching**: 1-hour TTL reduces Cognito API calls
4. **EventBridge Reliability**: Async event publishing with error logging
5. **GraphQL Client**: Timeout and retry recommended (not implemented in starter)
6. **RDS IAM Authentication**: For production, use IAM DB auth instead of passwords
7. **Secret Management**: Use AWS Secrets Manager for RDS credentials

## Troubleshooting

### Database Connection Failed
```bash
# Check RDS_DSN format
echo $RDS_DSN

# Test connection
psql "$RDS_DSN"

# Verify GORM auto-migration logs
# Look for: "✅ Database connected and migrated"
```

### GraphQL Query Failed
```bash
# Verify ProductService is running
curl http://product-service:8082/health

# Test GraphQL query
curl -X POST http://product-service:8082/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { health }"}'
```

### EventBridge Events Not Firing
```bash
# Check IAM permissions for EventBridge PutEvents
aws events list-rules --region us-east-1

# Check service logs
kubectl logs -l app=order-service | grep EventBridge
```

### JWT Validation Fails
```bash
# Verify Cognito User Pool ID
echo $COGNITO_USER_POOL_ID

# Test JWKS endpoint
curl https://cognito-idp.us-east-1.amazonaws.com/us-east-1_eJvqfLh2p/.well-known/jwks.json
```

## API Examples

### Create Order (with stock check)
```bash
# Buyer creates order with valid JWT
curl -X POST http://localhost:8083/createOrder \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{
    "items": [
      {"productId": "prod-123", "quantity": 2}
    ]
  }'

# Response:
# {
#   "orderId": "order-uuid",
#   "paymentUrl": "/simulatePayment/order-uuid"
# }
```

### Simulate Payment
```bash
# Get payment page
curl http://localhost:8083/simulatePayment/order-uuid

# Mark payment done
curl -X POST http://localhost:8083/markPaymentDone/order-uuid \
  -H "Content-Type: application/json" \
  -d '{"paid": true}'

# Get confirmed order
curl http://localhost:8083/orderConfirmed/order-uuid
```

### Seller Updates Order Status
```bash
curl -X PUT http://localhost:8083/updateStatus/order-uuid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <seller-jwt>" \
  -d '{"status": "shipped"}'
```

## Service Communication

```
┌──────────────┐                   ┌─────────────────┐
│    Buyer     │  POST /createOrder│     Order       │
│  (UserService│ ─────────────────→│    Service      │
│   JWT)       │                   │    (8083)       │
└──────────────┘                   └─────────────────┘
                                           │
                       ┌───────────────────┼──────────────┐
                       │ GraphQL           │ EventBridge  │
                       ↓                   ↓              ↓
               ┌──────────────┐    ┌──────────────┐
               │   Product    │    │ EventBridge  │
               │   Service    │    │   (order-    │
               │   (8082)     │    │   placed)    │
               └──────────────┘    └──────────────┘
                   ↑ check stock       │
                   │                   ↓ listener
                   │           ┌──────────────┐
                   │           │   Product    │
                   │           │   Service    │
                   │           │ (stock-=qty) │
                   │           └──────────────┘
                   │
┌──────────────┐   │
│   Seller     │   │ GET/PUT orders
│  (SellerSvc  │ ──┤
│   JWT)       │   │
└──────────────┘   ↓
               ┌─────────────────┐
               │     Order       │
               │    Service      │
               │    (8083)       │
               └─────────────────┘
```

## Next Steps

- Implement GraphQL client retry/timeout
- Add order cancellation logic
- Implement refund flow
- Add order history pagination
- Implement webhook notifications for status updates
- Add metrics and observability (CloudWatch, Prometheus)
- Implement rate limiting for API endpoints
