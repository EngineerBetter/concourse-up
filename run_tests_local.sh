#!/bin/bash

set -eu

docker run -it -v "${PWD}:/mnt/concourse-up" -v "${PWD}/../concourse-up-ops:/mnt/concourse-up-ops" engineerbetter/pcf-ops bash -c "cp -r /mnt/concourse-up* .; ./concourse-up/ci/tasks/lint.sh"
