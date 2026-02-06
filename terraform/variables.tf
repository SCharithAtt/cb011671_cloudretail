# ─────────────────────────────────────────────────────────────────────────────
# Variables
# ─────────────────────────────────────────────────────────────────────────────

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "prod"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "cloudretail"
}

# ── Cognito ──────────────────────────────────────────────────────────────────

variable "cognito_user_pool_id" {
  description = "Existing Cognito User Pool ID"
  type        = string
  default     = "us-east-1_eJvqfLh2p"
}

variable "cognito_client_id" {
  description = "Existing Cognito App Client ID"
  type        = string
  default     = "2tkqjdk1i7r7uefcsargsrb3tq"
}

variable "cognito_client_secret" {
  description = "Cognito App Client secret (leave empty for public client)"
  type        = string
  default     = ""
  sensitive   = true
}

# ── RDS ──────────────────────────────────────────────────────────────────────

variable "db_master_username" {
  description = "Master username for RDS PostgreSQL"
  type        = string
  default     = "postgres"
}

variable "db_master_password" {
  description = "Master password for RDS PostgreSQL"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "cloudretail"
}

# ── ECS ──────────────────────────────────────────────────────────────────────

variable "ecs_cpu" {
  description = "ECS task CPU units (1024 = 1 vCPU)"
  type        = number
  default     = 512
}

variable "ecs_memory" {
  description = "ECS task memory (MiB)"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Desired number of ECS tasks per service"
  type        = number
  default     = 2
}

# ── EC2 for ECS ──────────────────────────────────────────────────────────────

variable "ecs_instance_type" {
  description = "EC2 instance type for ECS cluster"
  type        = string
  default     = "t3.micro"
}

variable "ecs_asg_min_size" {
  description = "Minimum number of EC2 instances in Auto Scaling Group"
  type        = number
  default     = 2
}

variable "ecs_asg_max_size" {
  description = "Maximum number of EC2 instances in Auto Scaling Group"
  type        = number
  default     = 6
}

variable "ecs_asg_desired_capacity" {
  description = "Desired number of EC2 instances in Auto Scaling Group"
  type        = number
  default     = 2
}

variable "ecs_key_pair_name" {
  description = "Optional EC2 key pair name for SSH access to ECS instances"
  type        = string
  default     = ""
}

# ── Amplify ──────────────────────────────────────────────────────────────────

variable "github_repo_url" {
  description = "GitHub repository URL for Amplify"
  type        = string
  default     = ""
}

variable "github_token" {
  description = "GitHub personal access token for Amplify"
  type        = string
  default     = ""
  sensitive   = true
}

variable "amplify_branch" {
  description = "Git branch for Amplify deployment"
  type        = string
  default     = "main"
}
