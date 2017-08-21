#!/bin/bash

set -eu

cp -r concourse-up/. concourse-up-new
cp compilation-vars/compilation-vars.json concourse-up-new/compilation-vars.json
version=$(cat version/version)

pushd concourse-up-new
  git add compilation-vars.json
  git config --global user.email "systems@engineerbetter.com"
  git config --global user.name "CI"
  git commit -m "add compilation-vars.json for version $version"
popd