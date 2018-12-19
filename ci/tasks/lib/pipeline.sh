#!/bin/bash

#to be used when setting a new pipeline
function assertPipelineIsSettableAndRunnable() {
  cert=$1
  domain=$2
  username=$3
  password=$4
  manifest=$5
  job=$6

  flyLogin
  setPipeline
  triggerJob

}

#to be used when checking that existing pipeline works
function assertPipelineIsRunnable() {
  cert=$1
  domain=$2
  username=$3
  password=$4
  job=$5

  flyLogin
  triggerJob
}


function setPipeline() {
  fly --target system-test set-pipeline \
  --non-interactive \
  --pipeline hello \
  --config "$manifest"

fly --target system-test unpause-pipeline \
  --pipeline hello
}
function triggerJob() {
  fly --target system-test trigger-job \
    --job hello/"$job" \
    --watch
}
function flyLogin() {
  fly --target system-test login \
    --ca-cert "$cert" \
    --concourse-url https://"$domain" \
    --username "$username" \
    --password "$password"

curl -k https://"$domain":3000

fly target system-test sync

}