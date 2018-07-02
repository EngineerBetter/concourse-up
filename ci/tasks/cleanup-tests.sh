#!/bin/bash

set -e
[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; export BOSH_LOG_PATH=bosh.log; }

cp release/concourse-up-linux-amd64 ./cup
chmod +x ./cup

aws s3 ls \
| awk -F- '/concourse-up-systest/{print "yes yes | ./cup destroy --region "$8"-"$9"-"$10" "$5"-"$6"-"$7}' \
| sort -u \
| xargs -P 8 -I {} bash -c '{}'
