data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.project_name}-preview-${var.environment}-${var.branch_slug}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy_attachment" "lambda_logging" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "lambda_dynamodb" {
  statement {
    effect = "Allow"

    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:PutItem",
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:Scan",
      "dynamodb:Query",
      "dynamodb:UpdateItem",
    ]

    resources = [
      var.aws_dynamodb_table_arn,
      "${var.aws_dynamodb_table_arn}/index/*"
    ]
  }
}

resource "aws_iam_role_policy" "lambda_dynamodb" {
  name   = "${var.project_name}-${var.environment}-lambda-dynamodb"
  role   = aws_iam_role.lambda.id
  policy = data.aws_iam_policy_document.lambda_dynamodb.json
}

resource "aws_lambda_function" "functions" {
  for_each = local.lambdas

  function_name = "${var.project_name}-preview-${var.environment}-${var.branch_slug}-${each.key}"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"

  filename         = each.value.zip_path
  source_code_hash = filebase64sha256(each.value.binary_path)

  architectures = [ "arm64" ]
  memory_size = 1024

  environment {
    variables = merge({
      TABLE_NAME = var.table_name
    })
  }

  tags = var.tags
}

