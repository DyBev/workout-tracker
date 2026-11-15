# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AppBuckets",
            "Effect": "Allow",
            "Action": [
                "s3:CreateBucket",
                "s3:DeleteBucket",
                "s3:ListBucket",
                "s3:GetBucketLocation",
                "s3:GetBucketPolicy",
                "s3:PutBucketPolicy",
                "s3:DeleteBucketPolicy",
                "s3:GetBucketPublicAccessBlock",
                "s3:PutBucketPublicAccessBlock",
                "s3:GetBucketVersioning",
                "s3:PutBucketVersioning",
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::workout-tracker-*",
                "arn:aws:s3:::workout-tracker-*/*"
            ]
        }
    ]
}
```
