# ─────────────────────────────────────────────────────────────────────────────
# Lambda – Stock Updater (EventBridge consumer)
# ─────────────────────────────────────────────────────────────────────────────

# The Lambda zip must be built before terraform apply – see deploy.sh
resource "aws_lambda_function" "stock_updater" {
  function_name = "${local.name}-stock-updater"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 30
  memory_size   = 128

  filename         = "${path.module}/../lambda/stock_updater/bootstrap.zip"
  source_code_hash = filebase64sha256("${path.module}/../lambda/stock_updater/bootstrap.zip")

  environment {
    variables = {
      PRODUCTS_TABLE = aws_dynamodb_table.products.name
      AWS_REGION_VAL = var.aws_region
    }
  }

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }

  tags = { Name = "${local.name}-stock-updater" }
}

resource "aws_cloudwatch_log_group" "stock_updater" {
  name              = "/aws/lambda/${aws_lambda_function.stock_updater.function_name}"
  retention_in_days = 14
}
