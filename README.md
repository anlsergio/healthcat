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
export APP_VERSION=`grep 'appVersion' helm/Chart.yaml | cut -d: -f 2 | sed 's/^[[:blank:]]*//'`
echo "APP_VERSION: $APP_VERSION"
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
```
$ cd do-k8s-cluster-health-check

$ helm upgrade \
chc helm/ \
--install \
--namespace chc \
-f helm_vars/wpng/dev/values.yaml \
--debug --dry-run
```
