output "arn" {
  value = aws_cloudfront_distribution.web.arn
}

output "domain_name" {
  value = aws_cloudfront_distribution.web.domain_name
}

output "id" {
  value = aws_cloudfront_distribution.web.id
}

output "hosted_zone_id" {
  value = aws_cloudfront_distribution.web.hosted_zone_id
}
