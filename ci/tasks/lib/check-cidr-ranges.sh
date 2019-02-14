#!/bin/bash

function assertNetworkCidrsCorrect() {
  echo "About to check network CIDR ranges"

  expected_public_cidr=${1:-"10.0.0.0/24"}
  expected_private_cidr=${2:-"10.0.1.0/24"}
  expected_vpc_cidr=${3:-"10.0.0.0/16"}
  expected_rds1_cidr=${4:-"10.0.4.0/24"}
  expected_rds2_cidr=${5:-"10.0.5.0/24"}

  if [ "$IAAS" = "AWS" ]; then
    vpc_cidr="$(aws --region "$region" ec2 describe-vpcs --filters "Name=tag:concourse-up-project,Values=${deployment}" | jq -r ".Vpcs[0].CidrBlock")"
    if [ "$vpc_cidr" != "$expected_vpc_cidr" ]; then
      echo "Unexpected VPC CIDR: $vpc_cidr"
      exit 1
    fi

    public_cidr="$(aws --region "$region" ec2 describe-subnets --filters "Name=tag:Name,Values=concourse-up-${deployment}-public" | jq -r ".Subnets[0].CidrBlock")"
    if [ "$public_cidr" != "$expected_public_cidr" ]; then
      echo "Unexpected public subnet CIDR: $public_cidr"
      exit 1
    fi

    private_cidr="$(aws --region "$region" ec2 describe-subnets --filters "Name=tag:Name,Values=concourse-up-${deployment}-private" | jq -r ".Subnets[0].CidrBlock")"
    if [ "$private_cidr" != "$expected_private_cidr" ]; then
      echo "Unexpected private subnet CIDR: $private_cidr"
      exit 1
    fi

    rds1_cidr="$(aws --region "$region" ec2 describe-subnets --filters "Name=tag:Name,Values=concourse-up-${deployment}-rds-a" | jq -r ".Subnets[0].CidrBlock")"
    if [ "$rds1_cidr" != "$expected_rds1_cidr" ]; then
      echo "Unexpected RDS1 subnet CIDR: $rds1_cidr"
      exit 1
    fi

    rds2_cidr="$(aws --region "$region" ec2 describe-subnets --filters "Name=tag:Name,Values=concourse-up-${deployment}-rds-b" | jq -r ".Subnets[0].CidrBlock")"
    if [ "$rds2_cidr" != "$expected_rds2_cidr" ]; then
      echo "Unexpected RDS2 subnet CIDR: $rds2_cidr"
      exit 1
    fi

  elif [ "$IAAS" = "GCP" ]; then
    public_cidr="$(gcloud compute networks subnets describe "concourse-up-${deployment}-${region}-public" --region "$region" --format json | jq -r ".ipCidrRange")"
    if [ "$public_cidr" != "$expected_public_cidr" ]; then
      echo "Unexpected public subnet CIDR: $public_cidr"
      exit 1
    fi

    private_cidr="$(gcloud compute networks subnets describe "concourse-up-${deployment}-${region}-private" --region "$region" --format json | jq -r ".ipCidrRange")"
    if [ "$private_cidr" != "$expected_private_cidr" ]; then
      echo "Unexpected private subnet CIDR: $private_cidr"
      exit 1
    fi
  else
    echo "Unknown iaas: $IAAS"
    exit 1
  fi

  echo "Network CIDR ranges correct"
}
