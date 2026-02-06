# ─────────────────────────────────────────────────────────────────────────────
# DynamoDB Tables (product_service)
# ─────────────────────────────────────────────────────────────────────────────

resource "aws_dynamodb_table" "products" {
  name         = "Products"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "productId"

  attribute {
    name = "productId"
    type = "S"
  }

  attribute {
    name = "sellerId"
    type = "S"
  }

  global_secondary_index {
    name            = "SellerIdIndex"
    hash_key        = "sellerId"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = { Name = "Products" }
}

resource "aws_dynamodb_table" "reviews" {
  name         = "Reviews"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "reviewId"

  attribute {
    name = "reviewId"
    type = "S"
  }

  attribute {
    name = "productId"
    type = "S"
  }

  global_secondary_index {
    name            = "ProductIdIndex"
    hash_key        = "productId"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = { Name = "Reviews" }
}
