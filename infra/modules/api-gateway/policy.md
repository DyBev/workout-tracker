# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
          "Sid": "ApiGatewayV2",
          "Effect": "Allow",
          "Action": ["apigateway:GET","apigateway:POST","apigateway:PATCH","apigateway:DELETE"],
          "Resource": "*"
        }
    ]
}
```
