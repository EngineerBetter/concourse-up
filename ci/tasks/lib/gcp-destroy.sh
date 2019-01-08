#!/bin/bash

function recordDeployedState() {
    echo "Record ids of Concourse componenents"
    vpc_name=$(gcloud compute networks list --filter="name ~ concourse-up-$deployment" --format=json | jq -r '.[0].name') 
    instances=$(gcloud compute instances list --filter="networkInterfaces[0].network:$vpc_name" --format=json)
    
    existing_volumes=$(echo "$instances" | tr -d "[:cntrl:]" | jq -r '[.[].disks[].source]')
    echo "Get terraform state from bucket"
    config_bucket="concourse-up-$deployment-europe-west1-config"
    gsutil cp "gs://$config_bucket/default.tfstate" .

    echo "Record name of db instance"
    cloud_sql_instance_name=$(terraform output -state=default.tfstate -json | jq -r '.db_name.value')
    
}

function assertEverythingDeleted() {
    echo "Check that volumes have been deleted"
    all_volumes=$(gcloud compute instances list --format=json | jq '[.[].disks[].source]')
    volumes_count=$(echo "$existing_volumes $all_volumes" | jq --slurp '[.[0][] as $x | .[1][] | select($x == .)] | length')
    echo "Volumes still remaining after deletion: $volumes_count"
    [ "$volumes_count" -eq 0 ]

    instances=$(gcloud compute instances list --filter="networkInterfaces[0].network:$vpc_name" --format=json)
    instances_count=$(echo "$instances" |tr -d "[:cntrl:]" | jq '. | length')
    echo "Instances still remaining after deletion: $instances_count"
    [ "$instances_count" -eq 0 ]

    echo "Check that the CloudSQL instance has been deleted"
    cloud_sql_count=$(gcloud sql instances list --filter="name=$cloud_sql_instance_name" --format json | jq '. | length')
    [ "$cloud_sql_count" -eq 0 ]


}