#!/bin/bash

set -eu

dirname=$(dirname $0)
. $dirname/../script_setup.sh

go install

echo "Checking help command"
concourse-up --help | grep -q "Concourse-Up - A CLI tool to deploy Concourse CI" || (echo "Failed to find description in help output" && exit 1)
