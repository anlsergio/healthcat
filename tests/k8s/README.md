# Testing CHC

## Test Summary
`buildpods.sh` creates POD_TOTAL number of Pods and Services. Most of the Pods will have `/healthz` nginx configuration while some small number of Pods will be deployed withour the `/healthz` endpoint.

CHC should be able to detect both types of Pods - with and without health check, and provide the corresponding status.

## Test CHC

### Deploy k8s Kind cluster

For more details see https://kind.sigs.k8s.io/docs/user/quick-start/

Check kind clusters
```
$ cd test/k8s
$ kind get clusters
```

Create a cluster called `nginx` if missing using the `Kind` config files.

There are two Kind config files:
1. multi-node cluster - kind/Cluster-multi-nodes-1.17
   One control plane and 2 worker nodes.
2. single-node cluster - kind/Cluster-single-node-1.17
   Control plane node only

The single node cluster is faster to spin.

Create the cluster that suit your needs. Keep the name `nginx`.
It might be used elsewhere.
```
$ cd test/k8s

# create a single node cluster
$ kind create cluster --name nginx \
> --config kind/Cluster-single-node-1.17
Creating cluster "nginx" ...
 ✓ Ensuring node image (kindest/node:v1.17.11) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-nginx"
You can now use your cluster with:

kubectl cluster-info --context kind-nginx

Have a nice day! 👋
```

Inpsect `kind-nginx` context
```
kubectl cluster-info --context kind-nginx
Kubernetes master is running at https://127.0.0.1:54687
KubeDNS is running at https://127.0.0.1:54687/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

### Create namespaces
For CHC and Nginx tests pods
```
kubectl create ns chc
kubectl create ns nginx
```

Create the ConfigMap with additional Nginx config that includes
the liveness probe.
```
kubectl -n nginx apply -f configmap.yaml
```

### Deploy CHC
Build and deploy using the make command.
```
$ cd do-k8s-cluster-health-check
$ make docker-kind
```

### Create pods manifests

Modify `POD_TOTAL` var if necessary, and then run `./build-pods.sh` script. The resulting k8s manifest will contain the Pods and Services required for the test.

An example. Create 10 tests Pods.
```
$ ./build-pods.sh 10
 SLEEP_TIME:
 POD_TOTAL: 10
 IS_HEALTH: notok
 POD_TEST_FILE: pods-test.yaml
 NOTOK_THRESHOLD: 1
Removing file pods-test.yaml ...
Creating Pod# 1
Creating Pod# 2
Creating Pod# 3
Creating Pod# 4
Creating Pod# 5
Creating Pod# 6
Creating Pod# 7
Creating Pod# 8
Creating Pod# 9
Creating Pod# 10
```

Apply the manifest. Always use a temporary namespace to be able to clean up tests resources easily.
```
$ kubectl apply -n nginx -f pods-test.yaml
pod/nginx-1 created
service/nginx-1 created
pod/nginx-2 created
service/nginx-2 created
pod/nginx-3 created
service/nginx-3 created
pod/nginx-4 created
service/nginx-4 created
pod/nginx-5 created
service/nginx-5 created
pod/nginx-6 created
service/nginx-6 created
pod/nginx-7 created
service/nginx-7 created
pod/nginx-8 created
service/nginx-8 created
pod/nginx-9 created
service/nginx-9 created
pod/nginx-10 created
service/nginx-10 created
```

Check the number of Pods with and without a `/healthz` endpoint
```
$ kubectl -n nginx get pods -l health=ok
NAME       READY   STATUS    RESTARTS   AGE
nginx-1    1/1     Running   0          30s
nginx-10   1/1     Running   0          30s
nginx-2    1/1     Running   0          30s
nginx-3    1/1     Running   0          30s
nginx-4    1/1     Running   0          30s
nginx-5    1/1     Running   0          30s
nginx-6    1/1     Running   0          30s
nginx-7    1/1     Running   0          30s
nginx-8    1/1     Running   0          30s
nginx-9    1/1     Running   0          30s

$ kubectl -n nginx get pods -l health=notok
No resources found in nginx namespace.
```

Start the port proxy tunnel
```
# get a chc Pod name
$ kubectl -n chc get pods
$
$ kubectl -n chc port-forward chc-698cff547-v2zr6 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

Check the `/services` and `/status` endpoints
```
 curl -s http://localhost:8080/services | jq .
 {
   "cluster": {
     "name": "wpngdev",
     "healthy": true,
     "total": 10,
     "failed": 0
   },
   "services": [
     {
       "name": "nginx-9",
       "healthy": true
     },
     {
       "name": "nginx-1",
       "healthy": true
     },
     {
       "name": "nginx-4",
       "healthy": true
     },
     {
       "name": "nginx-8",
       "healthy": true
     },
     {
       "name": "nginx-6",
       "healthy": true
     },
     {
       "name": "nginx-7",
       "healthy": true
     },
     {
       "name": "nginx-10",
       "healthy": true
     },
     {
       "name": "nginx-2",
       "healthy": true
     },
     {
       "name": "nginx-3",
       "healthy": true
     },
     {
       "name": "nginx-5",
       "healthy": true
     }
   ]
 }
```

Get the `/status`
```
$ curl http://localhost:8080/status
OK
```

### Cleanup
```
kind delete cluster --name nginx
```
