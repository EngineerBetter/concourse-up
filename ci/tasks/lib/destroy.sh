#!/bin/bash

function recordDeployedState() {
    echo "Record ids of Concourse components"
    vpc_id=$(aws ec2 describe-vpcs --filter "Name=tag:Name,Values=concourse-up-$deployment" --region us-east-1 | jq -r '.Vpcs[0].VpcId')
    instances=$(aws ec2 describe-instances --filter "Name=vpc-id,Values=$vpc_id" --region us-east-1)
    volume_ids=$(echo "$instances" | jq -r '.Reservations[].Instances[].BlockDeviceMappings[].Ebs.VolumeId' | tr '\n' ',' | sed 's/,$//')


    echo "Get terraform state out of S3"
    config_bucket="concourse-up-$deployment-us-east-1-config"
    aws s3 cp s3://$config_bucket/terraform.tfstate .

    echo "Record name of RDS instance"
    rds_instance_name=$(terraform output -json | jq -r '.bosh_db_address.value' | awk -F. '{print $1}')
}

function assertEverythingDeleted() {
    echo "Check that EBS volumes have been deleted"
    volumes=$(aws ec2 describe-volumes --filters "Name=volume-id,Values=$volume_ids" --region us-east-1 | jq '.Volumes')
    volumes_count=$(echo "$volumes" | jq '. | length')
    echo "Volumes still remaining after deletion: $volumes_count"
    [ "$volumes_count" -eq 0 ]

    echo "Check that EC2 instances have been deleted"
    instances=$(aws ec2 describe-instances --filter "Name=vpc-id,Values=$vpc_id" --region us-east-1)
    instances_count=$(echo "$instances" | jq '.Reservations | length')
    echo "Instances still remaining after deletion: $instances_count"
    [ "$instances_count" -eq 0 ]

    echo "Check that the RDS instance has been deleted"
    set +e
    aws rds describe-db-instances --region us-east-1 --db-instance-identifier "$rds_instance_name"
    RdsExitCode=$?
    set -e

    echo "RDS instance check for $rds_instance_name returned exit code of $RdsExitCode (expecting non-zero)"
    [ "$RdsExitCode" -ne 0 ]
}
