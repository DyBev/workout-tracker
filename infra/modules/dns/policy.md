# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Route53",
            "Effect": "Allow",
            "Action": [
                "route53:GetHostedZone",
                "route53:ListResourceRecordSets",
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": "arn:aws:route53:::hostedzone/<ROUTE53_ZONE_ID>"
        }
    ]
}
```
