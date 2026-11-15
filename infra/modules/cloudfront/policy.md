# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "CloudFront",
            "Effect": "Allow",
            "Action": [
                "cloudfront:CreateDistribution",
                "cloudfront:GetDistribution",
                "cloudfront:GetDistributionConfig",
                "cloudfront:UpdateDistribution",
                "cloudfront:DeleteDistribution",
                "cloudfront:CreateInvalidation",
                "cloudfront:CreateOriginAccessControl",
                "cloudfront:GetOriginAccessControl",
                "cloudfront:UpdateOriginAccessControl",
                "cloudfront:DeleteOriginAccessControl",
                "cloudfront:TagResource",
                "cloudfront:UntagResource",
                "cloudfront:ListTagsForResource"
            ],
            "Resource": "*"
        }
    ]
}
```
