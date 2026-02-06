# ─────────────────────────────────────────────────────────────────────────────
# CloudRetail – Terraform Variable Values
# Copy this file to terraform.tfvars and fill in sensitive values
# ─────────────────────────────────────────────────────────────────────────────

aws_region   = "us-east-1"
environment  = "prod"
project_name = "cloudretail"

# Cognito (existing pool)
cognito_user_pool_id = "us-east-1_eJvqfLh2p"
cognito_client_id    = "2tkqjdk1i7r7uefcsargsrb3tq"
cognito_client_secret = ""  # fill in

# RDS Aurora
db_master_username = "postgres"
db_master_password = "CHANGE_ME_STRONG_PASSWORD"  # ← CHANGE THIS
db_name            = "cloudretail"

# ECS sizing
ecs_cpu       = 512
ecs_memory    = 1024
desired_count = 1  # use 1 for dev, 2+ for prod

# Amplify (set your GitHub repo URL + personal access token)
github_repo_url = ""  # e.g. "https://github.com/youruser/cloudretail"
github_token    = ""  # GitHub PAT with repo scope
amplify_branch  = "main"
