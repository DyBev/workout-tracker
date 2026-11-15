# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "StateBucket",
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::<state-bucket-name>"
        },
        {
            "Sid": "StateObjects",
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
            "s3:PutObject",
            "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::<state-bucket-name>/workout-tracker/*"
            ]
        }
    ]
}
```
