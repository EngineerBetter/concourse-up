#!/bin/bash

set -eu

dirname=$(dirname $0)
. $dirname/../script_setup.sh

go install

name=$RANDOM
config_file=".concourse-up-${name}"

echo "Checking help command"
concourse-up --help | grep -q "Concourse-Up - A CLI tool to deploy Concourse CI" || (echo "Failed to find description in help output" && exit 1)
concourse-up --help | grep -q "deploy, d  Deploys or updates a Concourse" || (echo "Failed to find deploy instructions in help output" && exit 1)

echo "Checking deploy command"
concourse-up deploy --help | grep -q "Deploys or updates a Concourse" || (echo "Failed to display help output" && exit 1)
concourse-up deploy | grep -q 'Usage is `concourse-up deploy <name>`' || (echo "Failed to show usage when name isnt' provided'" && exit 1)
rm -f ~/$config_file
! aws ec2 describe-key-pairs --key-name "${name}-bosh" || (echo "keypair ${name}-bosh found unexpectedly in AWS" && exit 1)
concourse-up deploy $name || (echo "concourse-up deploy failed" && exit 1)
# ls -lart ~/ | grep -q $name || (echo "$config_file not created" && exit 1)
pushd ~/
  ls -la "${config_file}" || (echo "$config_file not created" && exit 1)
  grep -q private_key "${config_file}" || (echo "private_key not written to config file" && exit 1)
  config_md5=$(md5sum $config_file | awk {'print $1'})
popd
aws ec2 describe-key-pairs --key-name "${name}-bosh" || (echo "keypair ${name}-bosh not found in AWS" && exit 1)
concourse-up deploy $name || (echo "second concourse-up deploy failed" && exit 1)
pushd ~/
  config2_md5=$(md5sum $config_file | awk {'print $1'})
popd
diff  <(echo "$config_md5" ) <(echo "$config2_md5") &>/dev/null || (echo "local key was changed" && exit 1)

# TEMP delete until destroy command is implemented
aws ec2 delete-key-pair --key-name "${name}-bosh"
rm -f ~/$config_file