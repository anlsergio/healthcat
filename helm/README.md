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
