# do-k8s-cluster-health-check
Provides status on k8s cluster health
https://confluence.wiley.com/display/DEVOPS/Kubernetes+Cluster+Health+Check

## Building and testing

### Binary
To test the project, run
```sh
go test ./...
```
in the project root directory.


To build the project, run
```sh
go build -o chc
```
in the project root directory.

### Docker

Set the correct image tag. Either extract it from `helm/Chart.yaml` file or
set it manually.
```
export APP_VERSION=`grep 'appVersion' helm/Chart.yaml | cut -d: -f 2 | sed 's/^[[:blank:]]*//'` \
&& echo "APP_VERSION: $APP_VERSION"
```

Build the image first
1) With traditional docker build engine
```
docker build \
-t 681504496077.dkr.ecr.us-east-1.amazonaws.com/chc:${APP_VERSION} .
```

2) Or with [buildkit](https://github.com/moby/buildkit) to build OCI images
```
docker buildx build \
-t 681504496077.dkr.ecr.us-east-1.amazonaws.com/chc:${APP_VERSION} .
```

Push Docker image to ECR
```
docker push 681504496077.dkr.ecr.us-east-1.amazonaws.com/chc:${APP_VERSION}
```

## Deploy to Kubernetes

Assumptions:
1. Use [helm v3](https://helm.sh/docs/intro/install/)
2. Docker image is available in ECR
3. Make sure `chc` namespace exists in k8s. You need RBAC privileges
   to create it.
   `kubectl create ns chc`

Upgrade `chc` release; Install the release if missing.
```bash
helm upgrade \
   chc helm/ \
   --install \
   --namespace chc \
   -f helm_vars/wpng/dev/values.yaml \
   --debug --dry-run
```
> Remove the `--dry-run` flag before flight!

<br />

## Configuration

This application expects to receive system parameters in basically 3 ways, respecting their order of precedence:

1. CLI flags: override everything else;
2. Environment Variables: overrides YAML config;
3. YAML config: is the lowest priority configuration source;

>**Note:** Default CLI Flags values are the fallback if any of the configuration sources are being used, except if the parameter is required. See the `Required` column in the [parameters](#parameters) section.

<br />

### Configuration file observations

**Using a custom file:** 

The config file must be in `YAML` format, and can be located anywhere in the filesystem, as long as the file location is provided through the `--config` flag.

<br />

**Default file:**

The default config file that is going to be used if the `--config` parameter isn't set is located at:

CLI: [`./config/config.yml`](./config/config.yml)

Helm\*: [`./helm/config/config.yml`](./helm/config/config.yml)

>\*The file is going to be parsed by the Helm templating system and injected as a configmap data into the deployment pods.


<br />

### Parameters

Refer to the following table in order to get to know all the parameters that can be configured through any of the aforementioned config sources:

<br />

| CLI Flag                      | Environment Variable            | YAML parameter        | Required\* | Description                                                              | Default                                                     |
|-------------------------------|---------------------------------|-----------------------|-----------|---------------------------------------------------------------------------|-------------------------------------------------------------|
| `--listen-address`, `-l`      | `HEALTHCAT_LISTEN_ADDRESS`      | `listen-address`      | No        | Bind address                                                              | `"*"`                                                       |
| `--cluster-id`, `-i`          | `HEALTHCAT_CLUSTER_ID`          | `cluster-id`          | Yes       | The cluster ID                                                            | not applicable                                              |
| `--namespaces`, `-n`          | `HEALTHCAT_NAMESPACES`          | `namespaces`          | No        | List of namespaces to watch                                               | `""`                                                        |
| `--excluded-namespaces`, `-N` | `HEALTHCAT_EXCLUDED_NAMESPACES` | `excluded-namespaces` | No        | List of namespaces to exclude                                             | `"kube-system,default,kube-public,istio-system,monitoring"` |
| `--time-between-hc`, `-t`     | `HEALTHCAT_TIME_BETWEEN_HC`     | `time-between`        | No        | Interval between two consecutive health checks                            | `"1m"`                                                      |
| `--successful-hc-cnt`, `-s`   | `HEALTHCAT_SUCCESSFUL_HC_CNT`   | `successful-hc`       | No        | Number of successful consecutive health checks counts                     | `1`                                                         |
| `--failed-hc-cnt`, `-F`       | `HEALTHCAT_FAILED_HC_CNT`       | `failed-hc`           | No        | Number of failed consecutive health checks counts                         | `2`                                                         |
| `--status-threshold`, `-P`    | `HEALTHCAT_STATUS_THRESHOLD`    | `status-threshold`    | No        | Percentage of successful health checks to set cluster status as OK        | `100`                                                       |
| `--port`, `-p`                | `HEALTHCAT_PORT`                | `port`                | No        | Bind port                                                                 | `8080`                                                      |
| `--log-preset`                | `HEALTHCAT_LOG_PRESET`          | `log-preset`          | No        | Log preset config (dev\|prod)                                             | `"dev"`                                                     |
| `--config`, `-f`              | not applicable                  | not applicable        | No        | Path to the config file to be used as an alternative configuration source | `"./config/config.yml"`                                     |

>\*If the parameter is required, that means it doesn't have a corresponding default value and therefore it must be provided by any of the following configuration sources: CLI Flag, Env. or Config File.

<br />

[Back to the top](#do-k8s-cluster-health-check)
