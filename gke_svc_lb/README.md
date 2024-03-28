# gRPC-GKE Service LoadBalancer "helloworld"


Sample application demonstrating RoundRobin gRPC loadbalancing  on Kubernetes for internal services.


gRPC can stream N messages over one connection.  When k8s services are involved, a single connection to the destination will terminate
at one pod.  If N messages are sent from the client, all N messages will get handled by that pod resulting in imbalanced load.

One way to address this issue is insteaad define the remote service as [Headless](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) and then use gRPC's client side loadbalancing constructs.

In this mode, k8s service does not return the single destination for Service but instead all destination addresses back to a lookup request.

Given the set of ip addresses, the grpc client will send each rpc to different pods and evenly distribute load.

`Update 5/14/20`: Also see [xds gRPC Loadbalancing](https://github.com/salrashid123/grpc_xds).  Using xDS client-side only loadbalancing is woudl be the recommended way for internal service->service LB at scale.  However, it is quite experimental at the moment and cannot be easily deployed on GKE/GCE (unless you use an xDS server like Traffic Director).

`Update 5/4/20`: 
Headless service will seed an initial set of POD addresses directly to the client and will not Refresh that set. That means once a client gets the list of IPs it will not know about _new_ pods that kubernetes spins up. It will only keep the existing set (you can remove stale/old clients using the [keepalive](https://godoc.org/google.golang.org/grpc/keepalive) but that will not inform you of new pods)

References: [gRPC Load Balancing on Kubernetes without Tears](https://kubernetes.io/blog/2018/11/07/grpc-load-balancing-on-kubernetes-without-tears/)

One option maybe to create a connection pool handler that periodically refreshes the set: from @teivah : [sample pool implementation](https://github.com/teivah/tourniquet/blob/master/tourniquet.go)

The other is to use xds balancer but that is much too experimental now


## Setup

### Setup Application

Create the image in  ```cd ~app/http_frontend```

or use the one i've setup `docker.io/salrashid123/http_frontend`

What the image under

[app/http_frontend](app/http_frontend) is an app that shows

- `/backend`:  make 10 gRPC requests over one connection via 'ClusterIP` k8s Service.  The expected response is all from one node

- `/backendlb`:  make 10 gRPC requests over one connection via k8s Headless Service.  The expected response is from different nodes


### Create a cluster

```bash
gcloud container  clusters create cluster-grpc --zone us-central1-a  --num-nodes 3 --enable-ip-alias
```

### Deploy

```bash
kubectl apply -f be-deployment.yaml  -f be-srv-lb.yaml  -f be-srv.yaml  -f fe-deployment.yaml  -f fe-srv.yaml
```

Create firewall rule

```bash
gcloud compute firewall-rules create allow-grpc-nlb --action=ALLOW --rules=tcp:8081 --source-ranges=0.0.0.0/0
```

Wait ~5mins till the Network Loadblancer IP is assigned

```bash
$ kubectl get no,po,deployment,svc
NAME                                               STATUS    ROLES     AGE       VERSION
node/gke-cluster-grpc-default-pool-aeb308a0-89dt   Ready     <none>    1h        v1.11.7-gke.12
node/gke-cluster-grpc-default-pool-aeb308a0-hv5f   Ready     <none>    1h        v1.11.7-gke.12
node/gke-cluster-grpc-default-pool-aeb308a0-vsf4   Ready     <none>    1h        v1.11.7-gke.12

NAME                                 READY     STATUS    RESTARTS   AGE
pod/be-deployment-757dd4f4bd-cpgl9   1/1       Running   0          10s
pod/be-deployment-757dd4f4bd-dsxdg   1/1       Running   0          10s
pod/be-deployment-757dd4f4bd-f8v2v   1/1       Running   0          10s
pod/be-deployment-757dd4f4bd-msp5f   1/1       Running   0          10s
pod/be-deployment-757dd4f4bd-qr4sd   1/1       Running   0          10s
pod/fe-deployment-59d7bb7df8-w4fwc   1/1       Running   0          11s

NAME                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/be-deployment   5         5         5            5           10s
deployment.extensions/fe-deployment   1         1         1            1           11s

NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)          AGE
service/be-srv       ClusterIP      10.23.242.113   <none>           50051/TCP        1h
service/be-srv-lb    ClusterIP      None            <none>           50051/TCP        1h
service/fe-srv       LoadBalancer   10.23.246.138   35.226.254.240   8081:30014/TCP   1h
service/kubernetes   ClusterIP      10.23.240.1     <none>           443/TCP          1h
```

### Connect via k8s Service

```golang
	address := fmt.Sprintf("be-srv.default.svc.cluster.local:50051")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce))
```

```
$ curl -sk https://35.226.254.240:8081/backend | jq '.'
{
  "frontend": "fe-deployment-59d7bb7df8-w4fwc",
  "backends": [
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9"
  ]
}
```

Note: all the responses are from one node

### Connect via k8s Headless Service

```golang
import (
   "google.golang.org/grpc/balancer/roundrobin"
   "google.golang.org/grpc/credentials"
)

  address := fmt.Sprintf("dns:///be-srv-lb.default.svc.cluster.local:50051")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce), grpc.WithBalancerName(roundrobin.Name))
	c := echo.NewEchoServerClient(conn)
```


```
$ curl -sk https://35.226.254.240:8081/backendlb | jq '.'
{
  "frontend": "fe-deployment-59d7bb7df8-w4fwc",
  "backends": [
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-dsxdg",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-msp5f",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-qr4sd",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-f8v2v",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-dsxdg",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-msp5f",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-qr4sd",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-f8v2v",
    "Hello unary RPC msg   from hostname be-deployment-757dd4f4bd-cpgl9"
  ]
}
```

Note: responses are distributed evenly.

## References
 - [https://github.com/jtattermusch/grpc-loadbalancing-kubernetes-examples#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy](https://github.com/jtattermusch/grpc-loadbalancing-kubernetes-examples#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy)
 - [https://kca.id.au/post/k8s_service/](https://kca.id.au/post/k8s_service/)
