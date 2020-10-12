#!/bin/sh
set -o errexit

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  echo "Deploying kind local registry"
  docker run \
    -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi


cluster=$(kind get clusters  | grep nginx)
if [[ "$cluster" != "nginx" ]]; then
  echo "Creating kind nginx cluster"
  kind create cluster \
  --name nginx \
  --config kind/Cluster-multi-nodes-1.17
else
  echo "Cluster nginx exists."
fi

echo "Check if docker kind network exists"
docker network ls -f name=kind
if [[ $? -gt 0 ]]; then
  echo "Creating the docker kind network"
  docker network connect "kind" "${reg_name}"
fi

echo "Annotate nginx nodes"
# tell https://tilt.dev to use the registry
# https://docs.tilt.dev/choosing_clusters.html#discovering-the-registry
for node in $(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do
  kubectl annotate node "${node}" "kind.x-k8s.io/registry=localhost:${reg_port}";
  kubectl annotate node "${node}" "kind.x-k8s.io/registry-from-cluster=${reg_name}:${reg_port}"
done
