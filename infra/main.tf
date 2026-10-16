terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region for S3 and supporting resources."
  type        = string
  default     = "eu-west-2"
}

variable "project_name" {
  description = "Project prefix for naming."
  type        = string
  default     = "workout-tracker"
}

variable "environment" {
  description = "Environment name."
  type        = string
  default     = "dev"
}

variable "branch_name" {
  description = "Branch name, e.g. feature/login-page"
  type        = string
}

variable "tags" {
  description = "Extra tags"
  type        = map(string)
  default     = {}
}

data "aws_caller_identity" "current" {}

locals {
  branch_slug = replace(lower(var.branch_name), "[^a-z0-9-]", "-")

  bucket_name = "${var.project_name}-preview-${var.environment}-${local.branch_slug}"
  table_name = "${var.project_name}-preview-${var.environment}-${local.branch_slug}"

  common_tags = merge(
    {
      Project     = var.project_name
      Environment = var.environment
      Branch      = var.branch_name
      ManagedBy   = "OpenTofu"
    },
    var.tags
  )
}

module "cognito" {
  source = "./modules/cognito"
}

module "lambda" {
  source = "./modules/lambda"
  tags = local.common_tags
  table_name = local.table_name
  environment = var.environment
  project_name = var.project_name
  branch_slug = local.branch_slug
  aws_dynamodb_table_arn = aws_dynamodb_table.workouts.arn
}

module "gateway" {
  source = "./modules/api-gateway"
  aws_region = var.aws_region
  branch_name = var.branch_name
  environment = var.environment
  project_name = var.project_name
  tags = local.common_tags
  lambda_functions = module.lambda.aws_lambdas
  functions_list = module.lambda.lambdas
  cognito_module_user_pool_endpoint = module.cognito.user_pool_endpoint
  cognito_module_workout_tracker_client_id = module.cognito.workout_tracker_client_id
}

module "cloudfront" {
  source = "./modules/cloudfront"
  bucket_name = local.bucket_name
  branch_name = var.branch_name
  branch_slug = local.branch_slug
  tags = local.common_tags
  environment = var.environment
  project_name = var.project_name
  gateway_api_endpoint = module.gateway.api_endpoint
  s3_id = module.s3.id
  s3_bucket_regional_domain_name = module.s3.bucket_regional_domain_name
}

module "s3" {
  source = "./modules/s3"
  bucket_name = local.bucket_name
  tags = local.common_tags
  aws_cloudfront_distribution_arn = module.cloudfront.arn
}
