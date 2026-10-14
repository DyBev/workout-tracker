resource "aws_cognito_user_pool" "this" {
    auto_verified_attributes  = [
        "email",
    ]

    deletion_protection       = "ACTIVE"

    mfa_configuration         = "OFF"
    name                      = "workout-tracker"

    tags                      = {}
    tags_all                  = {}
    user_pool_tier            = "ESSENTIALS"
    username_attributes       = [
        "email",
    ]

    account_recovery_setting {
        recovery_mechanism {
            name     = "verified_email"
            priority = 1
        }
        recovery_mechanism {
            name     = "verified_phone_number"
            priority = 2
        }
    }

    admin_create_user_config {
        allow_admin_create_user_only = false
    }

    email_configuration {
        email_sending_account = "COGNITO_DEFAULT"
    }

    password_policy {
        minimum_length                   = 8
        password_history_size            = 0
        require_lowercase                = true
        require_numbers                  = true
        require_symbols                  = true
        require_uppercase                = true
        temporary_password_validity_days = 7
    }

    sign_in_policy {
        allowed_first_auth_factors = [
            "PASSWORD",
        ]
    }

    username_configuration {
        case_sensitive = false
    }

    verification_message_template {
        default_email_option = "CONFIRM_WITH_CODE"
    }
}

resource "aws_cognito_user_pool_client" "workout_tracker" {
  name = "workout-tracker"
  user_pool_id = aws_cognito_user_pool.this.id

  access_token_validity = 60
  id_token_validity = 60
  refresh_token_validity = 90

  token_validity_units {
    access_token = "minutes"
    id_token = "minutes"
    refresh_token = "days"
  }
}
