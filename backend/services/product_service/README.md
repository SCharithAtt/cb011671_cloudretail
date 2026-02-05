# Product Service

A production-ready GraphQL microservice for managing products and reviews in the CloudRetail e-commerce platform.

## Features

- **GraphQL API** using gqlgen for type-safe schema-first development
- **JWT Authentication** with AWS Cognito (JWKS validation)
- **DynamoDB Integration** for Products and Reviews tables
- **EventBridge Listener** for real-time stock updates on order-placed events
- **Health Monitoring** endpoint for Kubernetes probes
- **Comprehensive Tests** with table-driven test patterns

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Product Service                    │
│                     (Port 8082)                     │
├─────────────────────────────────────────────────────┤
│  GraphQL API (gqlgen)                              │
│  ├─ Queries: getProductById, getAllProducts        │
│  └─ Mutations: addProduct, editProduct, addReview  │
├─────────────────────────────────────────────────────┤
│  JWT Middleware (Cognito JWKS Validation)         │
├─────────────────────────────────────────────────────┤
│  DynamoDB                EventBridge Listener      │
│  ├─ Products Table       ├─ Polls for order-placed │
│  └─ Reviews Table        └─ Updates stock on event │
└─────────────────────────────────────────────────────┘
```

## GraphQL Schema

### Queries
- `getProductById(id: ID!): Product` - Get single product by ID
- `getAllProducts(filter: ProductFilter): [Product!]!` - Get all products (optionally filtered by seller)
- `health: String!` - Health check

### Mutations (Require JWT)
- `addProduct(input: AddProductInput!): Product!` - Create new product (seller only)
- `editProduct(input: EditProductInput!): Product!` - Update product (ownership check)
- `addReview(input: AddReviewInput!): Review!` - Add product review

### Types
```graphql
type Product {
  productId: ID!
  name: String!
  price: Float!
  description: String
  stock: Int!
  sellerId: String!
  reviews: [Review!]!
  createdAt: String
  updatedAt: String
}

type Review {
  reviewId: ID!
  productId: String!
  text: String
  rating: Int
  userId: String
  createdAt: String
}
```

## Environment Variables

Create a `.env` file in the product_service directory:

```env
# AWS Configuration
AWS_REGION=us-east-1

# Cognito Configuration
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p

# DynamoDB Tables
PRODUCTS_TABLE=Products
REVIEWS_TABLE=Reviews

# EventBridge Configuration
EVENT_BUS_NAME=default

# Server Configuration
PORT=8082
```

## Dependencies

- **gqlgen** v0.17.86 - GraphQL server
- **AWS SDK v2** - DynamoDB, EventBridge
- **Gin** v1.11.0 - HTTP server
- **golang-jwt/jwt** v5 - JWT validation
- **godotenv** v1.5.1 - Environment loading

## Local Development

### Run the Service
```bash
cd backend/services/product_service
go run .
```

### Run Tests
```bash
go test -v
```

### Build
```bash
go build .
```

## Docker Build

```bash
docker build -t product-service:latest .
```

## Kubernetes Deployment

### Deploy to EKS
```bash
# Build and push to ECR
docker build -t 111546515511.dkr.ecr.us-east-1.amazonaws.com/product-service:latest .
docker push 111546515511.dkr.ecr.us-east-1.amazonaws.com/product-service:latest

# Apply Kubernetes manifests
kubectl apply -f deployment.yaml
```

### Verify Deployment
```bash
kubectl get pods -l app=product-service
kubectl logs -l app=product-service
```

## DynamoDB Setup

### Products Table
- **Primary Key**: `productId` (String)
- **Attributes**: name, price, description, stock, sellerId, createdAt, updatedAt

### Reviews Table
- **Primary Key**: `reviewId` (String)
- **Attributes**: productId, text, rating, userId, createdAt

### Create Tables (AWS CLI)
```bash
aws dynamodb create-table \
  --table-name Products \
  --attribute-definitions AttributeName=productId,AttributeType=S \
  --key-schema AttributeName=productId,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1

aws dynamodb create-table \
  --table-name Reviews \
  --attribute-definitions AttributeName=reviewId,AttributeType=S \
  --key-schema AttributeName=reviewId,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

## EventBridge Integration

The service listens for `order-placed` events to update product stock automatically:

**Event Structure:**
```json
{
  "detail": {
    "productId": "product-123",
    "quantity": 5
  }
}
```

**Stock Update Logic:**
1. Receive order-placed event
2. Get current product stock from DynamoDB
3. Check if stock >= quantity
4. If available: Update stock -= quantity
5. If insufficient: Log error

## Integration with Seller Service

The seller_service calls product_service GraphQL mutations:

```go
// Seller adds product
mutation {
  addProduct(input: {
    name: "New Product"
    price: 99.99
    description: "Product description"
    stock: 100
    sellerId: "seller-id-from-jwt"
  }) {
    productId
    name
  }
}

// Seller edits product
mutation {
  editProduct(input: {
    productId: "product-123"
    name: "Updated Name"
    price: 89.99
    stock: 50
  }) {
    productId
    updatedAt
  }
}
```

## Authentication Flow

1. **Seller logs in** via seller_service → Receives JWT from Cognito
2. **Seller makes request** with `Authorization: Bearer <JWT>`
3. **Product service validates JWT**:
   - Extracts `kid` from token header
   - Fetches JWKS from Cognito (cached 1hr)
   - Verifies RSA signature
4 **Extracts seller ID** from `sub` claim
5. **Authorization checks**:
   - `addProduct`: Verify sellerId in input matches JWT sub
   - `editProduct`: Verify product ownership (product.sellerId == JWT sub)

## Testing

### Test Coverage
- ✅ Health endpoint
- ✅ GraphQL authentication requirements
- ✅ JWT validation logic
- ✅ Query/mutation validation
- ✅ OrderPlacedEvent structure
- ✅ DynamoDB model structures
- ✅ JWKS cache logic
- ✅ Environment variables

**Total: 11 test suites, all passing**

### Run Specific Tests
```bash
go test -v -run TestHealthEndpoint
go test -v -run TestGraphQL
go test -v -run TestValidateJWT
```

## API Examples

### Query Product
```bash
curl -X POST http://localhost:8082/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { getProductById(id: \"123\") { productId name price stock } }"
  }'
```

### Add Product (Authenticated)
```bash
curl -X POST http://localhost:8082/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "query": "mutation { addProduct(input: { name: \"Test Product\", price: 49.99, stock: 100, sellerId: \"seller-123\" }) { productId name } }"
  }'
```

### GraphQL Playground
Visit `http://localhost:8082/playground` for interactive API exploration.

## Service Communication

```
┌──────────────┐     GraphQL      ┌─────────────────┐
│   Seller     │ ─────────────────→│    Product      │
│   Service    │   (addProduct,    │    Service      │
│   (8081)     │    editProduct)   │    (8082)       │
└──────────────┘                   └─────────────────┘
                                           │
                                           │ Read/Write
                                           ↓
                                   ┌─────────────────┐
                                   │    DynamoDB     │
                                   │  Products/      │
                                   │  Reviews        │
                                   └─────────────────┘
                                           ↑
                                           │ Stock Update
┌──────────────┐                   ┌─────────────────┐
│    Order     │  order-placed     │   EventBridge   │
│   Service    │ ─────────────────→│                 │
│   (8083)     │   event          └─────────────────┘
└──────────────┘
```

## Production Considerations

1. **IAM Permissions**: Service needs DynamoDB read/write and EventBridge receive permissions
2. **JWKS Caching**: 1-hour TTL reduces Cognito API calls
3. **Health Probes**: Kubernetes liveness/readiness use `/health`
4. **Resource Limits**: 256Mi-512Mi memory, 250m-500m CPU
5. **Replicas**: 2 replicas for high availability
6. **ClusterIP Service**: Internal access only (seller_service integration)

## Troubleshooting

### DynamoDB Access Denied
```bash
# Check IAM role attached to EKS service account
kubectl describe serviceaccount default

# Verify IAM policy allows dynamodb:GetItem, PutItem, UpdateItem, Scan
```

### JWT Validation Fails
```bash
# Verify Cognito User Pool ID
echo $COGNITO_USER_POOL_ID

# Test JWKS endpoint
curl https://cognito-idp.us-east-1.amazonaws.com/us-east-1_eJvqfLh2p/.well-known/jwks.json
```

### EventBridge Not Receiving Events
```bash
# Check EventBridge rule exists
aws events list-rules --region us-east-1

# Verify target is configured
aws events list-targets-by-rule --rule <rule-name> --region us-east-1
```

## Next Steps

- Implement actual EventBridge polling (currently placeholder)
- Add pagination for getAllProducts
- Add product search/filtering capabilities
- Implement review moderation
- Add product image uploads (S3 integration)
- Metrics and observability (CloudWatch, Prometheus)
