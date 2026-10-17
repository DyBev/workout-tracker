output "lambdas" {
  value = var.lambdas
}

output "aws_lambdas" {
  value = aws_lambda_function.functions
}
