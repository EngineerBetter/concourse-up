#!/bin/bash

set -eu

cp -R concourse-up/* concourse-up-new
cp compilation-vars/compilation-vars.json concourse-up-new/compilation-vars.json

pushd
  concourse-up-new
  git add compilation-vars.json
popd