#!/bin/bash

# If we're testing GCP, we need credentials to be available as a file
function setGoogleCreds() {
    echo "${GOOGLE_APPLICATION_CREDENTIALS_CONTENTS}" > googlecreds.json
    export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
}
