package concourse

import (
	"fmt"
	"io"
	"io/ioutil"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Deploy deploys a concourse instance
func Deploy(name, region string,
	terraformClientFactory terraform.ClientFactory,
	boshInitClientFactory bosh.BoshInitClientFactory,
	configClient config.IClient,
	stdout, stderr io.Writer) error {
	config, err := configClient.LoadOrCreate(name)
	if err != nil {
		return err
	}

	userIP, err := util.FindUserIP()
	if err != nil {
		return err
	}

	config.SourceAccessIP = userIP
	stdout.Write([]byte(fmt.Sprintf(
		"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))

	terraformFile, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}

	terraformClient, err := terraformClientFactory(terraformFile, stdout, stderr)
	if err != nil {
		return err
	}

	defer func() {
		err = terraformClient.Cleanup()
	}()

	err = terraformClient.Apply()
	if err != nil {
		return err
	}

	metadata, err := terraformClient.Output()
	if err != nil {
		return err
	}

	boshInitClient, manifestBytes, err := createBoshInitClient(config, metadata, boshInitClientFactory, stdout, stderr)
	if err != nil {
		return err
	}
	if err = configClient.StoreAsset(config.Project, "director.yml", manifestBytes); err != nil {
		return err
	}

	boshStateBytes, err := boshInitClient.Deploy()
	if err != nil {
		return err
	}

	if err = configClient.StoreAsset(config.Project, "director-state.json", boshStateBytes); err != nil {
		return err
	}

	stdout.Write([]byte(fmt.Sprintf(
		"\nDEPLOY SUCCESSFUL. Bosh connection credentials:\n\tIP Address: %s\n\tUsername: %s\n\tPassword: %s\n\n",
		metadata.DirectorPublicIP,
		config.DirectorUsername,
		config.DirectorPassword,
	)))

	return nil
}

func createBoshInitClient(config *config.Config, metadata *terraform.Metadata, boshInitClientFactory bosh.BoshInitClientFactory, stdout, stderr io.Writer) (bosh.IBoshInitClient, []byte, error) {
	keyFile, err := ioutil.TempFile("", config.Deployment)
	if err != nil {
	}

	if _, err = keyFile.WriteString(config.PrivateKey); err != nil {
	}
	if err = keyFile.Sync(); err != nil {
	}

	manifestBytes, err := bosh.GenerateAWSDirectorManifest(config, keyFile.Name(), metadata)
	if err != nil {
	}

	manifestFile, err := ioutil.TempFile("", config.Deployment)
	if err != nil {
	}

	if _, err = manifestFile.Write(manifestBytes); err != nil {
	}
	if err = manifestFile.Sync(); err != nil {
	}

	return boshInitClientFactory(manifestFile.Name(), stdout, stderr), manifestBytes, nil
}

const template = `
terraform {
	backend "s3" {
		bucket = "<% .ConfigBucket %>"
		key    = "<% .TFStatePath %>"
		region = "<% .Region %>"
	}
}

variable "rds_instance_class" {
  type = "string"
	default = "<% .RDSInstanceClass %>"
}

variable "rds_instance_username" {
  type = "string"
	default = "<% .RDSUsername %>"
}

variable "rds_instance_password" {
  type = "string"
	default = "<% .RDSPassword %>"
}

variable "source_access_ip" {
  type = "string"
	default = "<% .SourceAccessIP %>"
}

variable "region" {
  type = "string"
	default = "<% .Region %>"
}

variable "availability_zone" {
  type = "string"
	default = "<% .AvailabilityZone %>"
}

variable "deployment" {
  type = "string"
	default = "<% .Deployment %>"
}

variable "rds_default_database_name" {
  type = "string"
	default = "<% .RDSDefaultDatabaseName %>"
}

variable "public_key" {
  type = "string"
	default = "<% .PublicKey %>"
}

variable "project" {
  type = "string"
	default = "<% .Project %>"
}

provider "aws" {
	region = "<% .Region %>"
}

resource "aws_key_pair" "default" {
	key_name_prefix = "${var.deployment}"
	public_key      = "${var.public_key}"
}

resource "aws_s3_bucket" "blobstore" {
  bucket        = "${var.deployment}-blobstore"
  force_destroy = true

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_iam_user" "blobstore" {
  name = "${var.deployment}-blobstore"
}

resource "aws_iam_access_key" "blobstore" {
  user = "${var.deployment}-blobstore"
  depends_on = ["aws_iam_user.blobstore"]
}

resource "aws_iam_user_policy" "blobstore" {
  name = "${var.deployment}-blobstore"
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
  name = "${var.deployment}-bosh"
}

resource "aws_iam_access_key" "bosh" {
  user = "${var.deployment}-bosh"
  depends_on = ["aws_iam_user.bosh"]
}

resource "aws_iam_user_policy" "bosh" {
  name = "${var.deployment}-bosh"
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

resource "aws_subnet" "director" {
  vpc_id                  = "${aws_vpc.default.id}"
  availability_zone       = "${var.availability_zone}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags {
    Name = "${var.deployment}-director"
    concourse-up-project = "${var.project}"
    concourse-up-component = "bosh"
  }
}

resource "aws_eip" "director" {
  vpc = true
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
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32"]
  }

  ingress {
    from_port   = 6868
    to_port     = 6868
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32"]
  }

  ingress {
    from_port   = 25555
    to_port     = 25555
    protocol    = "tcp"
    cidr_blocks = ["${var.source_access_ip}/32"]
  }

  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "udp"
    self      = true
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
    concourse-up-component = "concourse"
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
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

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
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

resource "aws_route_table_association" "rds_c" {
  subnet_id      = "${aws_subnet.rds_c.id}"
  route_table_id = "${aws_route_table.rds.id}"
}

resource "aws_subnet" "rds_a" {
  vpc_id            = "${aws_vpc.default.id}"
  availability_zone = "${var.region}a"
  cidr_block        = "10.0.4.0/24"

  tags {
    Name = "${var.deployment}-rds-a"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_subnet" "rds_b" {
  vpc_id            = "${aws_vpc.default.id}"
  availability_zone = "${var.region}b"
  cidr_block        = "10.0.5.0/24"

  tags {
    Name = "${var.deployment}-rds-b"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_subnet" "rds_c" {
  vpc_id            = "${aws_vpc.default.id}"
  availability_zone = "${var.region}c"
  cidr_block        = "10.0.6.0/24"

  tags {
    Name = "${var.deployment}-rds-c"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_db_subnet_group" "default" {
  name       = "${var.deployment}"
  subnet_ids = ["${aws_subnet.rds_a.id}", "${aws_subnet.rds_b.id}", "${aws_subnet.rds_c.id}"]

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

resource "aws_db_instance" "default" {
  allocated_storage      = 10
  port                   = 5432
  engine                 = "postgres"
  instance_class         = "${var.rds_instance_class}"
  engine_version         = "9.6.1"
  name                   = "${var.rds_default_database_name}"
  username               = "${var.rds_instance_username}"
  password               = "${var.rds_instance_password}"
  publicly_accessible    = false
  multi_az               = true
  vpc_security_group_ids = ["${aws_security_group.rds.id}"]
  db_subnet_group_name   = "${aws_db_subnet_group.default.name}"
  skip_final_snapshot    = true

  tags {
    Name = "${var.deployment}"
    concourse-up-project = "${var.project}"
    concourse-up-component = "rds"
  }
}

output "director_key_pair" {
  value = "${aws_key_pair.default.key_name}"
}

output "director_public_ip" {
  value = "${aws_eip.director.public_ip}"
}

output "director_security_group_id" {
  value = "${aws_security_group.director.id}"
}

output "vms_security_group_id" {
  value = "${aws_security_group.vms.id}"
}

output "director_subnet_id" {
  value = "${aws_subnet.director.id}"
}

output "blobstore_bucket" {
  value = "${aws_s3_bucket.blobstore.id}"
}

output "blobstore_user_access_key_id" {
  value = "${aws_iam_access_key.blobstore.id}"
}

output "blobstore_user_secret_access_key" {
  value = "${aws_iam_access_key.blobstore.secret}"
}

output "bosh_user_access_key_id" {
  value = "${aws_iam_access_key.bosh.id}"
}

output "bosh_user_secret_access_key" {
  value = "${aws_iam_access_key.bosh.secret}"
}

output "bosh_db_username" {
  value = "${var.rds_instance_username}"
}

output "bosh_db_password" {
  value = "${var.rds_instance_password}"
}

output "bosh_db_port" {
  value = "${aws_db_instance.default.port}"
}

output "bosh_db_address" {
  value = "${aws_db_instance.default.address}"
}
`
