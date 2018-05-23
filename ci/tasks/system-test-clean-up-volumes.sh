#!/bin/bash

set -e
[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; export BOSH_LOG_PATH=bosh.log; }
if [ -z "$SYSTEM_TEST_ID" ]; then
    SYSTEM_TEST_ID=$RANDOM
fi
deployment="system-test-$SYSTEM_TEST_ID"
set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

./cup deploy $deployment

sleep 60

vpc_id=$(aws ec2 describe-vpcs --filter "Name=tag:Name","Values=concourse-up-$deployment" --region eu-west-1 | jq -r '.Vpcs[0].VpcId')
volume_ids=$(aws ec2 describe-instances --filter "Name=vpc-id","Values=$vpc_id" --region eu-west-1 | jq -r '.Reservations[].Instances[].BlockDeviceMappings[].Ebs.VolumeId' | tr '\n' ',' | sed 's/,$//')

./cup --non-interactive destroy $deployment

sleep 180

volumes=$(aws ec2 describe-volumes --filters "Name=volume-id","Values=$volume_ids" --region eu-west-1 | jq '.Volumes')
volumes_count=$(echo $volumes | jq '. | length')

echo "Volumes still remaining after deletion: $volumes_count"

[ $volumes_count -eq 0 ]
