output "bucket_name" {
  description = "Dev branch S3 bucket name."
  value       = module.s3.bucket
}

output "cloudfront_domain_name" {
  description = "CloudFront domain for this dev branch."
  value       = module.cloudfront.domain_name
}
