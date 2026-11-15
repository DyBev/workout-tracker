# AWS IAM policy
This is the policy required for AWS, if you are attaching this to an existing policy you only need the content contained in the "Statement" array.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ResourceGroups",
            "Effect": "Allow",
            "Action": [
                "resource-groups:CreateGroup",
                "resource-groups:GetGroup",
                "resource-groups:UpdateGroup",
                "resource-groups:DeleteGroup",
                "resource-groups:Tag",
                "resource-groups:Untag"
            ],
            "Resource": "*"
        }
    ]
}
```
