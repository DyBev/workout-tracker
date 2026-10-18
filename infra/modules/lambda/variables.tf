variable "aws_dynamodb_table_arn" { type = string }
variable "branch_slug" { type = string }
variable "environment" { type = string }
variable "project_name" { type = string }
variable "source_path_relative" {
  default = "../.."
  type = string
}
variable "table_name" { type = string }
variable "tags" { }

locals {
  lambda_configs = {
    "read-workout" = {
      api_path   = "/read"
      api_method = "GET"
      authentication = true
    }

    "save-workout" = {
      api_path   = "/save"
      api_method = "POST"
      authentication = true
    }

    "exercises-read" = {
      api_path   = "/exercises/read"
      api_method = "GET"
      authentication = true
    }

    "exercises-save" = {
      api_path   = "/exercises/save"
      api_method = "POST"
      authentication = true
    }

    "exercises-archive" = {
      api_path   = "/exercises/archive"
      api_method = "PATCH"
      authentication = true
    }

    "register-interest" = {
      api_path   = "/register-interest"
      api_method = "POST"
      authentication = false
    }

    "update-body-weight" = {
      api_path   = "/workout/update/bodyweight"
      api_method = "PATCH"
      authentication = true
    }
  }

  lambdas = {
    for name, config in local.lambda_configs : name => merge(config, {
      binary_path = "${var.source_path_relative}/build/${name}/bootstrap"
      zip_path    = "${var.source_path_relative}/build/${name}.zip"
    })
  }
}
