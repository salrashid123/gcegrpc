# GKE gRPC Ingress LoadBalancing

Baseline samle showing gRPC clients connecting via Ingress.

In this mode one gRPC connection sends 10 rpc messages.  Ingress L7 intercepts the ssl conneciton and then transmits each RPC back to differnet pods.
Since each RPC goes to differnet endpoints, the load is more evenly distributed between all pods.

- Image: `salrashid123/grpc_backend`
  gRPC: port: `:50051`
  http2: port: `:8081`  (used for LB healthchecks)

## Setup

First setup a static IP:

```
gcloud compute addresses create gke-ingress --global

gcloud compute addresses list
NAME         REGION  ADDRESS        STATUS
gke-ingress          35.241.41.138  RESERVED
```

the static address `gke-address` is referenced in the GKE ingress file later.

Setup a GKE clsuter

```
gcloud container  clusters create cluster-grpc --zone us-central1-a  --num-nodes 3
```

setup a firewall rule to test direct access to the gRPC server via Network LB (just to show the diffence)

```
gcloud compute firewall-rules create grpc-nlb-firewall --allow tcp:50051
```

Deploy

```
kubectl apply -f fe-deployment.yaml -f fe-ingress.yaml -f fe-secret.yaml -f fe-srv-ingress.yaml -f fe-srv-lb.yaml 
```

Wait maybe 10 mins for the Ingress object to give an IP and provision the LB (yes, it may take up 10mins)


```
$ kubectl get no,po,rs,ing,svc
NAME                                             STATUS    ROLES     AGE       VERSION
no/gke-cluster-grpc-default-pool-8350b7da-34ss   Ready     <none>    1h        v1.10.5-gke.4
no/gke-cluster-grpc-default-pool-8350b7da-4n0q   Ready     <none>    1h        v1.10.5-gke.4
no/gke-cluster-grpc-default-pool-8350b7da-ctbh   Ready     <none>    1h        v1.10.5-gke.4

NAME                               READY     STATUS    RESTARTS   AGE
po/fe-deployment-db5bbf479-5pl7q   1/1       Running   0          1m
po/fe-deployment-db5bbf479-ft899   1/1       Running   0          1m
po/fe-deployment-db5bbf479-n95wn   1/1       Running   0          1m
po/fe-deployment-db5bbf479-sjc7q   1/1       Running   0          1m

NAME                         DESIRED   CURRENT   READY     AGE
rs/fe-deployment-db5bbf479   4         4         4         1m

NAME             HOSTS     ADDRESS         PORTS     AGE
ing/fe-ingress   *         35.241.41.138   80, 443   1m

NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP       PORT(S)                          AGE
svc/fe-srv-ingress   NodePort       10.23.250.59    <none>            50051:31897/TCP,8081:31051/TCP   1m
svc/fe-srv-lb        LoadBalancer   10.23.253.156   104.155.151.124   50051:30122/TCP,8081:32144/TCP   1m
svc/kubernetes       ClusterIP      10.23.240.1     <none>            443/TCP                          1h
```

Note the Loadbalancer and Ingress IPs assigned.  In the example above, 

- LB: `104.155.151.124`
- Ingress: `35.241.41.138 `


Make gRPC calls via the LB:

The response back shows the podname that handled the request.  Note all 10 requests from the client is handled by one pod.  This is imbalanced load.
```
docker run --add-host server.domain.com:104.155.151.124  -t gcr.io/mineral-minutia-820/grpc_backend /grpc_client --host server.domain.com:50051
2018/09/08 08:18:46 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-ft899" 
2018/09/08 08:18:47 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-ft899" 
2018/09/08 08:18:48 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-ft899" 
2018/09/08 08:18:49 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-ft899" 
```


Now connect via the ingress L8.  Each response is from a differnt pod (blanced)

```
$ docker run --add-host server.domain.com:35.241.41.138   -t gcr.io/mineral-minutia-820/grpc_backend /grpc_client --host server.domain.com:443
2018/09/08 08:24:33 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-5pl7q" 
2018/09/08 08:24:34 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-ft899" 
2018/09/08 08:24:36 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-sjc7q" 
2018/09/08 08:24:37 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-sjc7q" 
2018/09/08 08:24:38 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-5pl7q" 
2018/09/08 08:24:39 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-sjc7q" 
2018/09/08 08:24:40 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-db5bbf479-5pl7q" 
```