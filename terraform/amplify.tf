# ─────────────────────────────────────────────────────────────────────────────
# AWS Amplify – Frontend Deployment
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_amplify_app" "frontend" {
  count = var.github_repo_url != "" ? 1 : 0

  name       = "${local.name}-frontend"
  repository = var.github_repo_url

  access_token = var.github_token

  build_spec = <<-YAML
    version: 1
    frontend:
      phases:
        preBuild:
          commands:
            - cd frontend
            - npm ci
        build:
          commands:
            - npm run build
      artifacts:
        baseDirectory: frontend/dist
        files:
          - '**/*'
      cache:
        paths:
          - frontend/node_modules/**/*
  YAML

  custom_rule {
    source = "/<*>"
    status = "404-200"
    target = "/index.html"
  }

  environment_variables = {
    VITE_API_BASE_URL    = aws_apigatewayv2_api.main.api_endpoint
    VITE_COGNITO_DOMAIN  = "https://${local.name}.auth.${var.aws_region}.amazoncognito.com"
    VITE_COGNITO_CLIENT_ID = var.cognito_client_id
    VITE_REDIRECT_URI    = "https://main.${aws_amplify_app.frontend[0].default_domain}/callback"
  }

  tags = { Name = "${local.name}-frontend" }
}

resource "aws_amplify_branch" "main" {
  count = var.github_repo_url != "" ? 1 : 0

  app_id      = aws_amplify_app.frontend[0].id
  branch_name = var.amplify_branch

  stage = "PRODUCTION"

  environment_variables = {
    VITE_API_BASE_URL = aws_apigatewayv2_api.main.api_endpoint
  }
}
