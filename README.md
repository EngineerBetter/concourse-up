# Concourse-Up

A tool for easily setting up a Concourse deployment in a single command.

# Features

- Deploys you a Concourse CI using BOSH, without you having to know anything about BOSH
- Idempotent deployment, using RDS for BOSH and Concourse databases

# Prerequisites

- AWS credentials, either configured via AWS CLI, or as envars
- Terraform 0.9.3 or newer

## Install

`go get` doesn't play nicely with private repos but there is an easy work-around:

```sh
mkdir $GOPATH/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/bitbucket.org/engineerbetter
git clone git@bitbucket.org:engineerbetter/concourse-up.git
go get -u bitbucket.org/engineerbetter/concourse-up
```

# How to Use it

`concourse-up deploy <your-project-name>`

## Tests

`ginkgo -r`

## CI

Set the pipeline with:

```sh
fly -t sombrero set-pipeline -p concourse-up -c ci/pipeline.yml --var private_key="$(cat path/to/key)"
```
