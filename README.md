# Deprecated - use [Control Tower](https://github.com/EngineerBetter/control-tower) instead

**Concourse-Up has been replaced with [Control Tower](https://github.com/EngineerBetter/control-tower). First-time users should deploy using `control-tower` and raise issues under that project.**

---

# Concourse-Up (Deprecated)

[![asciicast](https://asciinema.org/a/xVKD0dQuXdEmOcExt4A9WfbEN.svg)](https://asciinema.org/a/xVKD0dQuXdEmOcExt4A9WfbEN)

A tool for easily deploying [Concourse](https://concourse-ci.org) in a single command.

![](https://ci.engineerbetter.com/api/v1/teams/main/pipelines/concourse-up/jobs/system-test/badge)

## TL;DR

##### AWS

```sh
$ AWS_ACCESS_KEY_ID=<access-key-id> \
  AWS_SECRET_ACCESS_KEY=<secret-access-key> \
  concourse-up deploy <your-project-name>
```

##### GCP

```sh
$ GOOGLE_APPLICATION_CREDENTIALS=<path/to/googlecreds.json> \
  concourse-up deploy --iaas gcp <your-project-name> 
```

## Why Concourse-Up?

The goal of Concourse-Up is to be the world's easiest way to deploy and operate Concourse CI in production. 

In just one command you can deploy a new Concourse environment for your team, on either AWS or GCP. Your Concourse-Up deployment will *upgrade itself* and self-heal, restoring the underlying VMs if needed. Using the same command-line tool you can do things like manage DNS, scale your environment, or manage firewall policy. CredHub is provided for secrets management and Grafana for viewing your Concourse metrics.

You can keep up to date on Concourse-Up announcements by reading the [EngineerBetter Blog](http://www.engineerbetter.com/blog/)

## Feature Summary

- Deploys the latest version of Concourse CI on any region in AWS or GCP
- Manual upgrade or automatic self-upgrade
- Access your Concourse over https access by default, with auto-generated or self-provided cert.
- Deploy on your own domain, if you have a zone in Route53 or Cloud DNS.
- Scale your workers horizontally or vertically 
- Scale your Concourse database
- Presents workers on a single public IP to simplify external security policy
- Database encryption enabled by default
- Includes Grafana metrics dashboard (check http://your-concourse-url:3000)
- Includes CredHub for secret management (see: <https://concourse-ci.org/creds.html>)
- Saves you money by using AWS spot or GCP preemptible instances where possible, restarting them when needed
- Idempotent deployment and operations
- Easy destroy and cleanup

### Feature Table

| **Feature** | **AWS** | **GCP** |
|:------------|:-------:|:-------:|
| Concourse IP whitelisting | **+** | **+** |
| Credhub | **+** | **+** |
| Custom domains | **+** | **+** |
| Custom tagging | **BOSH only** | **BOSH only** |
| Custom TLS certificates | **+** | **+** |
| Database vertical scaling | **+** | **+** |
| GitHub authentication | **+** | **+** |
| Grafana | **+** | **+** |
| Interruptable worker support | **+** | **+** |
| Letsencrypt integration | **+** | **+** |
| Namespace support | **+** | **+** |
| Region selection | **+** | **+** |
| Retrieving deployment information | **+** | **+** |
| Retrieving deployment information as shell exports | **+** | **+** |
| Retrieving deployment information in JSON | **+** | **+** |
| Retrieving director NATS cert expiration | **+** | **+** |
| Rotating director NATS cert | **+** | **+** |
| Self-Update support | **+** | **+** |
| Teardown deployment | **+** | **+** |
| Web server vertical scaling | **+** | **+** |
| Worker horizontal scaling | **+** | **+** |
| Worker type selection | **+** | **N/A** |
| Worker vertical scaling | **+** | **+** |
| Zone selection | **+** | **+** |
| Customised networking | **+** | **+** |

## Prerequisites

- One of:
  - The environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are set.
  - Credentials for the default profile in `~/.aws/credentials` are present.
  - Credentials for a profile in `~/.aws/credentials` are present.
  - The environment variable `GOOGLE_APPLICATION_CREDENTIALS_CONTENTS` set to the path to a GCP credentials json file
- Ensure your credentials are *long lived credentials* and not *temporary security credentials*
- Ensure you have the correct local dependencies for [bootstrapping a BOSH VM](https://bosh.io/docs/cli-v2-install/#additional-dependencies)

## Install

Download the [latest release](https://github.com/EngineerBetter/concourse-up/releases) and install it into your `PATH`

## Usage

### Global flags

- `--region value`    AWS or GCP region (default: "eu-west-1" on AWS and "europe-west1" on GCP) [$AWS_REGION]
- `--namespace value` Any valid string that provides a meaningful namespace of the deployment - Used as part of the configuration bucket name [$NAMESPACE].
    >Note that if namespace has been provided in the initial `deploy` it will be required for any subsequent `concourse-up` calls against the same deployment.

#### Choosing an IAAS

The default IAAS for Concourse-Up is AWS. To choose a different IAAS use the `--iaas` flag. For every IAAS provider apart from AWS this flag is required for all commands.

Supported IAAS values: AWS, GCP

- `--iaas value` (optional) IAAS, can be AWS or GCP (default: "AWS") [$IAAS]

### Deploy

Deploy a new Concourse with:

```sh
concourse-up deploy <your-project-name>
```

eg:

```sh
$ concourse-up deploy ci

...
DEPLOY SUCCESSFUL. Log in with:
fly --target ci login --insecure --concourse-url https://10.0.0.0 --username  --password

Metrics available at https://10.0.0.0:3000 using the same username and password

Log into credhub with:
eval "$(concourse-up info ci --env)"
```

A new deploy from scratch takes approximately 20 minutes.

#### Flags

All flags are optional. Configuration settings provided via flags will persist in later deployments unless explicitly overriden.

- `--domain value`       Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com) [$DOMAIN]
    ```sh
    $ concourse-up deploy --domain chimichanga.engineerbetter.com chimichanga
    ```

    In the example above `concourse-up` will search for a hosted zone that matches `chimichanga.engineerbetter.com` or `engineerbetter.com` and add a record to the longest match (`chimichanga.engineerbetter.com` in this example).

- `--tls-cert value`     TLS cert to use with Concourse endpoint [$TLS_CERT]
- `--tls-key value`      TLS private key to use with Concourse endpoint [$TLS_KEY]

    By default `concourse-up` will generate a self-signed cert using the given domain. If you'd like to provide your own certificate instead, pass the cert and private key as strings using the `--tls-cert` and `--tls-key` flags respectively. eg:

    ```sh
    $ concourse-up deploy \
      --domain chimichanga.engineerbetter.com \
      --tls-cert "$(cat chimichanga.engineerbetter.com.crt)" \
      --tls-key "$(cat chimichanga.engineerbetter.com.key)" \
      chimichanga
    ```

- `--workers value`      Number of Concourse worker instances to deploy (default: 1) [$WORKERS]
- `--worker-type`        Specify a worker type for aws (m5 or m4) (default: "m4") [$WORKER_TYPE] (see comparison table below). **Note: this is an AWS-specific option**

> AWS does not offer m5 instances in all regions, and even for regions that do offer m5 instances, not all zones within that region may offer them. To complicate matters further, each AWS account is assigned AWS zones at random - for instance, `eu-west-1a` for one account may be the same as `eu-west-1b` in another account. If m5s are available in your chosen region but _not_ the zone Concourse-Up has chosen, create a new deployment, this time specifying another `--zone`.

- `--worker-size value`  Size of Concourse workers. Can be medium, large, xlarge, 2xlarge, 4xlarge, 10xlarge, 12xlarge, 16xlarge or 24xlarge depending on the worker-type (see above) (default: "xlarge") [$WORKER_SIZE]

    | --worker-size | AWS m4 Instance type | AWS m5 Instance type* | GCP Instance type |
    |---------------|----------------------|-----------------------|-------------------|
    | medium        | t2.medium            | t2.medium             | n1-standard-1     |
    | large         | m4.large             | m5.large              | n1-standard-2     |
    | xlarge        | m4.xlarge            | m5.xlarge             | n1-standard-4     |
    | 2xlarge       | m4.2xlarge           | m5.2xlarge            | n1-standard-8     |
    | 4xlarge       | m4.4xlarge           | m5.4xlarge            | n1-standard-16    |
    | 10xlarge      | m4.10xlarge          |                       | n1-standard-32    |
    | 12xlarge      |                      | m5.12xlarge           |                   |
    | 16xlarge      | m4.16xlarge          |                       | n1-standard-64    |
    | 24xlarge      |                      | m5.24xlarge           |                   |

    \* _m5 instances not available in all regions and all zones. See `--worker-type` for more info._

- `--web-size value`     Size of Concourse web node. Can be small, medium, large, xlarge, 2xlarge (default: "small") [$WEB_SIZE]

    | --web-size | AWS Instance type | GCP Instance type |
    |------------|-------------------|-------------------|
    | small      | t2.small          | n1-standard-1     |
    | medium     | t2.medium         | n1-standard-2     |
    | large      | t2.large          | n1-standard-4     |
    | xlarge     | t2.xlarge         | n1-standard-8     |
    | 2xlarge    | t2.2xlarge        | n1-standard-16    |

- `--db-size value`      Size of Concourse Postgres instance. Can be small, medium, large, xlarge, 2xlarge, or 4xlarge (default: "small") [$DB_SIZE]

    >Note that when changing the database size on an existing concourse-up deployment, the SQL instance will scaled by terraform resulting in approximately 3 minutes of downtime.

    The following table shows the allowed database sizes and the corresponding AWS RDS & CloudSQL instance types

    | --db-size | AWS Instance type | GCP Instance type  |
    |-----------|-------------------|--------------------|
    | small     | db.t2.small       | db-g1-small        |
    | medium    | db.t2.medium      | db-custom-2-4096   |
    | large     | db.m4.large       | db-custom-2-8192   |
    | xlarge    | db.m4.xlarge      | db-custom-4-16384  |
    | 2xlarge   | db.m4.2xlarge     | db-custom-8-32768  |
    | 4xlarge   | db.m4.4xlarge     | db-custom-16-65536 |

- `--allow-ips value`    Comma separated list of IP addresses or CIDR ranges to allow access to (default: "0.0.0.0/0") [$ALLOW_IPS]

    > Note: `allow-ips` governs what can access Concourse but not what can access the control plane (i.e. the BOSH director).

- `--github-auth-client-id value`      Client ID for a github OAuth application - Used for Github Auth [$GITHUB_AUTH_CLIENT_ID]
- `--github-auth-client-secret value`  Client Secret for a github OAuth application - Used for Github Auth [$GITHUB_AUTH_CLIENT_SECRET]
- `--add-tag key=value` Add a tag to the VMs that form your `concourse-up` deployment. Can be used multiple times in a single `deploy` command.
- `--spot=value` Use spot instances for workers. Can be true/false. Default is true.

    > Concourse Up uses spot instances for workers as a cost saving measure. Users requiring lower risk may switch this feature off by setting --spot=false.
- `--preemptible=value` Use preemptible instances for workers. Can be true/false. Default is true.

    > Be aware the [preemptible instances](https://cloud.google.com/preemptible-vms/) _will_ go down at least once every 24 hours so deployments with only one worker _will_ experience downtime with this feature enabled. BOSH will ressurect falled workers automatically.

    `spot` and `preemptible` are interchangeable so if either of them is set to false then interruptible instances will not be used regardless of your IaaS. i.e:

    ```sh
    # Results in an AWS deployment using non-spot workers
    concourse-up deploy --spot=true --preemptible=false <your-project-name>
    # Results in an AWS deployment using non-spot workers
    concourse-up deploy --preemptible=false <your-project-name>
    # Results in a GCP deployment using non-preemptible workers
    concourse-up deploy --iaas gcp --spot=false <your-project-name>
    ```

- `--zone`            Specify an availability zone [$ZONE] (cannot be changed after the initial deployment)

If any of the following 5 flags is set, all the required ones from this group need to be set
- `--vpc-network-range value`      Customise the VPC network CIDR to deploy into (required for AWS) [$VPC_NETWORK_RANGE]
- `--public-subnet-range value`    Customise public network CIDR (if IAAS is AWS must be within --vpc-network-range) (required) [$PUBLIC_SUBNET_RANGE]
- `--private-subnet-range value`   Customise private network CIDR (if IAAS is AWS must be within --vpc-network-range) (required) [$PRIVATE_SUBNET_RANGE]
- `--rds-subnet-range1 value`      Customise first rds network CIDR (must be within --vpc-network-range) (required for AWS) [$RDS_SUBNET_RANGE1]
- `--rds-subnet-range2 value`      Customise second rds network CIDR (must be within --vpc-network-range) (required for AWS) [$RDS_SUBNET_RANGE2]

    > All the ranges above should be in the CIDR format of IPv4/Mask. The sizes can vary as long as `vpc-network-range` is big enough to contain all others (in case IAAS is AWS). The smallest CIDR for `public` and `private` subnets is a /28. The smallest CIDR for `rds1` and `rds2` subnets is a /29

### Info

To fetch information about your `concourse-up` deployment:

```sh
$ concourse-up info --json <your-project-name>
```

To load credentials into your environment from your `concourse-up` deployment:

```sh
$ eval "$(concourse-up info --env <your-project-name>)"
```

To check the expiry of the BOSH Director's NATS CA certificate:

```sh
$ concourse-up info --cert-expiry <your-project-name>
```

**Warning: if your deployment is approaching a year old, it may stop working due to expired certificates. For information please see this issue https://github.com/EngineerBetter/concourse-up/issues/81.**

#### Flags

All flags are optional

`--json`          Output as json [$JSON]
`--env`           Output environment variables
`--cert-expiry`   Output the expiry of the BOSH director's NATS certificate

### Destroy

To destroy your Concourse:

```sh
$ concourse-up destroy <your-project-name>
```

### Maintain

Handles maintenance operations in concourse-up

#### Flags

All flags are optional

- `--renew-nats-cert` Rotate the NATS certificate on the director
    >Note that the NATS certificate [is hardcoded to expire after 1 year](https://github.com/cloudfoundry/bosh-cli/blob/master/vendor/github.com/cloudfoundry/config-server/types/certificate_generator.go#L171). This command follows [the istructions on bosh.io](https://bosh.io/docs/nats-ca-rotation/) to rotate this certificate. **This operation _will_ cause downtime on your Concourse** as it performs multiple full recreates.
- `--stage value` Specify a specific stage at which to start the NATS certificate renewal process. If not specified, the stage will be determined automatically. See the following table for details.

    | Stage | Description |
    |-------|-------------|
    | 0     | Adding new CA (create-env) |
    | 1     | Recreating VMs for the first time (recreate) |
    | 2     | Removing old CA (create-env) |
    | 3     | Recreating VMs for the second time (recreate) |
    | 4     | Cleaning up director-creds.yml |

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

By default, `concourse-up` deploys to the AWS eu-west-1 (Ireland) region or the GCP europe-west1 (Belgium) region, and uses spot instances for large and xlarge Concourse VMs. The estimated monthly cost is as follows:

### AWS

| Component     | Size             | Count | Price (USD) |
|---------------|------------------|-------|------------:|
| BOSH director | t2.small         |     1 |       18.30 |
| Web Server    | t2.small         |     1 |       18.30 |
| Worker        | m4.xlarge (spot) |     1 |      ~50.00 |
| RDS instance  | db.t2.small      |     1 |       28.47 |
| NAT Gateway   |         -        |     1 |       35.15 |
| gp2 storage   | 20GB (bosh, web) |     2 |        4.40 |
| gp2 storage   | 200GB (worker)   |     1 |       22.00 |
| **Total**     |                  |       |  **176.62** |

### GCP

| Component     | Size                              | Count | Price (USD) |
|---------------|-----------------------------------|-------|------------:|
| BOSH director | n1-standard-1                     |     1 |       26.73 |
| Web Server    | n1-standard-1                     |     1 |       26.73 |
| Worker        | n1-standard-4 (preemptible)       |     1 |       32.12 |
| DB instance   | db-g1-small                       |     1 |       27.25 |
| NAT Gateway   | n1-standard-1                     |     1 |       26.73 |
| disk storage  | 20GB (bosh, web) + 200GB (worker) |   -   |       40.80 |
| **Total**     |                                   |       |  **180.35** |

## What it does

`concourse-up` first creates an S3 or GCS bucket to store its own configuration and saves a `config.json` file there.

It then uses Terraform to deploy the following infrastructure:

- AWS
  - Key pair
  - S3 bucket for the blobstore
  - IAM user that can access the blobstore
    - IAM access key
    - IAM user policy
  - IAM user that can deploy EC2 instances
    - IAM access key
    - IAM user policy
  - VPC
  - Internet gateway
  - Route for internet_access
  - NAT gateway
  - Route table for private
  - Subnet for public
  - Subnet for private
  - Route table association for private
  - Route53 record for Concourse
  - EIP for director, ATC, and NAT
  - Security groups for director, vms, RDS, and ATC
  - Route table for RDS
  - Route table associations for RDS
  - Subnets for RDS
  - DB subnet group
  - DB instance
- GCP
  - A DNS A record pointing to the ATC IP
  - A Compute route for the nat instance
  - A Compute instance for the nat
  - A Compute network
  - Public and Private Compute subnetworks
  - Compute firewalls for director, nat, atc-one, atc-two, vms, atc-three, internal, and sql
  - A Service account for for bosh
  - A Service account key for bosh
  - A Project iam member for bosh
  - Compute addresses for the ATC and Director
  - A Sql database instance
  - A Sql database
  - A Sql user

Once the terraform step is complete, `concourse-up` deploys a BOSH director on an t2.small/n1-standard-1 instance, and then uses that to deploy a Concourse with the following settings:

- One t2.small/n1-standard-1 for the Concourse web server
- One m4.xlarge [spot](https://aws.amazon.com/ec2/spot/)/n1-standard-4 [preemptible](https://cloud.google.com/preemptible-vms/) instance used as a Concourse worker
- Access via over HTTP and HTTPS using a user-provided certificate, or an auto-generated self-signed certificate if one isn't provided.

## Using a dedicated AWS IAM account

If you'd like to run concourse-up with it's own IAM account, create a user with the following permissions:

![](http://i.imgur.com/Q0mOUjv.png)

## Using a dedicated GCP IAM member

A IAM Primitive role of `roles/owner` for the target GCP Project is required

## Project

[CI Pipeline](https://ci.engineerbetter.com/teams/main/pipelines/concourse-up) (deployed with Concourse Up!)

## Development

### Pre-requisites

To build and test you'll need:

- Golang 1.11+
- to have installed `github.com/mattn/go-bindata`

### Building locally

`concourse-up` uses [golang compile-time variables](https://github.com/golang/go/wiki/GcToolchainTricks#including-build-information-in-the-executable) to set the release versions it uses. To build locally use the `build_local.sh` script, rather than running `go build`.

You will also need to clone [`concourse-up-ops`](https://github.com/EngineerBetter/concourse-up-ops) to the same level as `concourse-up` to get the manifest and ops files necessary for building. Check the latest release of `concourse-up` for the appropriate tag of `concourse-up-ops`

### Tests

Tests use the [Ginkgo](https://onsi.github.io/ginkgo/) Go testing framework. The tests require you to have set up AWS authentication locally.

Install ginkgo and run the tests with:

```sh
go get github.com/onsi/ginkgo/ginkgo
ginkgo -r
```

```sh
$ go get github.com/onsi/ginkgo/ginkgo
$ ginkgo -r
```

Go linting, shell linting, and unit tests can be run together in the same docker image CI uses with `./run_tests_local.sh`. This should be done before committing or raising a PR.

### Bumping Manifest/Ops File versions

The pipeline listens for new patch or minor versions of `manifest.yml` and `ops/versions.json` coming from the `concourse-up-ops` repo. In order to pick up a new major version first make sure it exists in the repo then modify `tag_filter: X.*.*` in the `concourse-up-ops` resource where `X` is the major version you want to pin to.
