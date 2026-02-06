#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# CloudRetail – Full Deployment Script
# Deploys infrastructure via Terraform, builds & pushes Docker images to ECR,
# and forces ECS service updates.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
PROJECT="cloudretail"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "═══════════════════════════════════════════════════════════════"
echo "  CloudRetail Deployment"
echo "  Region:  $REGION"
echo "═══════════════════════════════════════════════════════════════"

# ── Step 1: Build Lambda ────────────────────────────────────────────────────
echo ""
echo "▶ [1/5] Building Lambda stock-updater..."
cd "$SCRIPT_DIR/lambda/stock_updater"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip -j bootstrap.zip bootstrap
echo "  ✓ Lambda built: bootstrap.zip"

# ── Step 2: Terraform ───────────────────────────────────────────────────────
echo ""
echo "▶ [2/5] Running Terraform..."
cd "$SCRIPT_DIR/terraform"
terraform init
terraform plan -out=tfplan
echo ""
read -p "  Apply Terraform plan? (y/n) " confirm
if [[ "$confirm" != "y" ]]; then
  echo "  ✗ Aborted."
  exit 1
fi
terraform apply tfplan

# Extract outputs
API_URL=$(terraform output -raw api_gateway_url 2>/dev/null || echo "")
ALB_DNS=$(terraform output -raw alb_dns_name 2>/dev/null || echo "")
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "  ✓ Terraform applied"
echo "    API Gateway:  $API_URL"
echo "    ALB DNS:      $ALB_DNS"

# ── Step 3: Login to ECR ───────────────────────────────────────────────────
echo ""
echo "▶ [3/5] Logging in to ECR..."
aws ecr get-login-password --region "$REGION" | \
  docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"
echo "  ✓ ECR login successful"

# ── Step 4: Build & Push Docker Images ──────────────────────────────────────
echo ""
echo "▶ [4/5] Building and pushing Docker images..."
cd "$SCRIPT_DIR"

SERVICES=("user_service" "seller_service" "product_service" "order_service")
ECR_NAMES=("user-service" "seller-service" "product-service" "order-service")

for i in "${!SERVICES[@]}"; do
  svc="${SERVICES[$i]}"
  ecr_name="${ECR_NAMES[$i]}"
  repo_url="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$PROJECT/$ecr_name"

  echo "  Building $svc..."
  docker build -t "$PROJECT/$ecr_name:latest" "./backend/services/$svc"
  docker tag "$PROJECT/$ecr_name:latest" "$repo_url:latest"
  docker push "$repo_url:latest"
  echo "  ✓ $ecr_name pushed to ECR"
done

# ── Step 5: Force ECS Service Update ───────────────────────────────────────
echo ""
echo "▶ [5/5] Forcing ECS service updates..."
CLUSTER="${PROJECT}-cluster"

for ecr_name in "${ECR_NAMES[@]}"; do
  aws ecs update-service \
    --cluster "$CLUSTER" \
    --service "${ecr_name}" \
    --force-new-deployment \
    --region "$REGION" \
    --no-cli-pager > /dev/null
  echo "  ✓ ${ecr_name} deployment triggered"
done

# ── Done ────────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  ✓ Deployment Complete!"
echo ""
echo "  API Gateway URL:     $API_URL"
echo "  ALB DNS:             $ALB_DNS"
echo ""
echo "  Update your frontend .env.production with:"
echo "    VITE_API_GATEWAY_URL=$API_URL"
echo ""
echo "  If using Amplify, push to your repo's main branch to"
echo "  trigger a frontend deployment."
echo "═══════════════════════════════════════════════════════════════"
