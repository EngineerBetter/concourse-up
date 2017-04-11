# Concourse-Up

A tool for easily setting up a Concourse deployment in a single command.

## Install

`go get` doesn't play nicely with private repos but there is an easy work-around:

```sh
mkdir $GOPATH/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/bitbucket.org/engineerbetter
git clone git@bitbucket.org:engineerbetter/concourse-up.git
go get -u bitbucket.org/engineerbetter/concourse-up
```

## Tests

`ginkgo -r`

## CI

Set the pipeline with:

```sh
fly -t sombrero set-pipeline -p concourse-up -c ci/pipeline.yml --var private_key="$(cat path/to/key)"
```