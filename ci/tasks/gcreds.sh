#!/bin/bash
function setGoogleCreds() {
    echo "${GOOGLE_APPLICATION_CREDENTIALS_CONTENTS}" > googlecreds.json
    export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
}