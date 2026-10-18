output "lambdas" {
  value = local.lambda_configs
}

output "aws_lambdas" {
  value = aws_lambda_function.functions
}
