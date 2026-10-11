terraform {
  backend "s3" {
    bucket       = "dybev.uk-projects"
    region       = "eu-west-2"
    encrypt      = true
    use_lockfile = true
  }
}

