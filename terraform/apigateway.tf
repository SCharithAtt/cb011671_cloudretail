# ─────────────────────────────────────────────────────────────────────────────
# API Gateway (HTTP API) → VPC Link → ALB
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_apigatewayv2_api" "main" {
  name          = "${local.name}-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["Content-Type", "Authorization", "X-Amz-Date", "X-Api-Key"]
    max_age       = 3600
  }

  tags = { Name = "${local.name}-api-gateway" }
}

# ── VPC Link for private integration ────────────────────────────────────────

resource "aws_apigatewayv2_vpc_link" "main" {
  name               = "${local.name}-vpc-link"
  security_group_ids = [aws_security_group.alb.id]
  subnet_ids         = aws_subnet.private[*].id

  tags = { Name = "${local.name}-vpc-link" }
}

# ── Default stage with auto deploy ──────────────────────────────────────────

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 100
    throttling_rate_limit  = 50
  }

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId        = "$context.requestId"
      ip               = "$context.identity.sourceIp"
      requestTime      = "$context.requestTime"
      httpMethod       = "$context.httpMethod"
      routeKey         = "$context.routeKey"
      status           = "$context.status"
      protocol         = "$context.protocol"
      responseLength   = "$context.responseLength"
      integrationError = "$context.integrationErrorMessage"
    })
  }
}

resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${local.name}"
  retention_in_days = 14
}

# ── Integration → ALB ──────────────────────────────────────────────────────

resource "aws_apigatewayv2_integration" "alb" {
  api_id             = aws_apigatewayv2_api.main.id
  integration_type   = "HTTP_PROXY"
  integration_uri    = aws_lb_listener.http.arn
  integration_method = "ANY"
  connection_type    = "VPC_LINK"
  connection_id      = aws_apigatewayv2_vpc_link.main.id
}

# ── Catch-all route → ALB ──────────────────────────────────────────────────

resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

# ── Specific routes for each service ────────────────────────────────────────

# User service routes
resource "aws_apigatewayv2_route" "login" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /login"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

resource "aws_apigatewayv2_route" "callback" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /callback"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

# Seller service routes
resource "aws_apigatewayv2_route" "seller_login" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /sellerLogin"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

resource "aws_apigatewayv2_route" "seller_register" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /sellerRegister"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

resource "aws_apigatewayv2_route" "add_product" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /addProduct"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

# Product service (GraphQL)
resource "aws_apigatewayv2_route" "graphql" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /query"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

# Order service routes
resource "aws_apigatewayv2_route" "create_order" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /createOrder"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}

resource "aws_apigatewayv2_route" "get_orders" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /getOrders"
  target    = "integrations/${aws_apigatewayv2_integration.alb.id}"
}
