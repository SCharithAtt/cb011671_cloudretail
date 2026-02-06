# ─────────────────────────────────────────────────────────────────────────────
# ECR Repositories
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_ecr_repository" "user_service" {
  name                 = "${local.name}/user-service"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "seller_service" {
  name                 = "${local.name}/seller-service"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "product_service" {
  name                 = "${local.name}/product-service"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "order_service" {
  name                 = "${local.name}/order-service"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

# ── Lifecycle policy – keep last 10 images ──────────────────────────────────

resource "aws_ecr_lifecycle_policy" "cleanup" {
  for_each   = toset(["user-service", "seller-service", "product-service", "order-service"])
  repository = "${local.name}/${each.key}"

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 10 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = { type = "expire" }
    }]
  })

  depends_on = [
    aws_ecr_repository.user_service,
    aws_ecr_repository.seller_service,
    aws_ecr_repository.product_service,
    aws_ecr_repository.order_service,
  ]
}
