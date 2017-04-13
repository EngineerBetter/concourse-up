#!/bin/bash

set -eu

dirname=$(dirname $0)
. $dirname/../script_setup.sh

go install

name=cup-test
config_dir=".concourse-up"
key_file="concourse-up-${name}.pem"
key_file_path="${config_dir}/${key_file}"

echo "Checking help command"
concourse-up --help | grep -q "Concourse-Up - A CLI tool to deploy Concourse CI" || (echo "Failed to find description in help output" && exit 1)
concourse-up --help | grep -q "deploy, d  Deploys or updates a Concourse" || (echo "Failed to find deploy instructions in help output" && exit 1)

echo "Checking deploy command"
concourse-up deploy --help | grep -q "Deploys or updates a Concourse" || (echo "Failed to display help output" && exit 1)
concourse-up deploy | grep -q 'Usage is `concourse-up deploy <name>`' || (echo "Failed to show usage when name isnt' provided'" && exit 1)
rm -rf ~/$config_dir
! aws ec2 describe-key-pairs --key-name "${name}-bosh" || aws ec2 delete-key-pair --key-name "${name}-bosh"
concourse-up deploy $name || (echo "concourse-up deploy failed" && exit 1)
pushd ~/
  [[ -d $config_dir ]] || (echo "$config_dir not created" && exit 1)
  ls -la $key_file_path || (echo "$key_file not created" && exit 1)
  grep -q "BEGIN RSA PRIVATE KEY" $key_file_path || (echo "private_key not written to config file" && exit 1)
  config_md5=$(md5sum $key_file_path | awk {'print $1'})
popd
aws ec2 describe-key-pairs --key-name "${name}-bosh" || (echo "keypair ${name}-bosh not found in AWS" && exit 1)
concourse-up deploy $name || (echo "second concourse-up deploy failed" && exit 1)
pushd ~/
  config2_md5=$(md5sum $key_file_path | awk {'print $1'})
popd
diff  <(echo "$config_md5" ) <(echo "$config2_md5") &>/dev/null || (echo "local key was changed" && exit 1)

# TEMP delete until destroy command is implemented
aws ec2 delete-key-pair --key-name "${name}-bosh"
rm -rf ~/$config_dir