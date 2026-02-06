# Product Service API Documentation

## Overview

The Product Service provides GraphQL API for product and review management, using DynamoDB for storage and EventBridge for event publishing.

**Base URL:** `http://localhost:8082` (Development)  
**Production URL:** `https://44lkl1on22.execute-api.us-east-1.amazonaws.com`  
**Port:** 8082

## Architecture

- **API Type:** GraphQL
- **Framework:** Go with gqlgen
- **Database:** AWS DynamoDB (Products, Reviews tables)
- **Events:** AWS EventBridge
- **Authorization:** JWT validation via Cognito JWKS (mutations only)

---

## GraphQL Endpoint

**Endpoint:** `POST /graphql`

**Playground:** `GET /` (Development only)

---

## Schema

### Types

#### Product
```graphql
type Product {
  productId: ID!
  name: String!
  price: Float!
  description: String
  stock: Int!
  sellerId: String!
  imageUrl: String
  createdAt: String
  updatedAt: String
  reviews: [Review!]
}
```

#### Review
```graphql
type Review {
  reviewId: ID!
  productId: ID!
  text: String
  rating: Int!
  userId: String!
  createdAt: String
}
```

---

## Queries

### 1. Get All Products

Retrieve all products with optional seller filtering.

```graphql
query GetProducts($filter: ProductFilter) {
  products(filter: $filter) {
    productId
    name
    price
    description
    stock
    sellerId
    imageUrl
    createdAt
  }
}
```

**Variables:**
```json
{
  "filter": {
    "sellerId": "seller-uuid"
  }
}
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { products { productId name price stock } }"}'
```

**Response:**
```json
{
  "data": {
    "products": [
      {
        "productId": "prod-001",
        "name": "Wireless Bluetooth Headphones",
        "price": 17997.0,
        "stock": 150
      }
    ]
  }
}
```

---

### 2. Get Product by ID

Retrieve a single product with its reviews.

```graphql
query GetProduct($id: ID!) {
  product(id: $id) {
    productId
    name
    price
    description
    stock
    sellerId
    imageUrl
    reviews {
      reviewId
      text
      rating
      userId
      createdAt
    }
  }
}
```

**Variables:**
```json
{
  "id": "prod-001"
}
```

**Response:**
```json
{
  "data": {
    "product": {
      "productId": "prod-001",
      "name": "Wireless Bluetooth Headphones",
      "price": 17997.0,
      "description": "Premium wireless headphones...",
      "reviews": [
        {
          "reviewId": "rev-001",
          "text": "Amazing sound quality!",
          "rating": 5,
          "userId": "user-demo-001"
        }
      ]
    }
  }
}
```

---

### 3. Get Reviews for Product

Retrieve all reviews for a specific product.

```graphql
query GetReviews($productId: ID!) {
  reviews(productId: $productId) {
    reviewId
    text
    rating
    userId
    createdAt
  }
}
```

**Variables:**
```json
{
  "productId": "prod-001"
}
```

---

## Mutations

### 1. Add Product

Create a new product (requires authentication).

```graphql
mutation AddProduct($input: AddProductInput!) {
  addProduct(input: $input) {
    productId
    name
    price
  }
}
```

**Input:**
```graphql
input AddProductInput {
  name: String!
  price: Float!
  description: String
  stock: Int!
  sellerId: String!
  imageUrl: String
}
```

**Variables:**
```json
{
  "input": {
    "name": "Wireless Headphones",
    "price": 17997.0,
    "description": "Premium wireless headphones",
    "stock": 100,
    "sellerId": "seller-uuid",
    "imageUrl": "https://via.placeholder.com/400x300"
  }
}
```

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Example:**
```bash
curl -X POST https://44lkl1on22.execute-api.us-east-1.amazonaws.com/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGc..." \
  -d '{"query":"mutation($input:AddProductInput!){addProduct(input:$input){productId}}","variables":{"input":{"name":"Headphones","price":17997.0,"stock":100,"sellerId":"seller-123"}}}'
```

---

### 2. Edit Product

Update an existing product (requires authentication).

```graphql
mutation EditProduct($input: EditProductInput!) {
  editProduct(input: $input) {
    productId
    name
    price
    stock
  }
}
```

**Input:**
```graphql
input EditProductInput {
  productId: ID!
  name: String
  price: Float
  description: String
  stock: Int
  imageUrl: String
}
```

**Variables:**
```json
{
  "input": {
    "productId": "prod-001",
    "price": 14997.0,
    "stock": 150
  }
}
```

---

### 3. Delete Product

Remove a product (requires authentication).

```graphql
mutation DeleteProduct($productId: ID!) {
  deleteProduct(productId: $productId)
}
```

**Variables:**
```json
{
  "productId": "prod-001"
}
```

**Response:**
```json
{
  "data": {
    "deleteProduct": "Product deleted successfully"
  }
}
```

---

### 4. Add Review

Add a review for a product.

```graphql
mutation AddReview($input: AddReviewInput!) {
  addReview(input: $input) {
    reviewId
    text
    rating
  }
}
```

**Input:**
```graphql
input AddReviewInput {
  productId: ID!
  text: String!
  rating: Int!
  userId: String!
}
```

**Variables:**
```json
{
  "input": {
    "productId": "prod-001",
    "text": "Excellent product! Highly recommended.",
    "rating": 5,
    "userId": "user-123"
  }
}
```

---

## REST Endpoints

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "product-service"
}
```

---

## Authentication

### JWT Token Format

**Header:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Note:** Queries are public. Mutations require valid JWT token.

---

## Error Handling

### GraphQL Error Format

```json
{
  "errors": [
    {
      "message": "Product not found",
      "path": ["product"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ],
  "data": null
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHENTICATED` | Missing or invalid JWT token |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `INVALID_INPUT` | Invalid mutation input |
| `INTERNAL_ERROR` | Server error |

---

## Events

### ProductStockUpdated Event

Published to EventBridge when stock changes.

**Event Pattern:**
```json
{
  "source": "product.service",
  "detail-type": "ProductStockUpdated",
  "detail": {
    "productId": "prod-001",
    "oldStock": 100,
    "newStock": 98,
    "timestamp": "2026-02-07T10:30:00Z"
  }
}
```

---

## Environment Variables

```bash
# AWS Configuration
AWS_REGION=us-east-1
DYNAMODB_PRODUCTS_TABLE=Products
DYNAMODB_REVIEWS_TABLE=Reviews
EVENTBRIDGE_BUS_NAME=cloudretail-events

# Cognito Configuration
COGNITO_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_eJvqfLh2p

# Server Configuration
PORT=8082
```

---

## Database Schema

### Products Table (DynamoDB)

**Primary Key:** `productId` (String)

**Attributes:**
- `productId` (S) - UUID
- `name` (S) - Product name
- `price` (N) - Price in LKR
- `description` (S) - Product description
- `stock` (N) - Available quantity
- `sellerId` (S) - Seller UUID
- `imageUrl` (S) - Product image URL
- `createdAt` (S) - ISO 8601 timestamp
- `updatedAt` (S) - ISO 8601 timestamp

**GSI:** `sellerId-index` for querying products by seller

---

### Reviews Table (DynamoDB)

**Primary Key:** `reviewId` (String)

**Attributes:**
- `reviewId` (S) - UUID
- `productId` (S) - Product UUID
- `text` (S) - Review text
- `rating` (N) - Rating (1-5)
- `userId` (S) - User UUID
- `createdAt` (S) - ISO 8601 timestamp

**GSI:** `productId-index` for querying reviews by product

---

## Testing

```bash
# Start service
cd backend/services/product_service
go run main.go

# Test health endpoint
curl http://localhost:8082/health

# Test GraphQL query
curl -X POST http://localhost:8082/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ products { productId name price } }"}'

# Run unit tests
go test -v

# Open GraphQL playground (development)
# Open http://localhost:8082 in browser
```

---

## GraphQL Playground

Access the interactive playground at `http://localhost:8082/` in development mode.

**Example Query:**
```graphql
query {
  products {
    productId
    name
    price
    reviews {
      rating
      text
    }
  }
}
```

---

## Currency Note

All prices are stored and returned in **LKR (Sri Lankan Rupees)**.

**Conversion:** USD × 300 = LKR

Example:
- $59.99 USD = 17,997 LKR
- $29.99 USD = 8,997 LKR
