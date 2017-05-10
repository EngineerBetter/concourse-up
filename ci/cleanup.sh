#!/bin/bash

set -ux

# WIP script to clean up after a failed system test

deployment=concourse-up-$1

aws s3 rb s3://$deployment-config --force
aws s3 rb s3://$deployment-blobstore --force

aws iam delete-user-policy --user-name $deployment-blobstore --policy-name "$deployment-blobstore"
aws iam delete-user-policy --user-name $deployment-bosh --policy-name "$deployment-bosh"

blobstore_key=$(aws iam list-access-keys --user-name $deployment-blobstore | jq -r .AccessKeyMetadata[0].AccessKeyId)
aws iam delete-access-key --user-name $deployment-blobstore --access-key-id $blobstore_key

bosh_key=$(aws iam list-access-keys --user-name $deployment-bosh | jq -r .AccessKeyMetadata[0].AccessKeyId)
aws iam delete-access-key --user-name $deployment-bosh --access-key-id $bosh_key

aws iam delete-user --user-name $deployment-blobstore
aws iam delete-user --user-name $deployment-bosh

aws elb delete-load-balancer --load-balancer-name $deployment

