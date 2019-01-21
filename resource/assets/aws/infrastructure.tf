terraform {
	backend "s3" {
		bucket = "{{ .ConfigBucket }}"
		key    = "{{ .TFStatePath }}"
		region = "{{ .Region }}"
	}
}

data "aws_availability_zones" "available" {
  state = "available"
}

variable "rds_instance_class" {
  type = "string"
	default = "{{ .RDSInstanceClass }}"
}

variable "rds_instance_username" {
  type = "string"
	default = "{{ .RDSUsername }}"
}

variable "rds_instance_password" {
  type = "string"
	default = "{{ .RDSPassword }}"
}

variable "source_access_ip" {
  type = "string"
	default = "{{ .SourceAccessIP }}"
}

variable "region" {
  type = "string"
	default = "{{ .Region }}"
}

variable "availability_zone" {
  type = "string"
	default = "{{ .AvailabilityZone }}"
}

variable "deployment" {
  type = "string"
	default = "{{ .Deployment }}"
}

variable "rds_default_database_name" {
  type = "string"
	default = "{{ .RDSDefaultDatabaseName }}"
}

variable "public_key" {
  type = "string"
	default = "{{ .PublicKey }}"
}

variable "project" {
  type = "string"
	default = "{{ .Project }}"
}

variable "multi_az_rds" {
  type = "string"
  default = {{if .MultiAZRDS }}true{{else}}false{{end}}
}

{{if .HostedZoneID }}
variable "hosted_zone_id" {
  type = "string"
  default = "{{ .HostedZoneID }}"
}

variable "hosted_zone_record_prefix" {
  type = "string"
  default = "{{ .HostedZoneRecordPrefix }}"
}
{{end}}

provider "aws" {
	region = "{{ .Region }}"
}

resource "aws_key_pair" "default" {
	key_name_prefix = "${var.deployment}"
	public_key      = "${var.public_key}"
}

resource "aws_s3_bucket" "blobstore" {
  bucket        = "${var.deployment}-{{ .Namespace }}-blobstore"
  force_destroy = true
  region = "{{ .Region }}"

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_iam_user" "blobstore" {
  name = "${var.deployment}-{{ .Namespace }}-blobstore"
}

resource "aws_iam_access_key" "blobstore" {
  user = "${var.deployment}-{{ .Namespace }}-blobstore"
  depends_on = ["aws_iam_user.blobstore"]
}

resource "aws_iam_user_policy" "blobstore" {
  name = "${var.deployment}-{{ .Namespace }}-blobstore"
  user = "${aws_iam_user.blobstore.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::${aws_s3_bucket.blobstore.id}",
        "arn:aws:s3:::${aws_s3_bucket.blobstore.id}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_user" "bosh" {
  name = "${var.deployment}-${var.region}-bosh"
}

resource "aws_iam_access_key" "bosh" {
  user = "${var.deployment}-${var.region}-bosh"
  depends_on = ["aws_iam_user.bosh"]
}

resource "aws_iam_user_policy" "bosh" {
  name = "${var.deployment}-${var.region}-bosh"
  user = "${aws_iam_user.bosh.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "elasticloadbalancing:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_route" "internet_access" {
  route_table_id         = "${aws_vpc.default.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.default.id}"
}

 resource "aws_nat_gateway" "default" {
  allocation_id = "${aws_eip.nat.id}"
  subnet_id     = "${aws_subnet.public.id}"

  depends_on = ["aws_internet_gateway.default"]

    tags {
    name = "${var.deployment}"
    concourse-up-project = "${var.project}"
  }
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.default.id}"

  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.default.id}"
  }

  tags {
    Name = "${var.deployment}-private"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = "${aws_vpc.default.id}"
  availability_zone       = "${var.availability_zone}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags {
    Name = "${var.deployment}-public"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_subnet" "private" {
  vpc_id                  = "${aws_vpc.default.id}"
  availability_zone       = "${var.availability_zone}"
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = false

  tags {
    Name = "${var.deployment}-private"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id      = "${aws_subnet.private.id}"
  route_table_id = "${aws_route_table.private.id}"
}

{{if .HostedZoneID }}
resource "aws_route53_record" "concourse" {
  zone_id = "${var.hosted_zone_id}"
  name    = "${var.hosted_zone_record_prefix}"
  ttl     = "60"
  type    = "A"
  records = ["${aws_eip.atc.public_ip}"]
}
{{end}}

resource "aws_eip" "director" {
  vpc = true
  depends_on = ["aws_internet_gateway.default"]

    tags {
    name = "${var.deployment}-director"
    concourse-up-project = "${var.project}"
  }
}

resource "aws_eip" "atc" {
  vpc = true
  depends_on = ["aws_internet_gateway.default"]

    tags {
    name = "${var.deployment}-atc"
    concourse-up-project = "${var.project}"
  }
}

resource "aws_eip" "nat" {
  vpc = true
  depends_on = ["aws_internet_gateway.default"]

    tags {
    name = "${var.deployment}-nat"
    concourse-up-project = "${var.project}"
  }
}

resource "aws_security_group" "director" {
  name        = "${var.deployment}-director"
  description = "Concourse UP Default BOSH security group"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.deployment}-director"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }

  ingress {
    from_port   = 6868
    to_port     = 6868
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
  }

  ingress {
    from_port   = 25555
    to_port     = 25555
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32", "${aws_nat_gateway.default.public_ip}/32"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "vms" {
  name        = "${var.deployment}-vms"
  description = "Concourse UP VMs security group"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.deployment}-vms"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }

  ingress {
    from_port   = 6868
    to_port     = 6868
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 4222
    to_port     = 4222
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }


  ingress {
    from_port   = 25250
    to_port     = 25250
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 25555
    to_port     = 25555
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 25777
    to_port     = 25777
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 53
    to_port     = 53
    protocol    = "udp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 2222
    to_port     = 2222
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 5555
    to_port     = 5555
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 7777
    to_port     = 7777
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 7788
    to_port     = 7788
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 7799
    to_port     = 7799
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "icmp"
    cidr_blocks = ["10.0.0.0/16"]
  }
  ingress {
    from_port = 22
    to_port   = 22
    self      = true
    protocol  = "tcp"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.deployment}-rds"
  description = "Concourse UP RDS security group"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.deployment}-rds"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_security_group" "atc" {
  name        = "${var.deployment}-atc"
  description = "Concourse UP ATC security group"
  vpc_id      = "${aws_vpc.default.id}"
  depends_on = ["aws_eip.nat", "aws_eip.atc"]

  tags {
    Name = "${var.deployment}-atc"
    concourse-up-project = "${var.project}"
    concourse-up-component = "concourse"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    security_groups = ["${aws_security_group.vms.id}", "${aws_security_group.director.id}"]
    cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", {{ .AllowIPs }}]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", {{ .AllowIPs }}]
  }

  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["${aws_eip.nat.public_ip}/32", {{ .AllowIPs }}]
  }

  ingress {
    from_port   = 8844
    to_port     = 8844
    protocol    = "tcp"
    cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", {{ .AllowIPs }}]
  }

  ingress {
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    cidr_blocks = ["${aws_eip.nat.public_ip}/32", "${aws_eip.atc.public_ip}/32", {{ .AllowIPs }}]
  }
}

resource "aws_route_table" "rds" {
  vpc_id = "${aws_vpc.default.id}"

  tags {
    Name = "${var.deployment}-rds"
    concourse-up-project = "${var.project}"
    concourse-up-component = "concourse"
  }
}

resource "aws_route_table_association" "rds_a" {
  subnet_id      = "${aws_subnet.rds_a.id}"
  route_table_id = "${aws_route_table.rds.id}"
}

resource "aws_route_table_association" "rds_b" {
  subnet_id      = "${aws_subnet.rds_b.id}"
  route_table_id = "${aws_route_table.rds.id}"
}

resource "aws_subnet" "rds_a" {
  vpc_id            = "${aws_vpc.default.id}"
  availability_zone = "${element(sort(data.aws_availability_zones.available.names),0)}"
  cidr_block        = "10.0.4.0/24"

  tags {
    Name = "${var.deployment}-rds-a"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_subnet" "rds_b" {
  vpc_id            = "${aws_vpc.default.id}"
  availability_zone = "${element(sort(data.aws_availability_zones.available.names),1)}"
  cidr_block        = "10.0.5.0/24"

  tags {
    Name = "${var.deployment}-rds-b"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_db_subnet_group" "default" {
  name       = "${var.deployment}"
  subnet_ids = ["${aws_subnet.rds_a.id}", "${aws_subnet.rds_b.id}"]

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_db_instance" "default" {
  allocated_storage      = 10
  apply_immediately      = true
  port                   = 5432
  engine                 = "postgres"
  instance_class         = "${var.rds_instance_class}"
  engine_version         = "9.6.6"
  name                   = "${var.rds_default_database_name}"
  username               = "${var.rds_instance_username}"
  password               = "${var.rds_instance_password}"
  publicly_accessible    = false
  multi_az               = "${var.multi_az_rds}"
  vpc_security_group_ids = ["${aws_security_group.rds.id}"]
  db_subnet_group_name   = "${aws_db_subnet_group.default.name}"
  skip_final_snapshot    = true
  storage_type           = "gp2"
  lifecycle {
    ignore_changes = ["allocated_storage"]
  }
  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

output "vpc_id" {
  value = "${aws_vpc.default.id}"
}

output "source_access_ip" {
  value = "${var.source_access_ip}"
}

output "director_key_pair" {
  value = "${aws_key_pair.default.key_name}"
}

output "director_public_ip" {
  value = "${aws_eip.director.public_ip}"
}

output "atc_public_ip" {
  value = "${aws_eip.atc.public_ip}"
}

output "director_security_group_id" {
  value = "${aws_security_group.director.id}"
}

output "vms_security_group_id" {
  value = "${aws_security_group.vms.id}"
}

output "atc_security_group_id" {
  value = "${aws_security_group.atc.id}"
}

output "nat_gateway_ip" {
  value = "${aws_nat_gateway.default.public_ip}"
}

output "public_subnet_id" {
  value = "${aws_subnet.public.id}"
}

output "private_subnet_id" {
  value = "${aws_subnet.private.id}"
}

output "blobstore_bucket" {
  value = "${aws_s3_bucket.blobstore.id}"
}

output "blobstore_user_access_key_id" {
  value = "${aws_iam_access_key.blobstore.id}"
}

output "blobstore_user_secret_access_key" {
  value     = "${aws_iam_access_key.blobstore.secret}"
  sensitive = true
}

output "bosh_user_access_key_id" {
  value = "${aws_iam_access_key.bosh.id}"
}

output "bosh_user_secret_access_key" {
  value     = "${aws_iam_access_key.bosh.secret}"
  sensitive = true
}

output "bosh_db_port" {
  value = "${aws_db_instance.default.port}"
}

output "bosh_db_address" {
  value = "${aws_db_instance.default.address}"
}
