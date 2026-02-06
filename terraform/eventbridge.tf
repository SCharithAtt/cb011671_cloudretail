# ─────────────────────────────────────────────────────────────────────────────
# EventBridge · Custom Event Bus · Rules
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_cloudwatch_event_bus" "main" {
  name = "${local.name}-events"
  tags = { Name = "${local.name}-event-bus" }
}

# Rule: capture order-placed events → trigger Lambda stock updater
resource "aws_cloudwatch_event_rule" "order_placed" {
  name           = "${local.name}-order-placed"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  description    = "Captures order-placed events to update product stock"

  event_pattern = jsonencode({
    source      = ["cloudretail.order-service"]
    detail-type = ["order-placed"]
  })

  tags = { Name = "${local.name}-order-placed-rule" }
}

resource "aws_cloudwatch_event_target" "stock_updater" {
  rule           = aws_cloudwatch_event_rule.order_placed.name
  event_bus_name = aws_cloudwatch_event_bus.main.name
  target_id      = "stock-updater-lambda"
  arn            = aws_lambda_function.stock_updater.arn
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.stock_updater.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.order_placed.arn
}
