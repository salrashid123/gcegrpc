# gRPC-GKE Service LoadBalancer "helloworld"


Sample application demonstrating RoundRobin gRPC loadbalancing  on Kubernetes for interenal services.


gRPC can stream N messages over one connection.  When k8s services are involved, a single connection to the destination will terminate
at one pod.  If N messages are sent from the client, all N messages will get handled by that pod resulting in imbalanced load.

One way to address this issue is insteaad define the remote service as [Headless](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) and then use gRPC's client side loadbalancing constructs.

In this mode, k8s service does not return the single destination for Service but instead all destination addresses back to a lookup request.

Given the set of ip addresses, the grpc client will send each rpc to different pods and evenly distribute load.

## Setup

### Setup Applicaiton

Create the image in  ```cd ~app/http_frontend```

or use the one i've setup `docker.io/salrashid123/http_frontend`

What the image under

[app/http_frontend](app/http_frontend) is an app that shows

- `/backend`:  make 10 gRPC requests over one connection via 'ClusterIP` k8s Service.  The expected response is all from one node

- `/backendlb`:  make 10 gRPC requests over one connection via k8s Headless Service.  The expected response is from different nodes


### Create a cluster
```
gcloud container  clusters create cluster-grpc --zone us-central1-a  --num-nodes 4
```

### Deploy

```
kubectl apply -f be-deployment.yaml  -f be-srv-lb.yaml  -f be-srv.yaml  -f fe-deployment.yaml  -f fe-srv.yaml
```


Wait ~5mins till the Network Loadblancer IP is assigned

```
$ kubectl get no,po,deployment,svc
NAME                                             STATUS    ROLES     AGE       VERSION
no/gke-grpc-cluster-default-pool-fb759232-jz6f   Ready     <none>    1h        v1.10.5-gke.4
no/gke-grpc-cluster-default-pool-fb759232-mf0v   Ready     <none>    1h        v1.10.5-gke.4
no/gke-grpc-cluster-default-pool-fb759232-nw2z   Ready     <none>    1h        v1.10.5-gke.4

NAME                                READY     STATUS    RESTARTS   AGE
po/be-deployment-6dc58d6898-2bpmh   1/1       Running   0          51s
po/be-deployment-6dc58d6898-5v2xf   1/1       Running   0          51s
po/be-deployment-6dc58d6898-jgj27   1/1       Running   0          51s
po/be-deployment-6dc58d6898-kb4cj   1/1       Running   0          51s
po/be-deployment-6dc58d6898-krnb5   1/1       Running   0          51s
po/fe-deployment-75756779d8-fmhtv   1/1       Running   0          51s

NAME                   DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/be-deployment   5         5         5            5           51s
deploy/fe-deployment   1         1         1            1           51s

NAME             TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)              AGE
svc/be-srv       ClusterIP      10.19.246.79    <none>           50051/TCP,8081/TCP   51s
svc/be-srv-lb    ClusterIP      None            <none>           50051/TCP,8081/TCP   51s
svc/fe-srv       LoadBalancer   10.19.249.102   104.154.194.89   8081:30388/TCP       51s
svc/kubernetes   ClusterIP      10.19.240.1     <none>           443/TCP              1h
```

### Connect via k8s Service

```golang
	address := fmt.Sprintf("be-srv.default.svc.cluster.local:50051")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce))
```

```
$ curl -sk https://104.154.194.89:8081/backend | jq '.'
{
  "frontend": "fe-deployment-75756779d8-fmhtv",
  "backends": [
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27"
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

	conn, err := grpc.Dial("dns:///be-srv-lb.default.svc.cluster.local", grpc.WithTransportCredentials(ce), grpc.WithBalancerName(roundrobin.Name))
	c := echo.NewEchoServerClient(conn)
```


```
$ curl -sk https://104.154.194.89:8081/backendlb | jq '.'
{
  "frontend": "fe-deployment-75756779d8-fmhtv",
  "backends": [
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-5v2xf",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-5v2xf",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-kb4cj",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-2bpmh",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-krnb5",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-5v2xf",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-kb4cj",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-jgj27",
    "Hello unary RPC msg   from hostname be-deployment-6dc58d6898-2bpmh"
  ]
}
```

Note: responses are distributed evenly.

## References
 - [https://github.com/jtattermusch/grpc-loadbalancing-kubernetes-examples#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy](https://github.com/jtattermusch/grpc-loadbalancing-kubernetes-examples#example-1-round-robin-loadbalancing-with-grpcs-built-in-loadbalancing-policy)
 - [https://kca.id.au/post/k8s_service/](https://kca.id.au/post/k8s_service/)

