# ─────────────────────────────────────────────────────────────────────────────
# Application Load Balancer · Target Groups · Listener Rules
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_lb" "main" {
  name               = "${local.name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  tags = { Name = "${local.name}-alb" }
}

# ── Target Groups (one per microservice) ────────────────────────────────────

resource "aws_lb_target_group" "user" {
  name        = "${local.name}-user-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

resource "aws_lb_target_group" "seller" {
  name        = "${local.name}-seller-tg"
  port        = 8081
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

resource "aws_lb_target_group" "product" {
  name        = "${local.name}-product-tg"
  port        = 8082
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

resource "aws_lb_target_group" "order" {
  name        = "${local.name}-order-tg"
  port        = 8083
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

# ── Listener + Path-Based Routing ───────────────────────────────────────────

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "application/json"
      message_body = "{\"error\":\"not found\"}"
      status_code  = "404"
    }
  }
}

# /user/* → user_service
resource "aws_lb_listener_rule" "user" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 100

  condition {
    path_pattern { values = ["/login*", "/callback*", "/logout*", "/health"] }
  }
  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.user.arn
  }
}

# /seller* → seller_service
resource "aws_lb_listener_rule" "seller" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 200

  condition {
    path_pattern { values = ["/seller*", "/addProduct*", "/editProduct*", "/orders*", "/updateOrderStatus*"] }
  }
  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.seller.arn
  }
}

# /graphql → product_service (GraphQL)
resource "aws_lb_listener_rule" "product" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 300

  condition {
    path_pattern { values = ["/graphql*", "/query*"] }
  }
  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.product.arn
  }
}

# /order* → order_service
resource "aws_lb_listener_rule" "order" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 400

  condition {
    path_pattern { values = ["/createOrder*", "/simulatePayment*", "/markPaymentDone*", "/orderConfirmed*", "/getOrders*", "/updateStatus*"] }
  }
  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.order.arn
  }
}
