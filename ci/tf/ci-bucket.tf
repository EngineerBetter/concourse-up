variable "region" {}
variable "bucket-name" {
  default = "concourse-up-ci-artifacts"
}

provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "ci" {
  bucket        = "${var.bucket-name}"
  acl           = "private"
  versioning {
    enabled = true
  }
}