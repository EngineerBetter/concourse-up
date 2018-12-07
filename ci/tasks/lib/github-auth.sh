#!/bin/bash

# Add flags to an array that should have been initialised previously
function addGitHubFlagsToArgs() {
  args+=(--github-auth-client-id "$GITHUB_AUTH_CLIENT_ID")
  args+=(--github-auth-client-secret "$GITHUB_AUTH_CLIENT_SECRET")
  args+=(--domain cup.engineerbetter.com)
  args+=(--tls-cert "$EB_WILDCARD_CERT")
  args+=(--tls-key "$EB_WILDCARD_KEY")
  args+=(--region us-east-1)
}

function assertGitHubAuthConfigured() {
  config=$(./cup info --region us-east-1 --json $deployment)
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')

  fly --target system-test login \
    --concourse-url "https://$domain" \
    --username "$username" \
    --password "$password"

  echo "Check for github credentials in self-update pipeline"
  fly --target system-test get-pipeline --pipeline=concourse-up-self-update > pipeline

  grep -q "$GITHUB_AUTH_CLIENT_ID" pipeline
  grep -q "$GITHUB_AUTH_CLIENT_SECRET" pipeline

  echo "Check that github auth is enabled"
  fly --target system-test set-team \
    --team-name=git-team \
    --github-user=EngineerBetterCI \
    --non-interactive

  ( ( fly --target system-test login --team-name=git-team 2>&1 ) >fly_out ) &

  sleep 5

  pkill -9 fly

  url=$(grep redirect fly_out | sed 's/ //g')

  curl -sL "$url" | grep -q '/sky/issuer/auth/github'

  echo "GitHub Auth test passed"
}
