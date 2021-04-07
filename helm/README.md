# Cluster Health Check

# Install

```
helm upgrade \
chc helm/ \
--install \
--namespace chc \
-f helm_vars/wpng/dev/values.yaml \
--set image.tag=dev \
--debug --dry-run
```

# Blue/Grin

## Deploy to blue

If missing
```
kubectl create ns chc-blue
```

Deploy
```
helm upgrade \
chc helm/ \
--install \
--namespace chc-blue \
-f helm_vars/edpub/nonprod/values-blue.yaml \
--debug --dry-run

helm upgrade \
podinfo podinfo/ \
--install \
--namespace chc-blue \
-f helm_vars/podinfo/r53-failover/values-blue.yaml \
--debug --dry-run

```
