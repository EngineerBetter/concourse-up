#!/bin/bash

dirname=$(dirname $0)
. $dirname/../script_setup.sh

aws ec2 delete-key-pair --key-name "${name}-bosh"
