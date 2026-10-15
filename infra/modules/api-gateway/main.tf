resource "aws_apigatewayv2_api" "api" {
  name          = "${var.project_name}-preview-${var.environment}-${var.branch_name}"
  protocol_type = "HTTP"

  tags = var.tags

  cors_configuration {
    allow_methods = [
      "GET",
      "POST",
      "PATCH",
      "OPTIONS"
    ]

    allow_headers = [
      "authorization",
      "content-type"
    ]

    max_age = 300
  }
}

resource "aws_apigatewayv2_integration" "backend" {
  for_each = var.lambda_functions

  api_id           = aws_apigatewayv2_api.api.id
  integration_type = "AWS_PROXY"

  integration_method     = "POST"
  integration_uri        = each.value.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "routes" {
  for_each = aws_apigatewayv2_integration.backend

  api_id    = aws_apigatewayv2_api.api.id
  route_key = "${var.functions_list[each.key].api_method} ${var.functions_list[each.key].api_path}"
  target    = "integrations/${each.value.id}"

  authorization_type = "JWT"
  authorizer_id = aws_apigatewayv2_authorizer.authoriser.id
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "api"
  auto_deploy = true
}

resource "aws_apigatewayv2_authorizer" "authoriser" {
  api_id = aws_apigatewayv2_api.api.id
  authorizer_type = "JWT"
  name = "jwt-auth"
  identity_sources = [
    "$request.header.Authorization"
  ]

  jwt_configuration {
    issuer = "https://${var.cognito_module_user_pool_endpoint}"
    audience = [
       var.cognito_module_workout_tracker_client_id
    ]
  }
}

resource "aws_lambda_permission" "api_gateway" {
  for_each = var.lambda_functions

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = each.value.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

