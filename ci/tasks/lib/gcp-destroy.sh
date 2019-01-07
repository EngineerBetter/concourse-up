#!/bin/bash

function recordDeployedState() {
    echo "Record ids of Concourse componenents"
    vpc_id=$(gcloud compute networks list --filter="name ~ concourse-up-$deployment" --format=json | jq -r '.[0].name')
    # both instances and volume_ids need finalising - 
    instances=$(gcloud compute instances list --filter="networkInterfaces:$vpc_id" --format=json)
    volume_ids=$(echo "$instances" | jq -r )

    echo "Get terraform state from bucket"
    config_bucket="concourse-up-$deployment-europe-west1-config"
    gsutil cp "gs://$config_bucket/terraform.tfstate" .

    echo "Record name of db instance"
    
}

function assertEverythingDeleted() {

}