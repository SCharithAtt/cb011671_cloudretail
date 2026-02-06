# CloudRetail – Cloud-Native E-Commerce Platform

A full-stack e-commerce platform built with **Go microservices**, **Vue 3** frontend, and deployed on **AWS** using **Terraform**.

## Architecture

```
┌─────────────┐     ┌───────────────────┐     ┌──────────────────────────────────────┐
│   Amplify    │     │   API Gateway     │     │          ECS Fargate Cluster          │
│  (Frontend)  │────▶│   (HTTP API)      │────▶│  ┌────────┐ ┌────────┐ ┌──────────┐ │
│  Vue 3 + TW  │     │   + VPC Link      │     │  │ User   │ │Seller  │ │ Product  │ │
└─────────────┘     └───────────────────┘     │  │ :8080  │ │ :8081  │ │  :8082   │ │
                                               │  └────────┘ └────────┘ └──────────┘ │
                                               │  ┌────────┐                          │
                                               │  │ Order  │                          │
                                               │  │ :8083  │                          │
                                               │  └───┬────┘                          │
                                               └──────┼──────────────────────────────┘
                                                      │
                          ┌───────────────────────────┼───────────────────┐
                          │                           │                   │
                   ┌──────▼──────┐            ┌───────▼─────┐    ┌───────▼──────┐
                   │    RDS       │            │ EventBridge │    │  DynamoDB    │
                   │ PostgreSQL   │            │  Event Bus  │    │  Products    │
                   │ (Orders DB)  │            └──────┬──────┘    │  Reviews     │
                   └─────────────┘                    │           └──────────────┘
                                               ┌──────▼──────┐          │
                                               │   Lambda     │──────────┘
                                               │ Stock Update │
                                               └─────────────┘
```

## Microservices

| Service | Port | Framework | Database | Auth |
|---------|------|-----------|----------|------|
| **user_service** | 8080 | net/http | - | Cognito OIDC/OAuth2 |
| **seller_service** | 8081 | Gin | - | Cognito SDK (Admin) |
| **product_service** | 8082 | Gin + gqlgen | DynamoDB | JWT |
| **order_service** | 8083 | Gin + GORM | RDS PostgreSQL | JWT |

## AWS Services Used

- **ECS Fargate** – Container orchestration for all 4 microservices
- **ECR** – Docker image registry
- **RDS PostgreSQL (db.t3.micro)** – Order database
- **DynamoDB** – Products & Reviews tables
- **API Gateway (HTTP API)** – Unified API endpoint with CORS
- **Application Load Balancer** – Path-based routing to services
- **EventBridge** – Async event bus for order-placed events
- **Lambda** – Stock updater (Go) triggered by EventBridge
- **Amplify** – Frontend hosting with CI/CD
- **Cognito** – User authentication (existing pool)
- **CloudWatch** – Logging and monitoring
- **VPC** – Networking with public/private subnets, NAT Gateway

## Project Structure

```
├── backend/services/
│   ├── user_service/       # OAuth2/OIDC user authentication
│   ├── seller_service/     # Seller auth + product/order proxy
│   ├── product_service/    # GraphQL API (DynamoDB)
│   └── order_service/      # REST API (PostgreSQL)
├── frontend/               # Vue 3 + Tailwind CSS (yellow theme)
├── lambda/stock_updater/   # Go Lambda for EventBridge events
├── terraform/              # Full AWS infrastructure (16 files)
├── docker-compose.yml      # Local development environment
├── deploy.sh               # Automated deployment script
└── amplify.yml             # Amplify build specification
```

## Quick Start – Local Development

```bash
# Start all services locally
docker-compose up --build

# Frontend: http://localhost:3000
# User Service: http://localhost:8080
# Seller Service: http://localhost:8081
# Product Service: http://localhost:8082/query (GraphQL)
# Order Service: http://localhost:8083
```

## Deployment to AWS

### Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform >= 1.5.0
- Docker
- Go >= 1.21

### Steps

1. **Configure variables:**
   ```bash
   cd terraform
   cp terraform.tfvars terraform.tfvars  # Edit with your values
   ```

2. **Run the deploy script:**
   ```bash
   chmod +x deploy.sh
   ./deploy.sh
   ```

   The script will:
   - Build the Lambda function
   - Run `terraform init` + `terraform apply`
   - Build all Docker images and push to ECR
   - Force ECS service re-deployments

3. **Frontend deployment (Amplify):**
   - Set your GitHub repo URL in `terraform.tfvars`
   - Or push to your repo's `main` branch after Terraform creates the Amplify app

### Manual Terraform Commands

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

## Environment Variables

### Backend Services (set in ECS task definitions)

| Variable | Service | Description |
|----------|---------|-------------|
| `COGNITO_USER_POOL_ID` | user, seller | Cognito User Pool ID |
| `COGNITO_CLIENT_ID` | user, seller | Cognito App Client ID |
| `COGNITO_CLIENT_SECRET` | user, seller | Cognito Client Secret |
| `DATABASE_URL` | order | PostgreSQL connection string |
| `PRODUCT_SERVICE_URL` | seller, order | Product service URL |
| `ORDER_SERVICE_URL` | seller | Order service URL |
| `EVENTBRIDGE_BUS_NAME` | order | EventBridge bus name |
| `DYNAMODB_ENDPOINT` | product | DynamoDB endpoint (empty for prod) |

### Frontend (set in Amplify / .env.production)

| Variable | Description |
|----------|-------------|
| `VITE_API_GATEWAY_URL` | API Gateway endpoint URL |
| `VITE_GRAPHQL_URL` | GraphQL endpoint URL |
| `VITE_COGNITO_DOMAIN` | Cognito hosted UI domain |
| `VITE_COGNITO_CLIENT_ID` | Cognito client ID |
| `VITE_REDIRECT_URI` | OAuth2 callback URL |

## Testing

```bash
# Run tests for each service
cd backend/services/user_service && go test -v ./...
cd backend/services/seller_service && go test -v ./...
cd backend/services/product_service && go test -v ./...
cd backend/services/order_service && go test -v ./...
```
