# CloudRetail Frontend

Vue 3 + TypeScript + Vite frontend application for the CloudRetail e-commerce platform.

## Features

- **Authentication**: User login via AWS Cognito OAuth2, Seller login via JWT
- **State Management**: Pinia stores for auth and shopping cart
- **GraphQL Integration**: Apollo Client for ProductService queries/mutations
- **REST API Integration**: Axios for OrderService, SellerService, UserService
- **Role-Based Access**: Navigation guards for buyer/seller routes
- **Shopping Cart**: Persistent cart with localStorage
- **Payment Simulation**: Checkbox-based payment flow
- **Real-time Reviews**: GraphQL mutations for product reviews

## Architecture

```
frontend/
├── src/
│   ├── stores/           # Pinia stores
│   │   ├── auth.ts       # Authentication state & JWT parsing
│   │   └── cart.ts       # Shopping cart state
│   ├── services/         # API clients
│   │   ├── api.ts        # Axios REST client with interceptors
│   │   └── graphql.ts    # Apollo GraphQL client
│   ├── router/           # Vue Router configuration
│   │   └── index.ts      # Routes & navigation guards
│   ├── views/            # Page components
│   │   ├── Home.vue
│   │   ├── Login.vue
│   │   ├── Register.vue
│   │   ├── OAuthCallback.vue
│   │   ├── ProductList.vue
│   │   ├── ProductDetails.vue
│   │   ├── Cart.vue
│   │   ├── Checkout.vue
│   │   ├── Payment.vue
│   │   ├── OrderConfirmed.vue
│   │   ├── Orders.vue
│   │   ├── SellerDashboard.vue
│   │   ├── SellerProducts.vue
│   │   ├── SellerOrders.vue
│   │   ├── AddProduct.vue
│   │   ├── EditProduct.vue
│   │   └── NotFound.vue
│   ├── components/       # Reusable components
│   │   ├── Header.vue
│   │   └── Footer.vue
│   ├── App.vue           # Root component
│   └── main.ts           # App entry point
├── package.json
├── vite.config.ts
├── tsconfig.json
└── .env                  # Development environment variables
```

## Tech Stack

- **Vue 3.4.0**: Progressive JavaScript framework
- **TypeScript 5.3.0**: Type safety
- **Vite 5.0.0**: Fast build tool
- **Pinia 2.1.7**: State management
- **Vue Router 4.2.0**: Routing with guards
- **Axios 1.6.0**: REST API client
- **Apollo Client 3.9.0**: GraphQL client
- **jwt-decode 4.0.0**: JWT parsing

## Setup

### Prerequisites

- Node.js 18+ and npm
- Backend services running (user, seller, product, order services)

### Installation

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Copy environment file
cp .env .env.local

# Update .env.local with your backend URLs (if different from defaults)
```

### Environment Variables

Create a `.env` file (already provided):

```bash
# Development (localhost)
VITE_API_GATEWAY_URL=http://localhost:8080
VITE_USER_SERVICE_URL=http://localhost:8080
VITE_SELLER_SERVICE_URL=http://localhost:8081
VITE_PRODUCT_SERVICE_URL=http://localhost:8082
VITE_ORDER_SERVICE_URL=http://localhost:8083
VITE_GRAPHQL_URL=http://localhost:8082/graphql
```

For production, create `.env.production`:

```bash
# Production (API Gateway)
VITE_API_GATEWAY_URL=https://api.cloudretail.example.com
VITE_GRAPHQL_URL=https://api.cloudretail.example.com/product/graphql
```

## Development

```bash
# Start development server (port 3000)
npm run dev

# Type checking
npm run typecheck

# Build for production
npm run build

# Preview production build
npm run preview
```

## API Integration

### Authentication Flow

**User Login (Cognito OAuth2)**:
1. User clicks "Login with Cognito"
2. Frontend redirects to `/login` endpoint in UserService
3. UserService redirects to Cognito hosted UI
4. After authentication, Cognito redirects to `/callback`
5. OAuthCallback.vue extracts tokens from URL query params
6. Tokens stored in auth store and localStorage

**Seller Login (JWT)**:
1. Seller enters email/password in Login.vue
2. POST to `/sellerLogin` endpoint in SellerService
3. Backend returns `id_token`, `access_token`, `refresh_token`
4. JWT parsed to extract `sub` (userId) and `custom:role` (seller)
5. Tokens stored in auth store and localStorage

### REST API Endpoints

**UserService (8080)**:
- `GET /login` - Redirect to Cognito
- `GET /callback` - OAuth2 callback
- `GET /logout` - Logout

**SellerService (8081)**:
- `POST /sellerLogin` - Seller authentication
- `POST /sellerRegister` - Seller registration
- `POST /addProduct` - Add product (calls ProductService GraphQL)
- `PUT /editProduct/:productId` - Edit product
- `GET /orders` - Get seller orders
- `PUT /updateOrderStatus/:orderId` - Update order status

**OrderService (8083)**:
- `POST /createOrder` - Create order (fires EventBridge event)
- `GET /simulatePayment/:orderId` - Get payment page
- `POST /markPaymentDone/:orderId` - Complete payment
- `GET /orderConfirmed/:orderId` - Check order status
- `GET /getOrders` - Get buyer orders
- `PUT /updateStatus/:orderId` - Update order status

### GraphQL API (ProductService 8082)

**Queries**:
```graphql
query GetAllProducts {
  getAllProducts {
    product_id
    name
    description
    price
    stock
    seller_id
    reviews {
      rating
      comment
      timestamp
    }
  }
}

query GetProductById($productId: ID!) {
  getProductById(productId: $productId) {
    product_id
    name
    description
    price
    stock
    seller_id
    reviews {
      rating
      comment
      timestamp
    }
  }
}
```

**Mutations**:
```graphql
mutation AddReview($productId: ID!, $rating: Int!, $comment: String!) {
  addReview(productId: $productId, rating: $rating, comment: $comment) {
    product_id
    reviews {
      rating
      comment
      timestamp
    }
  }
}
```

## Features Guide

### Shopping Cart

- **Add to Cart**: ProductList.vue, ProductDetails.vue
- **Persistent**: Stored in localStorage via cart store
- **Update Quantity**: Cart.vue allows increment/decrement
- **Remove Items**: Cart.vue
- **Clear on Checkout**: Cleared after order creation

### Order Flow

1. **Browse Products**: ProductList.vue (GraphQL getAllProducts)
2. **Add to Cart**: ProductDetails.vue
3. **Checkout**: Checkout.vue → POST /createOrder
4. **Payment Simulation**: Payment.vue → GET /simulatePayment/:orderId → POST /markPaymentDone/:orderId
5. **Confirmation**: OrderConfirmed.vue → GET /orderConfirmed/:orderId
6. **View Orders**: Orders.vue → GET /getOrders

### Seller Dashboard

- **Dashboard**: SellerDashboard.vue - Stats overview
- **Products**: SellerProducts.vue - List seller products
- **Add Product**: AddProduct.vue → POST /addProduct (SellerService)
- **Edit Product**: EditProduct.vue → PUT /editProduct/:productId
- **Orders**: SellerOrders.vue - Manage order status

## Routing & Guards

Navigation guards in [router/index.ts](router/index.ts):

- **requiresAuth**: Requires logged-in user
- **requiresRole**: Requires specific role (seller/buyer)
- **guest**: Only accessible when not logged in

Example routes:
- `/` - Home (public)
- `/products` - Product list (public)
- `/cart` - Shopping cart (requiresAuth)
- `/seller` - Seller dashboard (requiresAuth + requiresRole: seller)

## State Management

### Auth Store ([stores/auth.ts](stores/auth.ts))

```typescript
// State
token: string
refreshToken: string
role: string  // 'seller' or empty
userId: string
userEmail: string

// Actions
login(credentials: LoginCredentials)
registerSeller(credentials)
logout()
parseJWT(token: string)  // Extract sub and custom:role
handleOAuthCallback(idToken, accessToken, refreshToken)
refreshAccessToken()
```

### Cart Store ([stores/cart.ts](stores/cart.ts))

```typescript
// State
items: CartItem[]

// Computed
itemCount: number
total: number

// Actions
addItem(product: CartItem)
updateQuantity(productId: string, quantity: number)
removeItem(productId: string)
clear()
```

## Deployment

### Build for Production

```bash
npm run build
# Output: dist/
```

### Deployment Options

#### Option 1: AWS S3 + CloudFront

1. **Build**:
   ```bash
   npm run build
   ```

2. **Upload to S3**:
   ```bash
   aws s3 sync dist/ s3://your-bucket-name --delete
   ```

3. **Create CloudFront Distribution**:
   - Origin: S3 bucket
   - Behavior: Redirect all to index.html for SPA routing
   - Custom error pages: 404 → /index.html (200)

4. **Update Environment Variables**:
   - Set `VITE_API_GATEWAY_URL` in `.env.production` before build
   - Rebuild for each environment

#### Option 2: Docker + Nginx

```dockerfile
# Build stage
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**nginx.conf**:
```nginx
server {
  listen 80;
  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }
}
```

**Build & Run**:
```bash
docker build -t cloudretail-frontend .
docker run -p 80:80 cloudretail-frontend
```

## Troubleshooting

### CORS Issues

If you encounter CORS errors:
- Backend services must include CORS headers
- Check `Access-Control-Allow-Origin` in backend responses
- Vite proxy only works in development

### JWT Expiration

- Token refresh logic in [services/api.ts](services/api.ts) handles 401 errors
- User redirected to login if refresh fails
- Check token expiration in auth store

### GraphQL Errors

- Check Apollo Client devtools (browser extension)
- Verify GraphQL endpoint is reachable
- Check network tab for GraphQL responses

### Environment Variables

- Ensure `.env` file exists
- Vite requires `VITE_` prefix for environment variables
- Restart dev server after changing `.env`

## License

MIT
