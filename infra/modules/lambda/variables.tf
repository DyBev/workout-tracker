variable "lambdas" {
  type = map(object({
    zip_path = string
    api_method = string
    api_path = string
  }))

  default = {
    "read-workout" = {
      zip_path = "../../build/read-workout.zip",
      api_path = "/read",
      api_method = "GET",
    }

    "save-workout" = {
      zip_path = "../../build/save-workout.zip",
      api_path = "/save",
      api_method = "POST",
    }
  }
}

variable "project_name" { type = string }
variable "environment" { type = string }
variable "branch_slug" { type = string }
variable "aws_dynamodb_table_arn" { type = string }
variable "table_name" { type = string }
variable "tags" { }
