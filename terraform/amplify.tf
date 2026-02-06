# ─────────────────────────────────────────────────────────────────────────────
# AWS Amplify – Frontend Deployment
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_amplify_app" "frontend" {
  count = var.github_repo_url != "" ? 1 : 0

  name       = "${local.name}-frontend"
  repository = var.github_repo_url

  oauth_token = var.github_token

  build_spec = <<-YAML
    version: 1
    applications:
      - appRoot: frontend
        frontend:
          phases:
            preBuild:
              commands:
                - npm ci
            build:
              commands:
                - npm run build
          artifacts:
            baseDirectory: dist
            files:
              - '**/*'
          cache:
            paths:
              - node_modules/**/*
  YAML

  custom_rule {
    source = "/<*>"
    status = "404-200"
    target = "/index.html"
  }

  environment_variables = {
    AMPLIFY_MONOREPO_APP_ROOT = "frontend"
    AMPLIFY_DIFF_DEPLOY       = "false"
    VITE_API_GATEWAY_URL      = aws_apigatewayv2_api.main.api_endpoint
    VITE_GRAPHQL_URL          = "${aws_apigatewayv2_api.main.api_endpoint}/graphql"
    VITE_COGNITO_DOMAIN       = "https://us-east-1ejvqflh2p.auth.${var.aws_region}.amazoncognito.com"
    VITE_COGNITO_CLIENT_ID    = var.cognito_client_id
    VITE_REDIRECT_URI         = "https://main.d1zj0qo7tbzc0o.amplifyapp.com/callback"
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
    VITE_REDIRECT_URI = "https://main.${aws_amplify_app.frontend[0].default_domain}/callback"
  }
}
