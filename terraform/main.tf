# ─────────────────────────────────────────────────────────────────────────────
# CloudRetail – Terraform Root
# ─────────────────────────────────────────────────────────────────────────────

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Remote state – uncomment when S3 backend is ready
  # backend "s3" {
  #   bucket         = "cloudretail-terraform-state"
  #   key            = "prod/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "cloudretail-tf-lock"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "CloudRetail"
      ManagedBy   = "Terraform"
      Environment = var.environment
    }
  }
}

# ── Data Sources ─────────────────────────────────────────────────────────────

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_availability_zones" "available" { state = "available" }

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
  name       = "cloudretail"
  azs        = slice(data.aws_availability_zones.available.names, 0, 2)
}
