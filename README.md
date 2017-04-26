# Concourse-Up

A tool for easily deploying [Concourse](concourse.ci) in a single command.

## Why?

Concourse is easy to get started with, but as soon as you want your team to use it you've
previously had to learn BOSH. The goal of `concourse-up` is to hide the complexity of
BOSH, providing you with a single command for getting your Concourse up and keeping it running.

## Features

- Deploys you a Concourse CI on AWS, without you having to know anything about BOSH
- Idempotent deployment (uses RDS for BOSH and Concourse databases)
- Supports https access by default

## Prerequisites

- AWS credentials, either configured via `aws configure`, or as [env vars](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment)
- [Terraform](https://www.terraform.io/intro/getting-started/install.html) 0.9.3 or newer

## Install

`go get github.com/engineerbetter/concourse-up`

## Usage

To deploy a new Concourse:

`concourse-up deploy <your-project-name>`

To destroy a Concourse:

`concourse-up destroy <your-project-name>`

That's it!

## Tests

`ginkgo -r`

## CI

Set the pipeline with:

```sh
fly -t sombrero set-pipeline -p concourse-up -c ci/pipeline.yml --var private_key="$(cat path/to/key)" -l secret_credentials.yml
```
