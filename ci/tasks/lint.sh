#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up-ops/* "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up" || exit 1

go get -u github.com/mattn/go-bindata/...
grep -lr --include=*.go --exclude-dir=vendor "go:generate go-bindata" . | xargs -I {} go generate {}
gometalinter \
--disable-all \
--enable=goconst \
--enable=ineffassign \
--enable=vetshadow \
--enable=deadcode \
--exclude=bindata \
--exclude=resource/internal/file \
--vendor \
--enable-gc \
--deadline=120s \
./...

# Globally ignoring SC2154 as it doesn't play nice with variables
# set by Concourse for use in tasks.
# https://github.com/koalaman/shellcheck/wiki/SC2154
find . -name vendor -prune ! -type d -o -name '*.sh' -exec shellcheck -e SC2154 {} +
