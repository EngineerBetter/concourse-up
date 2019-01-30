#!/bin/bash
#shellcheck disable=SC2016

set -eu

docker run -it \
    -e GO111MODULE=off \
    -v "${PWD}:/mnt/concourse-up" \
    -v "${PWD}/../concourse-up-ops:/mnt/concourse-up-ops" \
    engineerbetter/pcf-ops \
    bash -c \
    'cp -r /mnt/concourse-up* .; ./concourse-up/ci/tasks/lint.sh && cd ${GOPATH}/src/github.com/EngineerBetter/concourse-up && ginkgo -r'
