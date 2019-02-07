#!/bin/bash

if [ "$IAAS" = "AWS" ]; then
    [[ -n "$AWS_ACCESS_KEY_ID" ]]
    [[ -n "$AWS_SECRET_ACCESS_KEY" ]]
    # shellcheck disable=SC2034
    region=eu-west-1
elif [ "$IAAS" = "GCP" ]; then
    [[ -n "$GOOGLE_APPLICATION_CREDENTIALS_CONTENTS" ]]
    # shellcheck disable=SC1091
    source concourse-up/ci/tasks/lib/gcreds.sh
    setGoogleCreds
    # shellcheck disable=SC2034
    region=europe-west1
fi
