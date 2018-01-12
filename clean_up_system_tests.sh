#!/bin/bash

aws s3 ls | awk -F- '/concourse-up-system-test/{print "echo concourse-up --non-interactive destroy --region "$8"-"$9"-"$10"  "$5"-"$6"-"$7}' | sort -u | xargs -I {} bash -c '{}'