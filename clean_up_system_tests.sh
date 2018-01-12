#!/bin/bash

aws s3 ls | awk -F- '/concourse-up-system-test/{print "--region "$8"-"$9"-"$10" "$5"-"$6"-"$7}' | sort -u | xargs -I {} concourse-up --non-interactive destroy {}