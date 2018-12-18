#!/bin/bash

#to be used when setting a new pipeline
function assertPipelineIsSettableAndRunnable() {
  cert=$1
  domain=$2
  username=$3
  password=$4
  manifest=$5
  job=$6

  login "$cert" "$domain" "$username" "$password"
  setPipeline "$manifest"
  triggerJob "$job"

}

#to be used when checking that existing pipeline works
function assertPipelineIsRunnable() {
  cert=$1
  domain=$2
  username=$3
  password=$4
  job=$5

  login "$cert" "$domain" "$username" "$password"
  triggerJob "$job"
}


function setPipeline() {
  fly --target system-test set-pipeline \
  --non-interactive \
  --pipeline hello \
  --config "$1"

fly --target system-test unpause-pipeline \
  --pipeline hello
}
function triggerJob() {
  fly --target system-test trigger-job \
    --job hello/"$job" \
    --watch
}
function login() {
  fly --target system-test login \
    --ca-cert "$1" \
    --concourse-url https://"$2" \
    --username "$3" \
    --password "$4"

curl -k https://"$2":3000

fly target system-test sync

}