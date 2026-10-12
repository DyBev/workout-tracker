terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
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
