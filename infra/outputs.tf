output "bucket_name" {
  description = "Dev branch S3 bucket name."
  value       = aws_s3_bucket.web.bucket
}

output "cloudfront_domain_name" {
  description = "CloudFront domain for this dev branch."
  value       = aws_cloudfront_distribution.web.domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID for invalidations."
  value       = aws_cloudfront_distribution.web.id
}
