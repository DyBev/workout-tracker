variable "cognito_module_workout_tracker_client_id" {
  description = "AWS cognito module user pool endpoint"
  type = string
}

variable "cognito_module_user_pool_endpoint" {
  description = "AWS cognito module user pool endpoint"
  type = string
}

variable "lambda_functions" {
  description = "AWS lambda functions module"

  type = map(object({
    function_name = string
    invoke_arn = string
    id = string
  }))
}

variable "functions_list" {
  description = "local lambda functions definition"

  type = map(object({
    api_method = string
    api_path = string
    authentication = bool
  }))
}

variable "aws_region" {
  description = "AWS region for S3 and supporting resources."
  type        = string
}

variable "project_name" {
  description = "Project prefix for naming."
  type        = string
}

variable "environment" {
  description = "Environment name."
  type        = string
}

variable "branch_name" {
  description = "Branch name, e.g. feature/login-page"
  type        = string
}

variable "tags" {
  description = "Extra tags"
  type        = map(string)
}
