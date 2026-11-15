variable "branch_slug" { }
variable "bucket_name" { }
variable "project_name" { }
variable "environment" { }
variable "branch_name" { }
variable "s3_bucket_regional_domain_name" { }
variable "s3_id" { }
variable "gateway_api_endpoint" { }
variable "tags" { }
variable "aliases" { type = list(string) }
variable "domain_certificate_arn" { type = string }

