# ─────────────────────────────────────────────────────────────────────────────
# Cognito – Reference existing user pool
# ─────────────────────────────────────────────────────────────────────────────

data "aws_cognito_user_pools" "main" {
  name = "cb011671_cloudretail"
}

# Optionally create a Cognito User Pool if one doesn't already exist
# Uncomment the block below and comment out the data source above to create new

# resource "aws_cognito_user_pool" "main" {
#   name = "${local.name}-users"
#
#   auto_verified_attributes = ["email"]
#   username_attributes      = ["email"]
#
#   password_policy {
#     minimum_length    = 8
#     require_uppercase = true
#     require_lowercase = true
#     require_numbers   = true
#     require_symbols   = false
#   }
#
#   schema {
#     name                = "email"
#     attribute_data_type = "String"
#     required            = true
#     mutable             = true
#   }
#
#   schema {
#     name                = "name"
#     attribute_data_type = "String"
#     required            = true
#     mutable             = true
#   }
#
#   tags = { Name = "${local.name}-user-pool" }
# }
#
# resource "aws_cognito_user_pool_client" "main" {
#   name         = "${local.name}-client"
#   user_pool_id = aws_cognito_user_pool.main.id
#
#   generate_secret           = true
#   explicit_auth_flows       = ["ALLOW_ADMIN_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]
#   allowed_oauth_flows       = ["code"]
#   allowed_oauth_scopes      = ["openid", "email", "profile"]
#   allowed_oauth_flows_user_pool_client = true
#   callback_urls             = ["http://${aws_lb.main.dns_name}/callback"]
#   logout_urls               = ["http://${aws_lb.main.dns_name}/"]
#   supported_identity_providers = ["COGNITO"]
# }
#
# resource "aws_cognito_user_pool_domain" "main" {
#   domain       = local.name
#   user_pool_id = aws_cognito_user_pool.main.id
# }
