# ─────────────────────────────────────────────────────────────────────────────
# Outputs
# ─────────────────────────────────────────────────────────────────────────────

output "api_gateway_url" {
  description = "API Gateway invoke URL"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "alb_dns_name" {
  description = "Application Load Balancer DNS"
  value       = aws_lb.main.dns_name
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint"
  value       = aws_db_instance.main.endpoint
}

output "ecr_repositories" {
  description = "ECR repository URLs"
  value = {
    user_service    = aws_ecr_repository.user_service.repository_url
    seller_service  = aws_ecr_repository.seller_service.repository_url
    product_service = aws_ecr_repository.product_service.repository_url
    order_service   = aws_ecr_repository.order_service.repository_url
  }
}

output "dynamodb_tables" {
  description = "DynamoDB table names"
  value = {
    products = aws_dynamodb_table.products.name
    reviews  = aws_dynamodb_table.reviews.name
  }
}

output "eventbridge_bus_arn" {
  description = "EventBridge bus ARN"
  value       = aws_cloudwatch_event_bus.main.arn
}

output "cognito_user_pool_id" {
  description = "Cognito User Pool ID"
  value       = var.cognito_user_pool_id
}

output "amplify_app_url" {
  description = "Amplify application URL"
  value       = var.github_repo_url != "" ? "https://${var.amplify_branch}.${aws_amplify_app.frontend[0].default_domain}" : "Amplify not configured – set github_repo_url"
}

output "frontend_env_vars" {
  description = "Environment variables to set in frontend .env.production"
  value = {
    VITE_API_GATEWAY_URL = aws_apigatewayv2_api.main.api_endpoint
    VITE_GRAPHQL_URL     = "${aws_apigatewayv2_api.main.api_endpoint}/product/graphql"
  }
}
