# Concourse-Up

A tool for easily deploying [Concourse](concourse.ci) in a single command.

![](https://ci.engineerbetter.com/api/v1/teams/main/pipelines/concourse-up/jobs/test/badge)

## Why?

Concourse is easy to get started with, but as soon as you want your team to use it you've
previously had to learn BOSH. Teams who just want great CI shouldn't need to think about this.
The goal of `concourse-up` is to hide the complexity of BOSH, while giving you all the benefits,
providing you with a single command for getting your Concourse up and keeping it running.

## Features

- Deploys you a Concourse CI on AWS, without you having to know anything about BOSH
- Idempotent deployment (uses RDS for BOSH and Concourse databases)
- Supports https access by default
- Uses cost effective AWS spot instances where possible (BOSH will take care of the service)
- Easy destroy and cleanup

## Prerequisites

- [Go](https://golang.org/doc/install) 1.6+
- An authenticated AWS environment. This can be done by doing one of:
  - Installing the [AWS CLI](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) and running `aws configure`
  - Exporting the following [environment variables](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment) before running `concourse-up`
    - `AWS_ACCESS_KEY_ID`
    - `AWS_SECRET_ACCESS_KEY`
    - `AWS_DEFAULT_REGION`
- [Terraform](https://www.terraform.io/intro/getting-started/install.html) 0.9.3 or newer

## Install

Run the following command to install Concourse-Up in your $PATH:

```
$ go get github.com/engineerbetter/concourse-up
```

## Usage

Deploy a new concourse with:

```
$ concourse-up deploy <your-project-name>
```

eg:

```
$ concourse-up deploy ci

...

DEPLOY SUCCESSFUL. Log in with:

fly --target ci login --concourse-url http://ci-concourse-up-1420669447.eu-west-1.elb.amazonaws.com --username admin --password abc123def456

```

To fetch information about your concourse-up deployment:

```
$ concourse-up info <your-project-name>
```

To destroy a Concourse:

```
$ concourse-up destroy <your-project-name>
```

That's it!

## What it does

Concourse up first creates an S3 bucket to store its own configuration and saves a `config.json` file there.

It then uses Terraform to deploy the following infrastructure:

- A VPC, with subnets and routing
- A load balancer
- An S3 bucket to use a BOSH blobstore
- An IAM user that can access the blobstore
- An IAM user that can deploy EC2 instances and update loadbalancers
- An AWS keypair for BOSH to use
- A security group to allow access to the BOSH director from your local IP
- A security group to allow access to Concourse from the internet
- An RDS instance (default: db.t2.small) for BOSH and Concourse to use

Once the terraform step is complete, concourse-up deploys a BOSH director on an t2.medium instance, and then uses that to deploy a concourse with the following settings:

- One m3.medium [spot](https://aws.amazon.com/ec2/spot/) instance for the Concourse web server
- One m3.xlarge spot instance used as a Concourse worker
- Access via a loadbalancer over HTTP and HTTPS using a self-signed cert created by concourse-up

## Tests

Tests use the [Ginkgo](https://onsi.github.io/ginkgo/) Go testing framework. The tests require that have set up AWS authentication locally.

Install ginkgo and run the tests with:

```
$ go get github.com/onsi/ginkgo/ginkgo
$ ginkgo -r
```

## Project

[Pivotal Tracker](https://www.pivotaltracker.com/n/projects/2011803)
