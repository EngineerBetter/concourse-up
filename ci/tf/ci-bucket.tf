variable "region" {}

provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "ci" {
  bucket        = "concourse-up-ci-artifacts"
  acl           = "private"
  versioning {
    enabled = true
  }
}