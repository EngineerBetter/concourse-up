# Concourse-Up

![](http://i.imgur.com/gZPuUW5.png)

A tool for easily deploying [Concourse](https://concourse.ci) in a single command.

![](https://ci.engineerbetter.com/api/v1/teams/main/pipelines/concourse-up/jobs/test/badge)

## TL;DR

```
$ AWS_ACCESS_KEY_ID=<access-key-id> \
  AWS_SECRET_ACCESS_KEY=<secret-access-key> \
  AWS_DEFAULT_REGION=<default-region> \
  concourse-up deploy <your-project-name>
```

## Why?

Concourse is easy to get started with, but as soon as you want your team to use it you've
previously had to learn BOSH. Teams who just want great CI shouldn't need to think about this.
The goal of `concourse-up` is to hide the complexity of BOSH, while giving you all the benefits,
providing you with a single command for getting your Concourse up and keeping it running.

## Features

- Deploys Concourse 2.7.3 CI on AWS, without you having to know anything about BOSH
- Idempotent deployment (uses RDS for BOSH and Concourse databases)
- Supports https access by default using a user-provided certificate or auto-generating a self-signed one
- Uses cost effective AWS spot instances where possible (BOSH will take care of the service)
- Easy destroy and cleanup

## Prerequisites

- An authenticated AWS environment. This can be done by doing one of:
  - Installing the [AWS CLI](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) and running `aws configure`
  - Exporting the following [environment variables](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment) before running `concourse-up`
    - `AWS_ACCESS_KEY_ID`
    - `AWS_SECRET_ACCESS_KEY`
    - `AWS_DEFAULT_REGION`
- [Terraform](https://www.terraform.io/intro/getting-started/install.html) 0.9.3 or newer

## Install

Download the [latest release](https://github.com/EngineerBetter/concourse-up/releases) and install it into your $PATH:

## Usage

Deploy a new Concourse with:

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

A new deploy from scratch takes approximately 40 minutes.

To fetch information about your `concourse-up` deployment:

```
$ concourse-up info <your-project-name>
```

To destroy a Concourse:

```
$ concourse-up destroy <your-project-name>
```

That's it!

### Worker Configuration

By default `concourse-up` deploys a single worker instance of the `m3.xlarge` type. To increase the number of workers pass in the `--workers` flag eg:

```
$ concourse-up deploy chimichanga --workers 3
```

You can also change the size of each worker instance using the `--worker-size` flag. eg:

```
$ concourse-up deploy chimichanga --worker-size xlarge
```

The following table shows the allowed worker sizes and the corresponding AWS instance types

| --worker-size | AWS Instance type |
|---------------|-------------------|
| medium        | m3.medium         |
| large         | m3.large          |
| xlarge        | m3.xlarge         |


### Custom Domains

You can use a custom domain using the `--domain` flag eg:

```
$ concourse-up deploy chimichanga --domain chimichanga.engineerbetter.com
```

In the example above `concourse-up` will search for a Route 53 hosted zone that matches `chimichanga.engineerbetter.com` or `engineerbetter.com` and add a record to the longest match (`chimichanga.engineerbetter.com` in this example).

By default `concourse-up` will generate a self-signed cert using the given domain. If you'd like to provide your own certificate instead, pass the cert and private key as strings using the `--tls-cert` and `--tls-key` flags respectively. eg:

```
$ concourse-up deploy chimichanga \
  --domain chimichanga.engineerbetter.com \
  --tls-cert "$(cat chimichanga.engineerbetter.com.crt)" \
  --tls-key "$(cat chimichanga.engineerbetter.com.key)"
```

## Estimated Cost

By default, `concourse-up` deploys to the AWS eu-west-1 (Ireland) region, and uses spot instances for the Concourse VMs. The estimated monthly cost is as follows:

| Component     | Size             | Count | Price (USD) |
|---------------|------------------|-------|------------:|
| BOSH director | t2.medium        |     1 |       36.50 |
| Web Server    | m3.medium (spot) |     1 |       10.80 |
| Worker        | m3.xlarge (spot) |     1 |       48.47 |
| RDS instance  | db.t2.small      |     1 |       28.47 |
| Load balancer |         -        |     1 |       20.44 |
| **Total**         |                  |       |      **144.68** |

## What it does

`concourse-up` first creates an S3 bucket to store its own configuration and saves a `config.json` file there.

It then uses Terraform to deploy the following infrastructure:

- A VPC, with subnets and routing
- A load balancer
- An S3 bucket which BOSH uses as a blobstore
- An IAM user that can access the blobstore
- An IAM user that can deploy EC2 instances and update load balancers
- An AWS keypair for BOSH to use when deploying VMs
- An RDS instance (default: db.t2.small) for BOSH and Concourse to use
- A security group to allow access to the BOSH director from your local IP
- A security group for BOSH-deployed VMs
- A security group to allow access to the Concourse web server from the internet
- A security group to allow access to the RDS database from BOSH and it's VMs


Once the terraform step is complete, `concourse-up` deploys a BOSH director on an t2.medium instance, and then uses that to deploy a Concourse with the following settings:

- One m3.medium [spot](https://aws.amazon.com/ec2/spot/) for the Concourse web server
- One m3.xlarge spot instance used as a Concourse worker
- Access via a load balancer over HTTP and HTTPS using a user-provided certificate, or an auto-generated self-signed certificate if one isn't provided.

## Using a dedicated AWS IAM account

If you'd like to run concourse-up with it's own IAM account, create a user with the following permissions:

![](http://i.imgur.com/Q0mOUjv.png)

## Tests

Tests use the [Ginkgo](https://onsi.github.io/ginkgo/) Go testing framework. The tests require you to have set up AWS authentication locally.

Install ginkgo and run the tests with:

```
$ go get github.com/onsi/ginkgo/ginkgo
$ ginkgo -r
```

## Project

[Pivotal Tracker](https://www.pivotaltracker.com/n/projects/2011803)

[CI Pipeline](https://ci.engineerbetter.com/teams/main/pipelines/concourse-up)
