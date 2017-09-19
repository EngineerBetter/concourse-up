#!/bin/bash

set -eu

cp -r concourse-up/. concourse-up-new
cp compilation-vars/compilation-vars.json concourse-up-new/compilation-vars.json

pushd concourse-up-new
  if [[ $(git status --porcelain) ]]; then
    git add compilation-vars.json
    git config --global user.email "systems@engineerbetter.com"
    git config --global user.name "CI"
    git commit -m "add compilation-vars.json latest version"
  fi
popd