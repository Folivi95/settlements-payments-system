# Settlements Payment System


## Context

This service is part of the settlements-payments-system.
See [here](https://saltpayco.atlassian.net/wiki/spaces/Settlement/pages/225739054/Payments+Automation) for documentation
around the problem we're solving.

### Related repositories

* [Settlements Payment Publisher](https://github.com/saltpay/settlements-payment-publisher): .NET Core service that
  picks UFX files from a Way4 drop location and puts them into an AWS S3 bucket.

## Architecture diagrams

[Some architecture diagrams](https://github.com/saltpay/settlements-payment-architecture) for a zoomed-out view of the
system and code

## Local development set up

Add to your git configuration (`~/.gitconfig`):

```
[url "git@github.com:"]
	insteadOf = https://github.com
```

Setup go env to work with a private registry so that you can download private deps:

```
go env -w GOPRIVATE="github.com/saltpay"
```

[Install Go](https://golang.org/doc/install)

For your test-double needs you should install moq `go install github.com/matryer/moq@latest`

## Running locally

These instructions are for Unix-based systems

### Pre-requisites

1. [Docker](https://www.docker.com/products/docker-desktop)
2. [Go](https://go.dev/)
3. [Make](https://www.gnu.org/software/make/)

### With Docker Compose

The app depends on AWS SQS and AWS S3 as external systems.

To run the app locally, we spin up the AWS services in localstack.

```
./scripts/docker-run-locally.sh
```

This will run the app in detached mode. You can follow the logs using:

```
docker ps -a | grep "settlements-payments-system_app_run" | awk '{print $1}' | xargs docker logs -f
```

or look for the right container using `docker ps` and then do `docker logs -f {container_id}`

### Run/Debug in Goland/IntelliJ

Add the following entries to `/etc/hosts` file:

```
127.0.0.1       localstack
127.0.0.1       postgres
127.0.0.1       kafka 
```

To spin up the aws services with localstack and postgres db, run `ENV_NAME=local make local`

Go to the IDE and open `cmd/app/main.go` and run **main** with debug mode.

## Migrations

Follow these steps if you are evolving the database structure:

1. Create a new files at `./internal/adapters/postgresql/migrations` with a prefix of the current timestamp
2. When the App is started, the new migration scripts will be automatically picked up 

## Tests

### Unit tests

You can run unit tests locally with Go or in Docker.

Locally: `./scripts/local-run-unit-tests.sh`

In Docker: `./scripts/docker-run-unit-tests.sh`

### All tests, including integration tests

You can run all tests together in Docker: `./scripts/docker-run-unit-and-integration-tests.sh`

## SaltPay Global Platform ##

Our system runs on
the [SaltPay Global Platform](https://saltpayco.atlassian.net/wiki/spaces/INFRAENG/pages/527929016/Platform+Onboarding)
and to be able to access our pipeline, dashboards, endpoints, etc. you will need DEV and PROD [_AWS VPN
Access_](https://saltpayco.atlassian.net/wiki/spaces/INFRAENG/pages/446202778/VPN+-+Connecting+to+AWS+Client+VPN)

Endpoint (dev): https://settlements-payments-system.platform-dev.eu-west-1.salt

Endpoint (prd): https://settlements-payments-system.platform-prd.eu-west-1.salt

### Deployment ###

Merging a PR (todo: enable TBD) will trigger
our [pipeline](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/pipelinerun.yaml) to run
and [deploy](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/deployment.yaml) to dev/staging.

To deploy to prod:

1. Draft a new [release](https://github.com/saltpay/settlements-payments-system/releases)
2. Choose a tag with version number and description e.g. _v0.0.37-refactor-feature-flag-svc_
3. Set the release title same as tag name and click _Auto-generate release notes_
4. Press the _Publish release_ button

Pipeline will run again, deploy to dev/staging and finally prod

### Observability ###

We have a number of tools to be able to see at all time our system internal state determined by its external outputs
such as monitors, logs, metrics, etc.

#### Tools ####

[Argo CD](https://argocd.shared.eu-west-1.salt/applications?proj=&sync=&health=&namespace=&cluster=&labels=&search=settlements-payments-system)

- Kubernetes controller which continuously monitors running applications and compares the current, live state against
  the desired target state.

[Tekton pipeline](https://dashboard.shared.eu-west-1.salt/#/pipelineruns?labelSelector=tki%3Dsettlements-payments-system)

- Kubernetes-native open source framework for creating continuous integration and delivery (CI/CD) systems.

[Grafana Loki](https://o11y-frontend-grafana.shared.eu-west-1.salt/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22loki%22,%7B%22expr%22:%22%22,%22refId%22:%22A%22,%22range%22:true%7D%5D)

- Loki is a horizontally scalable, highly available, multi-tenant log aggregation system inspired by Prometheus

Loki logs can be filtered by applying query e.g. |= "flow_step" for more information
see [Loki query documentation](https://grafana.com/docs/loki/latest/logql/log_queries/)

#### Shared logs ####

[Loki: localstack sidecar](https://o11y-frontend-grafana.shared.eu-west-1.salt/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22loki%22,%7B%22refId%22:%22A%22,%22expr%22:%22%7Bcontainer%3D%5C%22sidecar-localstack%5C%22,%20tki%3D%5C%22settlements-payments-system%5C%22%7D%22%7D%5D)

- S3 and SQS mock used
  in [pipeline](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/pipelinerun.yaml)

[Loki: postgres sidecar](https://o11y-frontend-grafana.shared.eu-west-1.salt/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22loki%22,%7B%22refId%22:%22A%22,%22expr%22:%22%7Bcontainer%3D%5C%22sidecar-postgres%5C%22,%20tki%3D%5C%22settlements-payments-system%5C%22%7D%20%22%7D%5D)

- postgres mock used
  in [pipeline](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/pipelinerun.yaml)

#### Integration (dev) ####

[Settlements Dashboard](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/dashboards/f/f13195f76c2456de1f912e223d7f0a26/settlementspaymentssystem)

[Loki Logs](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/explore?orgId=1&left=%5B%22now-15m%22,%22now%22,%22loki%22,%7B%22expr%22:%22%7Bcontainer%3D%5C%22settlements-payments-system%5C%22%7D%22,%22refId%22:%22A%22,%22range%22:true%7D%5D)

[Prometheus Metrics](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/explore?orgId=1&left=%5B%22now-5m%22,%22now%22,%22prometheus%22,%7B%22refId%22:%22A%22,%22exemplar%22:true,%22expr%22:%22%7Bcontainer%3D%5C%22settlements-payments-system%5C%22%7D%22%7D%5D)

[Compute Resources](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&refresh=10s&var-datasource=default&var-cluster=&var-namespace=settlements-payments)

[Networking](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/d/8b7a8b326d7a6f1f04244066368c67af/kubernetes-networking-namespace-pods?orgId=1&refresh=10s&var-datasource=default&var-cluster=&var-namespace=settlements-payments&var-resolution=5m&var-interval=4h)

[Loki: Black Box Tests](https://o11y-frontend-grafana.platform-dev.eu-west-1.salt/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22loki%22,%7B%22expr%22:%22%7Bcontainer%3D%5C%22black-box-tests%5C%22,%20namespace%3D%5C%22settlements-payments%5C%22%7D%22,%22refId%22:%22A%22,%22range%22:true%7D%5D)

[Kowl: Kafka Topics](https://saltdata-platform-kowl.platform-dev.eu-west-1.salt/topics/settlements-payments-system-transactions) 

#### Production (prd) ####

[Settlements Dashboard](https://o11y-frontend-grafana.platform-prd.eu-west-1.salt/dashboards/f/f13195f76c2456de1f912e223d7f0a26/settlementspaymentssystem)

[Loki Logs](https://o11y-frontend-grafana.platform-prd.eu-west-1.salt/explore?orgId=1&left=%5B%22now-15m%22,%22now%22,%22loki%22,%7B%22expr%22:%22%7Bcontainer%3D%5C%22settlements-payments-system%5C%22%7D%22,%22refId%22:%22A%22,%22range%22:true%7D%5D)

[Prometheus Metrics](https://o11y-frontend-grafana.platform-prd.eu-west-1.salt/explore?orgId=1&left=%5B%22now-5m%22,%22now%22,%22prometheus%22,%7B%22refId%22:%22A%22,%22exemplar%22:true,%22expr%22:%22%7Bcontainer%3D%5C%22settlements-payments-system%5C%22%7D%22%7D%5D)

[Compute Resources](https://o11y-frontend-grafana.platform-prd.eu-west-1.salt/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&refresh=10s&var-datasource=default&var-cluster=&var-namespace=settlements-payments)

[Networking](https://o11y-frontend-grafana.platform-prd.eu-west-1.salt/d/8b7a8b326d7a6f1f04244066368c67af/kubernetes-networking-namespace-pods?orgId=1&refresh=10s&var-datasource=default&var-cluster=&var-namespace=settlements-payments&var-resolution=5m&var-interval=4h)

### Provisioning Infrastructure ###

We can provision our system dependencies by adding specifications
in [infra.yaml](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/infra.yaml)

SaltPay Cloud Engineering documentation can be
found [here](https://saltpayco.atlassian.net/wiki/spaces/INFRAENG/pages/3753316348/04.+Infrastructure)

If you make changes to [infra.yaml](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/infra.yaml)
, don't forget the ``atlantis apply`` step in the [infra HCL repo](https://github.com/saltpay/infra-provision-hcl)

### Secrets ###

All AWS secrets are self-managed, but we can create user-defined secrets by adding specifications
in [secrets.yaml](https://github.com/saltpay/settlements-payments-system/blob/main/.cicd/secrets.yaml)

SaltPay Cloud Engineering documentation can be
found [here](https://saltpayco.atlassian.net/wiki/spaces/INFRAENG/pages/3753283928/03.+Secrets)

We encrypt secrets by using the [infra-secrets-tooling](https://github.com/saltpay/infra-secrets-tooling) (referenced in
documentation above) and we need to make sure they match
our [app catalog](https://github.com/saltpay/saltpay-app-catalog/blob/main/settlements-payments-system.json)

e.g. for updating the `payments-api-authorised-users` secret, perform the following steps:

1. Clone the infra tooling repo
2. Call `make` in that repo to build it
3. Keep the new plaintext value of the secret handy
4. Put the plaintext secret in ~/github/saltpay/settlements-payments-system./cicd/examples/plaintext-secret.yaml ; note
   the escaped quotes and all key-values on one line.
5. Prep this
   command: `./scripts/secret-update.py -r eu-west-1 -n settlements-payments -o platform-dev -t settlements-payments-system -c ~/github/saltpay/settlements-payments-system/.cicd/examples/encrypted-secret.yaml -s payments-api-authorised-users -d ~/github/saltpay/settlements-payments-system/.cicd/examples/plaintext-secret.yaml`
   a. Ensure the paths to the plaintext and encrypted files are correct b. Ensure the `-s` value is the name of the
   secret without the service (tki/name) prefix c. Check all other values
6. Run the command in the infra tooling repo clone
7. Copy the value from the `examples/encrypted-secret.yaml` to the correct part in the `.cicd/secrets.yaml`. Note
   the `scope` to ensure you're adding to the right environment.

## Executing payments in Production

0) check whether we have adequate funding before making the payment; ask Chiryne/Sara/Enrique

1) take a note of the number of payments in file (checksum) - file is in the network location; Einar currently is the
   only one with access

2) make a publisher change to add a filter for the filename we want to process (include todays date in the file); commit
   to auto-deploy the publisher change

3) Payments should start processing automatically now; start looking at the dashboard + logs

4) flow_step 5 -- count of the payment instructions created;

5) 11a -- is the success (also on the dashboard)

6) 11b is failed on the final status (also on the dashboard)

7) "MakeBankingCirclePayment failed" gives the validation errors from BC

8) (5) + (6) + (7) should be equal to (4)

9) check with the treasury team to make sure everything has gone through

# 🏃Tilt
Deploying payment-system in our cluster.
This may take a few minutes if the cluster doesn't already exist / it's your first time running this command / you `make tilt-down`.
If the dependencies are not running it will first get them deployed before deploying ISB-Service. (If your running on an M1 localstak may restart a few times, if it fails re-run `make tilt-up`)
🤑 Note: We are running payment-system in CI mode so NO hot-reloading happens. If you want to see the changes you need to redeploy the service. By running the same command:
```shell
make tilt-up
```

Alternatively you can use `tilt up` for hot reloading.

If you want to run payment-system outside the cluster you can spin just the dependencies.
```shell
make tilt-deps
```

☝️ Note: since the dependencies are running inside the cluster we have to expose them to the outside for payment-system to communicate with them `make expose-services`. 
Read below for more information.

### Exposing Services
Everything should now be running in the cluster so ports exposed by [Payment-System, LocalStack, Kafka, Postgres] are not exposed to your local machine.
For example, when payment-system is running in the cluster, when making a request to `localhost:8080/internal/healthcheck` will return a 404.
To make TCP requests to these services requires exposing them with the following command:

Expose ALL Services
```shell
make expose-services
```
Now calling `localhost:8080/internal/healthcheck` should return healthy. 

Exposing payment-system
```shell
make expose-isb
```

Exposing localstack
```shell
make expose-localstack
```

Exposing Postgres
```shell
make expose-pg
```

Exposing Kafka
```shell
make expose-kafka
```

Exposing Fake Providers
```shell
make expose-fakes
```

## Interacting with cluster resources
Uploading a file to the s3 bucket (prerequisite: ensure localstack is exposed)
```shell
make upload-ufx abspath=~/Downloads/OIC_Documents_RB_BORGUN_20220812_2.xml_352_5.xml
```

## Misc

### Using our webserver

We have an HTTP server with some helpful endpoints. They're detailed inside `scripts/requests.http`, if you're using
intellij you can execute them through the IDE. You'll need a bearer token (create a token and send it to Akash to add to
AWS secrets) (TODO: better instructions than that :D )

To run them you'll need an "environments" file (`http-client.private.env.json`)
. [See here for instructions](https://www.jetbrains.com/help/idea/exploring-http-syntax.html#example-working-with-environment-files)

Your file should look like this

```json
{
  "live": {
    "bearer": "your bearer token"
  }
}
```

So long as you name your file as documented you wont accidentally commit it, it's in the gitignore.

### Export all environment variables from local.env

`export $(grep -v '^#' local.env | xargs)`

### Useful AWS commands for localstack

Pre-requisite :
[Install AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)

List SQS queues:

```
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

Send a message to SQS:

```
aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/local-settlements-payments-bulk-ufx-payment-files --message-body "{\"Records\":[{\"s3\":{\"object\":{\"key\":\"my-object\"}}}]}"
```

Create a bucket on S3:

```
aws --endpoint-url=http://localhost:4566 s3api create-bucket --bucket my-bucket
```

Upload a file to S3:

```
aws --endpoint-url=http://localhost:4566 s3api put-object --bucket my-bucket --key my-object --body ./tmp.txt
```

Download a file from S3:

```
aws --endpoint-url=http://localhost:4566 s3api get-object --bucket my-bucket --key my-object tmp2.txt
```

### Banking Circle certificate conversion to PEM

Decode the private key:

```
openssl pkcs12 -in cert.pfx -out key.pem -nodes -clcerts
```

Decode the cert:

```
openssl pkcs12 -in cert.pfx -out cert.pem -nodes -cacerts
```

Then the private key in PEM format is the text block in `key.pem` beginning with `-----BEGIN PRIVATE KEY-----` and
ending with `-----END PRIVATE KEY-----`, both inclusive.

And the client certificate in PEM format is an append of two blocks:

1) the `-----BEGIN CERTIFICATE-----` and `-----END CERTIFICATE-----` block in key.pem
2) the `-----BEGIN CERTIFICATE-----` and `-----END CERTIFICATE-----` block in cert.pem

---------------------------


