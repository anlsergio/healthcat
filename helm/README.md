# Cluster Health Check

# Install

Create namespace
```
kubectl create ns healthcat
```

## Deploy

### EdPub NonProd
```
helm upgrade \
healthcat helm/ \
--install \
--namespace healthcat \
-f helm_vars/edpub/nonprod/values.yaml \
--set image.tag=eefe09e \
--debug --dry-run
```
### Local registry
```
helm upgrade \
healthcat helm/ \
--install \
--namespace healthcat \
-f helm_vars/local/dev/values.yaml \
--debug --dry-run
```
