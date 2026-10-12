resource "aws_dynamodb_table" "workouts" {
  name         = local.table_name
  billing_mode = "PAY_PER_REQUEST"

  hash_key  = "userId"
  range_key = "sk"

  attribute {
    name = "userId"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }

  attribute {
    name = "workoutId"
    type = "S"
  }

  global_secondary_index {
    name = "workoutId"
    projection_type = "KEYS_ONLY"
    hash_key = "userId"
    range_key = "workoutId"
  }

  ttl {
    enabled = false
  }

  tags = local.common_tags
}
