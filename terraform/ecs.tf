# ─────────────────────────────────────────────────────────────────────────────
# ECS Fargate – Cluster · CloudMap · Task Definitions · Services
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_ecs_cluster" "main" {
  name = "${local.name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Name = "${local.name}-cluster" }
}

# ── Service Discovery (Cloud Map) ──────────────────────────────────────────

resource "aws_service_discovery_private_dns_namespace" "main" {
  name = "${local.name}.local"
  vpc  = aws_vpc.main.id
}

resource "aws_service_discovery_service" "user" {
  name = "user-service"
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id
    dns_records {
      type = "A"
      ttl  = 10
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_service" "seller" {
  name = "seller-service"
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id
    dns_records {
      type = "A"
      ttl  = 10
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_service" "product" {
  name = "product-service"
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id
    dns_records {
      type = "A"
      ttl  = 10
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_service" "order" {
  name = "order-service"
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id
    dns_records {
      type = "A"
      ttl  = 10
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = 1
  }
}

# ── CloudWatch Log Groups ──────────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "user" {
  name              = "/ecs/${local.name}/user-service"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "seller" {
  name              = "/ecs/${local.name}/seller-service"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "product" {
  name              = "/ecs/${local.name}/product-service"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "order" {
  name              = "/ecs/${local.name}/order-service"
  retention_in_days = 14
}

# ─────────────────────────────────────────────────────────────────────────────
# Task Definitions
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_ecs_task_definition" "user" {
  family                   = "${local.name}-user"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name  = "user-service"
    image = "${aws_ecr_repository.user_service.repository_url}:latest"
    portMappings = [{ containerPort = 8080, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8080" },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "COGNITO_CLIENT_ID", value = var.cognito_client_id },
      { name = "COGNITO_CLIENT_SECRET", value = var.cognito_client_secret },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "COGNITO_REDIRECT_URL", value = "https://${aws_apigatewayv2_api.main.api_endpoint}/callback" },
      { name = "COGNITO_LOGOUT_URL", value = "https://main.${aws_amplify_app.frontend[0].default_domain}" },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.user.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "seller" {
  family                   = "${local.name}-seller"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name  = "seller-service"
    image = "${aws_ecr_repository.seller_service.repository_url}:latest"
    portMappings = [{ containerPort = 8081, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8081" },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "COGNITO_CLIENT_ID", value = var.cognito_client_id },
      { name = "COGNITO_CLIENT_SECRET", value = var.cognito_client_secret },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "PRODUCT_SERVICE_URL", value = "http://product-service.${local.name}.local:8082" },
      { name = "ORDER_SERVICE_URL", value = "http://order-service.${local.name}.local:8083" },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.seller.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "product" {
  family                   = "${local.name}-product"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name  = "product-service"
    image = "${aws_ecr_repository.product_service.repository_url}:latest"
    portMappings = [{ containerPort = 8082, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8082" },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "DYNAMODB_ENDPOINT", value = "" },
      { name = "PRODUCTS_TABLE", value = aws_dynamodb_table.products.name },
      { name = "REVIEWS_TABLE", value = aws_dynamodb_table.reviews.name },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.product.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "order" {
  family                   = "${local.name}-order"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name  = "order-service"
    image = "${aws_ecr_repository.order_service.repository_url}:latest"
    portMappings = [{ containerPort = 8083, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8083" },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "RDS_DSN", value = "postgres://${var.db_master_username}:${var.db_master_password}@${aws_db_instance.main.endpoint}/${var.db_name}?sslmode=require" },
      { name = "PRODUCT_GRAPHQL_URL", value = "http://product-service.${local.name}.local:8082/graphql" },
      { name = "EVENTBRIDGE_BUS_ARN", value = aws_cloudwatch_event_bus.main.arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.order.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

# ─────────────────────────────────────────────────────────────────────────────
# ECS Services
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_ecs_service" "user" {
  name            = "user-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.user.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.user.arn
    container_name   = "user-service"
    container_port   = 8080
  }

  service_registries {
    registry_arn = aws_service_discovery_service.user.arn
  }

  depends_on = [aws_lb_listener.http]
}

resource "aws_ecs_service" "seller" {
  name            = "seller-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.seller.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.seller.arn
    container_name   = "seller-service"
    container_port   = 8081
  }

  service_registries {
    registry_arn = aws_service_discovery_service.seller.arn
  }

  depends_on = [aws_lb_listener.http]
}

resource "aws_ecs_service" "product" {
  name            = "product-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.product.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.product.arn
    container_name   = "product-service"
    container_port   = 8082
  }

  service_registries {
    registry_arn = aws_service_discovery_service.product.arn
  }

  depends_on = [aws_lb_listener.http]
}

resource "aws_ecs_service" "order" {
  name            = "order-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.order.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.order.arn
    container_name   = "order-service"
    container_port   = 8083
  }

  service_registries {
    registry_arn = aws_service_discovery_service.order.arn
  }

  depends_on = [aws_lb_listener.http]
}
