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

Create a cluster called `nginx` if missing.
```
$ cd test/k8s
$ kind create cluster --name nginx \
--config kind/Cluster-multi-nodes-1.17
Creating cluster "nginx" ...
 ✓ Ensuring node image (kindest/node:v1.17.11) 🖼
 ✓ Preparing nodes 📦 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
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

### Deploy CHC
Build and deploy using the make command.
```
$ cd do-k8s-cluster-health-check
$ make docker-kind
```

### Create pods manifests

Modify `POD_TOTAL` var if necessary, and then run `./build-pods.sh` script. The resulting k8s manifest will contain the Pods and Services required for the test.

Apply the manifest. Always use a temporary namespace to be able to clean up tests resources easily.
```
$ kubectl -n nginx apply -f pods-test.yaml
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
pod/nginx-11 created
service/nginx-11 created
pod/nginx-12 created
service/nginx-12 created
pod/nginx-13 created
service/nginx-13 created
pod/nginx-14 created
service/nginx-14 created
pod/nginx-15 created
service/nginx-15 created
pod/nginx-16 created
service/nginx-16 created
pod/nginx-17 created
service/nginx-17 created
pod/nginx-18 created
service/nginx-18 created
pod/nginx-19 created
service/nginx-19 created
pod/nginx-20 created
service/nginx-20 created
```

Check the number of Pods with and without a `/healthz` endpoint
```
$ kubectl -n nginx get pods -l health=ok
NAME       READY   STATUS    RESTARTS   AGE
nginx-1    1/1     Running   0          106s
nginx-10   1/1     Running   0          99s
nginx-11   1/1     Running   0          99s
nginx-12   1/1     Running   0          98s
nginx-13   1/1     Running   0          97s
nginx-14   1/1     Running   0          97s
nginx-15   1/1     Running   0          96s
nginx-16   1/1     Running   0          95s
nginx-17   1/1     Running   0          94s
nginx-18   1/1     Running   0          92s
nginx-19   1/1     Running   0          91s
nginx-2    1/1     Running   0          106s
nginx-20   1/1     Running   0          90s
nginx-5    1/1     Running   0          103s
nginx-6    1/1     Running   0          102s
nginx-8    1/1     Running   0          101s
nginx-9    1/1     Running   0          100s

$ kubectl -n nginx get pods -l health=notok
NAME      READY   STATUS    RESTARTS   AGE
nginx-3   1/1     Running   0          110s
nginx-4   1/1     Running   0          109s
nginx-7   1/1     Running   0          106s
```


Start the port proxy tunnel
```
 $ kubectl -n chc port-forward chc-6f9bcdc45f-hn7m7  8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

Check the `/services` and `/status` endpoints
```
 curl http://localhost:8080/services | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   794  100   794    0     0   2255      0 --:--:-- --:--:-- --:--:--  2262
{
  "cluster": {
    "name": "wpngdev",
    "healthy": false,
    "total": 20,
    "failed": 20
  },
  "services": [
    {
      "name": "nginx-2",
      "healthy": false
    },
    {
      "name": "nginx-4",
      "healthy": false
    },
    {
      "name": "nginx-6",
      "healthy": false
    },
    {
      "name": "nginx-8",
      "healthy": false
    },
    {
      "name": "nginx-13",
      "healthy": false
    },
    {
      "name": "nginx-5",
      "healthy": false
    },
    {
      "name": "nginx-10",
      "healthy": false
    },
    {
      "name": "nginx-14",
      "healthy": false
    },
    {
      "name": "nginx-18",
      "healthy": false
    },
    {
      "name": "nginx-1",
      "healthy": false
    },
    {
      "name": "nginx-3",
      "healthy": false
    },
    {
      "name": "nginx-7",
      "healthy": false
    },
    {
      "name": "nginx-9",
      "healthy": false
    },
    {
      "name": "nginx-11",
      "healthy": false
    },
    {
      "name": "nginx-12",
      "healthy": false
    },
    {
      "name": "nginx-20",
      "healthy": false
    },
    {
      "name": "nginx-15",
      "healthy": false
    },
    {
      "name": "nginx-16",
      "healthy": false
    },
    {
      "name": "nginx-17",
      "healthy": false
    },
    {
      "name": "nginx-19",
      "healthy": false
    }
  ]
}
```

Get the `/status`
```
$ curl http://localhost:8080/status
Failure
```

### Cleanup
```
kind delete cluster --name nginx
```
