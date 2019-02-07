terraform {
  backend "gcs" {
    bucket = "<% .ConfigBucket %>"
    region = "<% .Region %>"
  }
}

variable "google_sql_database_instance" {
  type    = "string"
  default = "<% .RDSInstanceClass %>"
}

variable "deployment" {
  type = "string"
	default = "<% .Deployment %>"
}

variable "project" {
  type = "string"
	default = "<% .Project %>"
}

provider "google" {
  region = "<% .Region %>"
  project = "concourse-up"
}

resource "google_storage_bucket" "blobs" {
  name          = "${var.deployment}-<% .Namespace %>-blobs"
  storage_class = "REGIONAL"
  location      = "<% .Region %>"
  force_destroy = true

  labels = {
    Name                   = "${var.deployment}"
    concourse-up-project   = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "google_service_account" "blobs" {
  account_id   = "${var.deployment}-<% .Namespace %>-blobs"
  display_name = "${var.deployment}-<% .Namespace %>-blobs"
}

resource "google_service_account_iam_policy" "blobs" {
  service_account_id = "${google_service_account.blobs.account_id}"
  policy_data = "${data.google_iam_policy.blobs.policy_data}"
}

data "google_iam_policy" "blobs" {
  binding {
    role = "roles/storage.admin"

    members = [
      "serviceAccount:${google_service_account.blobs.email}",
    ]
  }
}

resource "google_service_account_key" "blobs" {
  service_account_id = "${google_service_account.blobs.account_id}"
}

resource "google_service_account" "bosh" {
  account_id   = "${var.deployment}-<% .Namespace %>-bosh"
  display_name = "${var.deployment}-<% .Namespace %>-bosh"
}

resource "google_service_account_iam_policy" "bosh" {
  service_account_id = "${google_service_account.bosh.account_id}"
  policy_data = "${data.google_iam_policy.bosh.policy_data}"
}

data "google_iam_policy" "bosh" {
  binding {
    role = "roles/compute.loadBalancerAdmin"

    members = [
      "serviceAccount:${google_service_account.bosh.email}",
    ]
  }
  binding {
    role = "roles/compute.instanceAdmin.v1"

    members = [
      "serviceAccount:${google_service_account.bosh.email}",
    ]
  }
}

resource "google_service_account_key" "bosh" {
  service_account_id = "${google_service_account.bosh.account_id}"
}

resource "google_compute_network" "default" {
  name                    = "${var.deployment}"
  ipv4_range = "10.0.0.0/16"
}

resource "google_compute_subnetwork" "public" {
  name          = "${var.deployment}-public}"
  ip_cidr_range = "10.0.0.0/24"
  region        = "<% .Region %>"
  network       = "${google_compute_network.default.self_link}"
}

resource "google_compute_subnetwork" "private" {
  name          = "${var.deployment}-private}"
  ip_cidr_range = "10.0.1.0/24"
  region        = "<% .Region %>"
  network       = "${google_compute_network.default.self_link}"
}