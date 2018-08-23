# Concourse-Up

![](http://i.imgur.com/gZPuUW5.png)

A tool for easily deploying [Concourse](https://concourse-ci.org) in a single command.

![](https://ci.engineerbetter.com/api/v1/teams/main/pipelines/concourse-up/jobs/system-test/badge)

## TL;DR

```
$ AWS_ACCESS_KEY_ID=<access-key-id> \
  AWS_SECRET_ACCESS_KEY=<secret-access-key> \
  concourse-up deploy <your-project-name>
```

## Why?

Concourse is easy to get started with, but as soon as you want your team to use it you've
previously had to learn BOSH. Teams who just want great CI shouldn't need to think about this.
The goal of `concourse-up` is to hide the complexity of [BOSH](https://bosh.io), while giving you all the benefits,
providing you with a single command for getting your Concourse up and keeping it running. You can read more about the rationale for this tool in [this blog post](http://www.engineerbetter.com/2017/05/03/introducing-concourse-up.html). Some newer features, including self-update, are described in [this blog post](http://www.engineerbetter.com/2017/09/18/perpetual-motion-software-updates.html).

## Features

- Deploys the latest version of Concourse CI on AWS, without you having to know anything about BOSH
- Idempotent deployment with either manual upgrade or automatic self-upgrade
- Supports https access by default using a user-provided certificate or auto-generating a self-signed one
- Supports custom domains for your Concourse URL
- Uses cost effective AWS spot instances where possible (BOSH will take care of the service)
- Uses precompiled BOSH packages to minimise install time
- Horizontal and vertical worker scaling
- Vertical database scaling
- Workers reside behind a single, persistent public IP to simplify external security
- Easy destroy and cleanup
- Deploy to any AWS region
- Metrics infrastructure deployed by default (check http://your-concourse-url:3000)
- DB encryption turned on by default
- Uses credhub for secret management (see: <https://concourse-ci.org/creds.html>)

## Prerequisites

- Export the following [environment variables](http://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html) before running `concourse-up`:
    - `AWS_ACCESS_KEY_ID`
    - `AWS_SECRET_ACCESS_KEY`
 - Ensure you have the correct local dependencies for [bootstrapping a BOSH VM](https://bosh.io/docs/cli-v2-install/#additional-dependencies)

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
fly --target ci login --insecure --concourse-url https://52.18.43.185 --username admin --password abc123def456

Metrics available at https://52.18.43.185:3000 using the same username and password

Log into credhub with:
credhub login -u credhub-cli -p foobar987 -s https://52.18.43.185:8844/ --ca-cert "..."
```

A new deploy from scratch takes approximately 12 minutes.

To fetch information about your `concourse-up` deployment:

```
$ concourse-up info --json <your-project-name>
```

To destroy a Concourse:

```
$ concourse-up destroy <your-project-name>
```

That's it!

### Region Configuration

By default `concourse-up` deploys the BOSH director and Concourse VMs into `eu-west-1` region. To change the region, use the `--region` flag eg:

```
$ concourse-up deploy --region us-east-1 chimichanga
```

When deploying to a non-default region, you *must* pass the `--region` flag with all subsequent commands eg:

```
$ concourse-up info --region us-east-1 chimichanga
$ concourse-up destroy --region us-east-1 chimichanga
```

### Worker Configuration

By default `concourse-up` deploys a single worker instance of the `m4.xlarge` type. To increase the number of workers pass in the `--workers` flag eg:

```
$ concourse-up deploy --workers 3 chimichanga
```

You can also change the size of each worker instance using the `--worker-size` flag. eg:

```
$ concourse-up deploy --worker-size xlarge chimichanga
```

The following table shows the allowed worker sizes and the corresponding AWS instance types

| --worker-size | AWS Instance type |
|---------------|-------------------|
| medium        | t2.medium         |
| large         | m4.large          |
| xlarge        | m4.xlarge         |
| 2xlarge       | m4.2xlarge        |
| 4xlarge       | m4.4xlarge        |
| 10xlarge      | m4.10xlarge       |
| 16xlarge      | m4.16xlarge       |


### Custom Domains

You can use a custom domain using the `--domain` flag eg:

```
$ concourse-up deploy --domain chimichanga.engineerbetter.com chimichanga
```

In the example above `concourse-up` will search for a Route 53 hosted zone that matches `chimichanga.engineerbetter.com` or `engineerbetter.com` and add a record to the longest match (`chimichanga.engineerbetter.com` in this example).

By default `concourse-up` will generate a self-signed cert using the given domain. If you'd like to provide your own certificate instead, pass the cert and private key as strings using the `--tls-cert` and `--tls-key` flags respectively. eg:

```
$ concourse-up deploy \
  --domain chimichanga.engineerbetter.com \
  --tls-cert "$(cat chimichanga.engineerbetter.com.crt)" \
  --tls-key "$(cat chimichanga.engineerbetter.com.key)" \
  chimichanga
```

## RDS Size Configuration

You can change the size of the RDS instance shared by BOSH and the Concourse using the `--db-size` flag. eg:

```
$ concourse-up deploy --db-size medium chimichanga
```

Note that when changing the database size on an existing concourse-up deployment, the RDS instance will scaled by terraform resulting in approximately 3 minutes of downtime.

The following table shows the allowed database sizes and the corresponding AWS RDS instance types

| --db-size | AWS Instance type |
|-----------|-------------------|
| small     | db.t2.small       |
| medium    | db.t2.medium      |
| large     | db.m4.large       |
| xlarge    | db.m4.xlarge      |
| 2xlarge   | db.m4.2xlarge     |
| 4xlarge   | db.m4.4xlarge     |

## Self-update

When Concourse-up deploys Concourse, it now adds a pipeline to the new Concourse called `concourse-up-self-update`. This pipeline continuously monitors our Github repo for new releases and updates Concourse in place whenever a new version of Concourse-up comes out.

This pipeline is paused by default, so just unpause it in the UI to enable the feature.

## Upgrading manually

Patch releases of `concourse-up` are compiled, tested and released automatically whenever a new stemcell or component release appears on [bosh.io](https://bosh.io).

To upgrade your Concourse, grab the [latest release](https://github.com/EngineerBetter/concourse-up/releases/latest) and run `concourse-up deploy <your-project-name>` again.

## Metrics

Concourse-up now automatically deploys Influxdb, Riemann, and Grafana on the web node. You can access Grafana on port 3000 of your regular concourse URL using the same username and password as your Concourse admin user. We put in a default dashboard that tracks

- Build times
- CPU usage
- Containers
- Disk usage

## Credential Management

Concourse-up deploys the [credhub](https://github.com/cloudfoundry-incubator/credhub) service alongside Concourse and configures Concourse to use it. More detail on how credhub integrates with Concourse can be found [here](https://concourse-ci.org/creds.html). You can log into credhub by running `$ eval "$(concourse-up info --env --region $region $deployment)"`.

## Firewall

Concourse-up normally allows incoming traffic from any address to reach your web node. You can use the `--allow-ips` flag to add firewall rules to prevent this.
For example to deploy Concourse-up and only allow traffic from your local machine, you could use the command `concourse-up deploy --allow-ips $(dig +short myip.opendns.com @resolver1.opendns.com)`.
`--allow-ips` takes a comma seperated list of IP addresses or CIDR ranges.

## Estimated Cost

By default, `concourse-up` deploys to the AWS eu-west-1 (Ireland) region, and uses spot instances for large and xlarge Concourse VMs. The estimated monthly cost is as follows:

| Component     | Size             | Count | Price (USD) |
|---------------|------------------|-------|------------:|
| BOSH director | t2.small         |     1 |       18.25 |
| Web Server    | t2.small         |     1 |       18.25 |
| Worker        | m4.xlarge (spot) |     1 |       40.00 |
| RDS instance  | db.t2.small      |     1 |       28.47 |
| NAT Gateway   |         -        |     1 |       35.04 |
| gp2 storage   | 20GB (bosh, web) |     2 |        4.40 |
| gp2 storage   | 220GB (worker)   |     1 |       22.00 |
| **Total**     |                  |       |  **170.81** |

## What it does

`concourse-up` first creates an S3 bucket to store its own configuration and saves a `config.json` file there.

It then uses Terraform to deploy the following infrastructure:

- A VPC, with public and private subnets and routing
- A NAT gateway for outbound traffic from the private subnet
- An S3 bucket which BOSH uses as a blobstore
- An IAM user that can access the blobstore
- An IAM user that can deploy EC2 instances
- An AWS keypair for BOSH to use when deploying VMs
- An RDS instance (default: db.t2.small) for BOSH and Concourse to use
- Concourse database is [encrypted](http://concourse-ci.org/encryption.html) by default
- A security group to allow access to the BOSH director from your local IP
- A security group for BOSH-deployed VMs
- A security group to allow access to the Concourse web server from the internet
- A security group to allow access to the RDS database from BOSH and it's VMs


Once the terraform step is complete, `concourse-up` deploys a BOSH director on an t2.micro instance, and then uses that to deploy a Concourse with the following settings:

- One t2.small for the Concourse web server
- One m4.xlarge [spot](https://aws.amazon.com/ec2/spot/) instance used as a Concourse worker
- Access via over HTTP and HTTPS using a user-provided certificate, or an auto-generated self-signed certificate if one isn't provided.

## Using a dedicated AWS IAM account

If you'd like to run concourse-up with it's own IAM account, create a user with the following permissions:

![](http://i.imgur.com/Q0mOUjv.png)

## Project

[Pivotal Tracker](https://www.pivotaltracker.com/n/projects/2011803)

[CI Pipeline](https://ci.engineerbetter.com/teams/main/pipelines/concourse-up) (deployed with Concourse Up!)

## Development

### Pre-requisites

To build and test you'll need:

* Golang 1.10
* to have installed `github.com/a-urth/go-bindata`

### Tests

Tests use the [Ginkgo](https://onsi.github.io/ginkgo/) Go testing framework. The tests require you to have set up AWS authentication locally.

Install ginkgo and run the tests with:

```
$ go get github.com/onsi/ginkgo/ginkgo
$ ginkgo -r
```

### Building locally

`concourse-up` uses [golang compile-time variables](https://github.com/golang/go/wiki/GcToolchainTricks#including-build-information-in-the-executable) to set the release versions it uses. To build locally use the `build_local.sh` script, rather than running `go build`.

### Bumping Manifest/Ops File versions

The `get-versions` job in the pipeline listens for new patch or minor versions of `manifest.yml` and `ops/versions.json` coming from the `concourse-up-manifest` repo. In order to pick up a new major version first make sure it exists on the `release` branch in `concourse-up-manifest` then modify `tag_filter: X.*.*` in the `concourse-up-manifest` resource where `X` is the major version you want to pin to.

### Pipeline Resource Names

In the `concourse-up` pipeline there are three different resources listening to this repo.

* `concourse-up` is the whole repo. This is used to run all the jobs but doesn't trigger anything
* `concourse-up-code` listens for changes to `concourse-up` code and ignores changes to this README, `resources/manifest.yml`, `resources/versions.json`, `resources/director-versions.json`, and `resources/shas.json`
* `concourse-up-versions` is the complement of `concourse-up-code` (i.e. it just listens for ops file and manifest changes)
