# ─────────────────────────────────────────────────────────────────────────────
# ECS EC2 – Cluster · CloudMap · Task Definitions · Services
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_ecs_cluster" "main" {
  name = "${local.name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Name = "${local.name}-cluster" }
}

# ── Capacity Provider ───────────────────────────────────────────────────────

resource "aws_ecs_capacity_provider" "ec2" {
  name = "${local.name}-ec2-capacity-provider"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.ecs.arn
    managed_termination_protection = "ENABLED"

    managed_scaling {
      status                    = "ENABLED"
      target_capacity           = 80
      minimum_scaling_step_size = 1
      maximum_scaling_step_size = 100
    }
  }

  tags = { Name = "${local.name}-capacity-provider" }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = [aws_ecs_capacity_provider.ec2.name]

  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ec2.name
    weight            = 100
    base              = 1
  }
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
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name         = "user-service"
    image        = "${aws_ecr_repository.user_service.repository_url}:latest"
    portMappings = [{ containerPort = 8080, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8080" },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "COGNITO_CLIENT_ID", value = var.cognito_client_id },
      { name = "COGNITO_CLIENT_SECRET", value = var.cognito_client_secret },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "COGNITO_REDIRECT_URL", value = "${aws_apigatewayv2_api.main.api_endpoint}/callback" },
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
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name         = "seller-service"
    image        = "${aws_ecr_repository.seller_service.repository_url}:latest"
    portMappings = [{ containerPort = 8081, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8081" },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "COGNITO_CLIENT_ID", value = var.cognito_client_id },
      { name = "COGNITO_CLIENT_SECRET", value = var.cognito_client_secret },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "PRODUCT_GRAPHQL_URL", value = "http://product-service.${local.name}.local:8082/graphql" },
      { name = "ORDER_REST_URL", value = "http://order-service.${local.name}.local:8083" },
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
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name         = "product-service"
    image        = "${aws_ecr_repository.product_service.repository_url}:latest"
    portMappings = [{ containerPort = 8082, protocol = "tcp" }]
    environment = [
      { name = "PORT", value = "8082" },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "COGNITO_REGION", value = var.aws_region },
      { name = "COGNITO_USER_POOL_ID", value = var.cognito_user_pool_id },
      { name = "DYNAMODB_ENDPOINT", value = "" },
      { name = "PRODUCTS_TABLE", value = aws_dynamodb_table.products.name },
      { name = "REVIEWS_TABLE", value = aws_dynamodb_table.reviews.name },
      { name = "EVENT_BUS_NAME", value = aws_cloudwatch_event_bus.main.name },
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
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name         = "order-service"
    image        = "${aws_ecr_repository.order_service.repository_url}:latest"
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

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ec2.name
    weight            = 100
    base              = 1
  }

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

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ec2.name
    weight            = 100
    base              = 1
  }

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

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ec2.name
    weight            = 100
    base              = 1
  }

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

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ec2.name
    weight            = 100
    base              = 1
  }

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

# ─────────────────────────────────────────────────────────────────────────────
# Service Auto Scaling (scale task count based on load)
# ─────────────────────────────────────────────────────────────────────────────

# ── User Service Auto Scaling ───────────────────────────────────────────────

resource "aws_appautoscaling_target" "user" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.user.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "user_cpu" {
  name               = "${local.name}-user-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.user.resource_id
  scalable_dimension = aws_appautoscaling_target.user.scalable_dimension
  service_namespace  = aws_appautoscaling_target.user.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "user_memory" {
  name               = "${local.name}-user-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.user.resource_id
  scalable_dimension = aws_appautoscaling_target.user.scalable_dimension
  service_namespace  = aws_appautoscaling_target.user.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# ── Seller Service Auto Scaling ─────────────────────────────────────────────

resource "aws_appautoscaling_target" "seller" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.seller.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "seller_cpu" {
  name               = "${local.name}-seller-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.seller.resource_id
  scalable_dimension = aws_appautoscaling_target.seller.scalable_dimension
  service_namespace  = aws_appautoscaling_target.seller.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "seller_memory" {
  name               = "${local.name}-seller-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.seller.resource_id
  scalable_dimension = aws_appautoscaling_target.seller.scalable_dimension
  service_namespace  = aws_appautoscaling_target.seller.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# ── Product Service Auto Scaling ────────────────────────────────────────────

resource "aws_appautoscaling_target" "product" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.product.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "product_cpu" {
  name               = "${local.name}-product-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.product.resource_id
  scalable_dimension = aws_appautoscaling_target.product.scalable_dimension
  service_namespace  = aws_appautoscaling_target.product.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "product_memory" {
  name               = "${local.name}-product-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.product.resource_id
  scalable_dimension = aws_appautoscaling_target.product.scalable_dimension
  service_namespace  = aws_appautoscaling_target.product.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# ── Order Service Auto Scaling ──────────────────────────────────────────────

resource "aws_appautoscaling_target" "order" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.order.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "order_cpu" {
  name               = "${local.name}-order-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.order.resource_id
  scalable_dimension = aws_appautoscaling_target.order.scalable_dimension
  service_namespace  = aws_appautoscaling_target.order.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "order_memory" {
  name               = "${local.name}-order-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.order.resource_id
  scalable_dimension = aws_appautoscaling_target.order.scalable_dimension
  service_namespace  = aws_appautoscaling_target.order.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
