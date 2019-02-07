#!/bin/bash

function assertNetworkCidrsCorrect() {
  echo "About to check network CIDR ranges"

  if [ "$IAAS" = "AWS" ]; then
    vpc_cidr="$(aws --region eu-west-1 ec2 describe-vpcs --filters Name=tag:concourse-up-project,Values="$deployment" | jq -r ".Vpcs[0].CidrBlock")"
    if [ "$vpc_cidr" != "10.0.0.0/16" ]; then
      echo "Unexpected VPC CIDR: $vpc_cidr"
      exit 1
    fi

    public_cidr="$(aws --region eu-west-1 ec2 describe-subnets --filters Name=tag:Name,Values=concourse-up-"$deployment"-public | jq -r ".Subnets[0].CidrBlock")"
    if [ "$public_cidr" != "10.0.0.0/24" ]; then
      echo "Unexpected public subnet CIDR: $public_cidr"
      exit 1
    fi

    private_cidr="$(aws --region eu-west-1 ec2 describe-subnets --filters Name=tag:Name,Values=concourse-up-"$deployment"-private | jq -r ".Subnets[0].CidrBlock")"
    if [ "$private_cidr" != "10.0.1.0/24" ]; then
      echo "Unexpected private subnet CIDR: $private_cidr"
      exit 1
    fi

  elif [ "$IAAS" = "GCP" ]; then
    public_cidr="$(gcloud compute networks subnets describe concourse-up-"$deployment"-europe-west1-public --region europe-west1 --format json | jq -r ".ipCidrRange")"
    if [ "$public_cidr" != "10.0.0.0/24" ]; then
      echo "Unexpected public subnet CIDR: $public_cidr"
      exit 1
    fi

    private_cidr="$(gcloud compute networks subnets describe concourse-up-"$deployment"-europe-west1-private --region europe-west1 --format json | jq -r ".ipCidrRange")"
    if [ "$private_cidr" != "10.0.1.0/24" ]; then
      echo "Unexpected private subnet CIDR: $private_cidr"
      exit 1
    fi
  else
    echo "Unknown iaas: $IAAS"
    exit 1
  fi

  echo "Network CIDR ranges correct"
}
