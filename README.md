# Concourse-Up

![Smiling Pilot](http://i.imgur.com/uLWVhJA.jpg)

A tool for easily deploying [Concourse](concourse.ci) in a single command.

# Features

- Deploys you a Concourse CI using BOSH, without you having to know anything about BOSH
- Idempotent deployment, using RDS for BOSH and Concourse databases

# Prerequisites

- AWS credentials, either configured via AWS CLI, or as [env vars](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment)
- [Terraform](https://www.terraform.io/intro/getting-started/install.html) 0.9.3 or newer

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
fly -t sombrero set-pipeline -p concourse-up -c ci/pipeline.yml --var private_key="$(cat path/to/key)" -l secret_credentials.yml
```
