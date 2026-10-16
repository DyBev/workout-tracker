output "bucket_name" {
  description = "Dev branch S3 bucket name."
  value       = module.s3.bucket
}

output "cloudfront_domain_name" {
  description = "CloudFront domain for this dev branch."
  value       = module.cloudfront.domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID for invalidations."
  value       = module.cloudfront.id
}
