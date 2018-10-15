terraform {
  backend "gcs" {
    bucket = "<% .ConfigBucket %>"
    region = "<% .Region %>"
  }
}

# variable "rds_instance_class" {
#   type = "string"
# 	default = "<% .RDSInstanceClass %>"
# }

variable "google_sql_database_instance" {
  type    = "string"
  default = "<% .RDSInstanceClass %>"
}

# variable "rds_instance_username" {
#   type = "string"
# 	default = "<% .RDSUsername %>"
# }
# variable "rds_instance_password" {
#   type = "string"
# 	default = "<% .RDSPassword %>"
# }
# variable "source_access_ip" {
#   type = "string"
# 	default = "<% .SourceAccessIP %>"
# }
# variable "region" {
#   type = "string"
# 	default = "<% .Region %>"
# }
# variable "availability_zone" {
#   type = "string"
# 	default = "<% .AvailabilityZone %>"
# }
variable "deployment" {
  type = "string"
	default = "<% .Deployment %>"
}
# variable "rds_default_database_name" {
#   type = "string"
# 	default = "<% .RDSDefaultDatabaseName %>"
# }
# variable "public_key" {
#   type = "string"
# 	default = "<% .PublicKey %>"
# }
variable "project" {
  type = "string"
	default = "<% .Project %>"
}
# variable "multi_az_rds" {
#   type = "string"
#   default = <%if .MultiAZRDS %>true<%else%>false<%end%>
# }
# <%if .HostedZoneID %>
# variable "hosted_zone_id" {
#   type = "string"
#   default = "<% .HostedZoneID %>"
# }
# variable "hosted_zone_record_prefix" {
#   type = "string"
#   default = "<% .HostedZoneRecordPrefix %>"
# }
# <%end%>

provider "google" {
  region = "<% .Region %>"
  project = "concourse-up"
}

# resource "aws_key_pair" "default" {
# 	key_name_prefix = "${var.deployment}"
# 	public_key      = "${var.public_key}"
# }

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


# resource "aws_subnet" "public" {
#   vpc_id                  = "${aws_vpc.default.id}"
#   availability_zone       = "${var.availability_zone}"
#   cidr_block              = "10.0.0.0/24"
#   map_public_ip_on_launch = true
#   tags {
#     Name = "${var.deployment}-public"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "bosh"
#   }
# }
# resource "aws_subnet" "private" {
#   vpc_id                  = "${aws_vpc.default.id}"
#   availability_zone       = "${var.availability_zone}"
#   cidr_block              = "10.0.1.0/24"
#   map_public_ip_on_launch = false
#   tags {
#     Name = "${var.deployment}-private"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "bosh"
#   }
# }
# resource "aws_route_table_association" "private" {
#   subnet_id      = "${aws_subnet.private.id}"
#   route_table_id = "${aws_route_table.private.id}"
# }
# <%if .HostedZoneID %>
# resource "aws_route53_record" "concourse" {
#   zone_id = "${var.hosted_zone_id}"
#   name    = "${var.hosted_zone_record_prefix}"
#   ttl     = "60"
#   type    = "A"
#   records = ["${aws_eip.atc.public_ip}"]
# }
# <%end%>
# resource "aws_eip" "director" {
#   vpc = true
# }
# resource "aws_eip" "atc" {
#   vpc = true
# }
# resource "aws_eip" "nat" {
#   vpc = true
# }
# resource "aws_security_group" "director" {
#   name        = "${var.deployment}-director"
#   description = "Concourse UP Default BOSH security group"
#   vpc_id      = "${aws_vpc.default.id}"
#   tags {
#     Name = "${var.deployment}-director"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "bosh"
#   }
#   ingress {
#     from_port   = 6868
#     to_port     = 6868
#     protocol    = "tcp"
#     cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
#   }
#   ingress {
#     from_port   = 25555
#     to_port     = 25555
#     protocol    = "tcp"
#     cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
#   }
#   ingress {
#     from_port   = 22
#     to_port     = 22
#     protocol    = "tcp"
#     cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
#   }
#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
# }
# resource "aws_security_group" "vms" {
#   name        = "${var.deployment}-vms"
#   description = "Concourse UP VMs security group"
#   vpc_id      = "${aws_vpc.default.id}"
#   tags {
#     Name = "${var.deployment}-vms"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "bosh"
#   }
#   ingress {
#     from_port   = 6868
#     to_port     = 6868
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 4222
#     to_port     = 4222
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 25250
#     to_port     = 25250
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 25555
#     to_port     = 25555
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 25777
#     to_port     = 25777
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 53
#     to_port     = 53
#     protocol    = "udp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 2222
#     to_port     = 2222
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 5555
#     to_port     = 5555
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 7777
#     to_port     = 7777
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 7788
#     to_port     = 7788
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 7799
#     to_port     = 7799
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "icmp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
#   ingress {
#     from_port = 22
#     to_port   = 22
#     self      = true
#     protocol  = "tcp"
#   }
#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
# }
# resource "aws_security_group" "rds" {
#   name        = "${var.deployment}-rds"
#   description = "Concourse UP RDS security group"
#   vpc_id      = "${aws_vpc.default.id}"
#   tags {
#     Name = "${var.deployment}-rds"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "rds"
#   }
#   ingress {
#     from_port   = 5432
#     to_port     = 5432
#     protocol    = "tcp"
#     cidr_blocks = ["10.0.0.0/16"]
#   }
# }
# resource "aws_security_group" "atc" {
#   name        = "${var.deployment}-atc"
#   description = "Concourse UP ATC security group"
#   vpc_id      = "${aws_vpc.default.id}"
#   depends_on = ["aws_eip.nat", "aws_eip.atc"]
#   tags {
#     Name = "${var.deployment}-atc"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "concourse"
#   }
#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
#   ingress {
#     from_port   = 80
#     to_port     = 80
#     protocol    = "tcp"
#     security_groups = ["${aws_security_group.vms.id}", "${aws_security_group.director.id}"]
#     cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", <% .AllowIPs %>]
#   }
#   ingress {
#     from_port   = 443
#     to_port     = 443
#     protocol    = "tcp"
#     cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", <% .AllowIPs %>]
#   }
#   ingress {
#     from_port   = 3000
#     to_port     = 3000
#     protocol    = "tcp"
#     cidr_blocks = ["${aws_eip.nat.public_ip}/32", <% .AllowIPs %>]
#   }
#   ingress {
#     from_port   = 8844
#     to_port     = 8844
#     protocol    = "tcp"
#     cidr_blocks = ["${aws_eip.nat.public_ip}/32", <% .AllowIPs %>]
#   }
#   ingress {
#     from_port   = 8443
#     to_port     = 8443
#     protocol    = "tcp"
#     cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", <% .AllowIPs %>]
#   }
# }
# resource "aws_route_table" "rds" {
#   vpc_id = "${aws_vpc.default.id}"
#   tags {
#     Name = "${var.deployment}-rds"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "concourse"
#   }
# }
# resource "aws_route_table_association" "rds_a" {
#   subnet_id      = "${aws_subnet.rds_a.id}"
#   route_table_id = "${aws_route_table.rds.id}"
# }
# resource "aws_route_table_association" "rds_b" {
#   subnet_id      = "${aws_subnet.rds_b.id}"
#   route_table_id = "${aws_route_table.rds.id}"
# }
# resource "aws_subnet" "rds_a" {
#   vpc_id            = "${aws_vpc.default.id}"
#   availability_zone = "${var.region}a"
#   cidr_block        = "10.0.4.0/24"
#   tags {
#     Name = "${var.deployment}-rds-a"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "rds"
#   }
# }
# resource "aws_subnet" "rds_b" {
#   vpc_id            = "${aws_vpc.default.id}"
#   availability_zone = "${var.region}b"
#   cidr_block        = "10.0.5.0/24"
#   tags {
#     Name = "${var.deployment}-rds-b"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "rds"
#   }
# }
# resource "aws_db_subnet_group" "default" {
#   name       = "${var.deployment}"
#   subnet_ids = ["${aws_subnet.rds_a.id}", "${aws_subnet.rds_b.id}"]
#   tags {
#     Name = "${var.deployment}"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "rds"
#   }
# }
# resource "aws_db_instance" "default" {
#   allocated_storage      = 10
#   apply_immediately      = true
#   port                   = 5432
#   engine                 = "postgres"
#   instance_class         = "${var.rds_instance_class}"
#   engine_version         = "9.6.6"
#   name                   = "${var.rds_default_database_name}"
#   username               = "${var.rds_instance_username}"
#   password               = "${var.rds_instance_password}"
#   publicly_accessible    = false
#   multi_az               = "${var.multi_az_rds}"
#   vpc_security_group_ids = ["${aws_security_group.rds.id}"]
#   db_subnet_group_name   = "${aws_db_subnet_group.default.name}"
#   skip_final_snapshot    = true
#   storage_type           = "gp2"
#   lifecycle {
#     ignore_changes = ["allocated_storage"]
#   }
#   tags {
#     Name = "${var.deployment}"
#     concourse-up-project = "${var.project}"
#     concourse-up-component = "rds"
#   }
# }
# output "vpc_id" {
#   value = "${aws_vpc.default.id}"
# }
# output "source_access_ip" {
#   value = "${var.source_access_ip}"
# }
# output "director_key_pair" {
#   value = "${aws_key_pair.default.key_name}"
# }
# output "director_public_ip" {
#   value = "${aws_eip.director.public_ip}"
# }
# output "atc_public_ip" {
#   value = "${aws_eip.atc.public_ip}"
# }
# output "director_security_group_id" {
#   value = "${aws_security_group.director.id}"
# }
# output "vms_security_group_id" {
#   value = "${aws_security_group.vms.id}"
# }
# output "atc_security_group_id" {
#   value = "${aws_security_group.atc.id}"
# }
# output "nat_gateway_ip" {
#   value = "${aws_nat_gateway.default.public_ip}"
# }
# output "public_subnet_id" {
#   value = "${aws_subnet.public.id}"
# }
# output "private_subnet_id" {
#   value = "${aws_subnet.private.id}"
# }
# output "blobs_bucket" {
#   value = "${aws_s3_bucket.blobs.id}"
# }
# output "blobs_user_access_key_id" {
#   value = "${aws_iam_access_key.blobs.id}"
# }
# output "blobs_user_secret_access_key" {
#   value     = "${aws_iam_access_key.blobs.secret}"
#   sensitive = true
# }
# output "bosh_user_access_key_id" {
#   value = "${aws_iam_access_key.bosh.id}"
# }
# output "bosh_user_secret_access_key" {
#   value     = "${aws_iam_access_key.bosh.secret}"
#   sensitive = true
# }
# output "bosh_db_port" {
#   value = "${aws_db_instance.default.port}"
# }
# output "bosh_db_address" {
#   value = "${aws_db_instance.default.address}"
# }
