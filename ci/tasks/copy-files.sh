#!/bin/bash

set -eu

cp -r concourse-up/. concourse-up-new
mv director-versions/director-versions.json concourse-up-new/resources/director-versions.json
mv concourse-up-manifest/manifest.yml concourse-up-new/resources/manifest.yml
mv concourse-up-manifest/ops/versions.json concourse-up-new/resources/versions.json

pushd concourse-up-new
  if [[ $(git status --porcelain) ]]; then
    git add resources
    git config --global user.email "systems@engineerbetter.com"
    git config --global user.name "CI"
    git commit -m "add resources latest version"
  fi
popd
