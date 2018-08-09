#!/bin/bash

set -e
[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; export BOSH_LOG_PATH=bosh.log; }

cp binary-linux/concourse-up-linux-amd64 ./cup
chmod +x ./cup

aws s3 ls \
| awk -F- '/concourse-up-systest/{print "yes yes | ./cup destroy --region "$7"-"$8"-"$9" "$5"-"$6}' \
| sort -u \
| xargs -P 8 -I {} bash -c '{}'
