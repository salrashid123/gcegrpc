# GKE gRPC Ingress LoadBalancing

Sample showing gRPC clients connecting via Ingress.

In this mode one gRPC connection sends 10 rpc messages.  Ingress L7 intercepts the ssl connection and then transmits each RPC back to different pods.
Since each RPC goes to different endpoints, the load is more evenly distributed between all pods.

>> **Update 8/10/20**:  GCP now support a BackendConfig that supports independent HealthChecks over HTTP that Ingress understands:

- [Custom health check configuration](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#direct_health)


The BackendConfig makes the workarounds using mux and envoy described in previous commits in this repo obsolete but you still need an a POD that proxies HTTP healthcheck requests.

You still need to run an HTTP listener Container on the same gRPC Service POD (ie run your grpc service in one container and run an http proxy healthcheck in another).  Previously, you had to effectively run HTTP and gRPC on the same Serving Port.

For example, you can run an HTTP listener on `:8080` and a GRPC service on `:50051`.  GCP will send healthcheck requests to `:8080` which makes a GRPC healtcheck request internally to `:50051`

- `podSpec(http_grpc_proxy:8080, grpc_service:50051)`

The following [grpc_health_proxy](https://github.com/salrashid123/grpc_health_proxy) translates HTTP requests into gRPC HealCheck Protocol.

To deploy, start GKE Cluster `1.18` or higher

```
gcloud container  clusters create cluster-grpc \
 --zone us-central1-a  --num-nodes 3 --enable-ip-alias \
 --cluster-version "1.18"
```

```bash
cd gcegrpc/gke_ingress_lb
kubectl apply -f .
```

Wait for the Ingress and Loadbalancer configurations to allocate an IP and test as described below.
(wait maybe 8mins)


Look at `fe-srv-ingress.yaml` file for the `BackendConfig`:


```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: fe-srv-ingress
  labels:
    type: fe-srv
  annotations:
    cloud.google.com/app-protocols: '{"fe":"HTTP2"}'
    cloud.google.com/neg: '{"ingress": true, "exposed_ports": {"50051":{}}}'
    cloud.google.com/backend-config: '{"default": "fe-grpc-backendconfig"}'
spec:
  type: ClusterIP 
  ports:
  - name: fe
    port: 50051
    protocol: TCP
    targetPort: 50051
  selector:
    app: fe
---
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: fe-grpc-backendconfig
spec:
  healthCheck:
    type: HTTP2
    requestPath: /
    port: 8080
```


What that describes is http healthcheck that will send requests to `/` on port `:8080`.  The Ingress rule specifies HTTP2 traffic to port `:50051` for `selector: app:fe` and also uses the healthchecks defined by `fe-grpc-backendconfig`.

See [Creating a Service for a container-native load balancer](https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing)

>>  A Service of type ClusterIP is recommended unless you explicitly need the nodePort provided by a NodePort Service.


>> **NOTE**: HTTP2 Healthchecks on GCP _requires_ https ([health-check-concepts](https://cloud.google.com/load-balancing/docs/health-check-concepts#category_and_protocol))

A couple of notes about SSL.  The configuration described in this article uses TLS from start to finish:

- `client -> SSL -> L7LB (ingress) -> SSL -> (gRPC Application Service)`
- `GCP HTTP2 Healtcheck -> SSL -> (healthCheck Proxy) --> SSL (gRPC HealthCheck Service)`

The effective configuration for the HealthCheck Proxy then handles TLS from GCP's Healthcheck and also makes a new TLS connection to the service POD
```yaml
    spec:
      containers:
      - name: hc-proxy
        image: docker.io/salrashid123/grpc_health_proxy:1.0.0
        args: [
          "--http-listen-addr=0.0.0.0:8080",
          "--grpcaddr=localhost:50051",
          "--service-name=echo.EchoServer",
          "--https-listen-ca=/config/CA_crt_hc.pem",
          "--https-listen-cert=/certs/http_server_crt.pem",
          "--https-listen-key=/certs/http_server_key.pem",
          "--grpctls",        
          "--grpc-sni-server-name=server.domain.com",
          "--grpc-ca-cert=/config/CA_crt_grpc_server.pem",
          "--logtostderr=1",
          "-v=1"
        ]
      - name: grpc-app
        image: salrashid123/grpc_only_backend
        args: [
          "/grpc_server",
          "--grpcport=0.0.0.0:50051",
          "--tlsCert=/certs/grpc_server_crt.pem",
          "--tlsKey=/certs/grpc_server_key.pem"        
        ]
        ports:
        - containerPort: 50051    
```

---

## Setup

```bash

$ gcloud container  clusters create cluster-grpc \
   --zone us-central1-a  --num-nodes 3 --enable-ip-alias \
   --cluster-version "1.19"
```


If using ILB, see [Setting up ILB Subnet](https://cloud.google.com/load-balancing/docs/l7-internal/setting-up-l7-internal#configuring_the_proxy-only_subnet)

first create an ILB subnet in the appropriate range (in this case, its `10.5.0.0/20`)
```bash
$ gcloud compute networks subnets create proxy-only-subnet  \
   --purpose=INTERNAL_HTTPS_LOAD_BALANCER   --role=ACTIVE \
   --region=us-central1   --network=default   --range=10.5.0.0/20
```

Then, 

```bash
$ kubectl apply -f .

(wait 8 mins, really wait)

$  kubectl get po,rs,ing,svc
NAME                                READY   STATUS    RESTARTS   AGE
pod/fe-deployment-5956bb98d-7l4g8   2/2     Running   0          14m
pod/fe-deployment-5956bb98d-sm7xc   2/2     Running   0          14m

NAME                                      DESIRED   CURRENT   READY   AGE
replicaset.apps/fe-deployment-5956bb98d   2         2         2       14m

NAME                                CLASS    HOSTS   ADDRESS         PORTS     AGE
ingress.extensions/fe-ilb-ingress   <none>   *       10.128.0.95     80, 443   14m
ingress.extensions/fe-ingress       <none>   *       34.120.99.146   80, 443   14m

NAME                     TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)           AGE
service/fe-srv-ingress   NodePort       10.4.0.141   <none>          50051:31473/TCP   14m
service/fe-srv-lb        LoadBalancer   10.4.0.30    35.224.74.161   50051:32216/TCP   14m
service/kubernetes       ClusterIP      10.4.0.1     <none>          443/TCP           15m
```

### External L7 LB

Test with Network LB:

```bash
$ docker run --add-host server.domain.com:35.224.74.161 \
  -t salrashid123/grpc_backend /grpc_client \
  --host server.domain.com:50051

2020/09/21 21:38:00 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:01 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:02 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:03 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:04 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:05 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:06 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:07 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:08 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:09 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8"
```

Now test with the Ingress LB:

```bash
$ docker run --add-host server.domain.com:34.120.99.146 \
  -t salrashid123/grpc_backend /grpc_client \
  --host server.domain.com:443

2020/09/21 21:38:13 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:38:14 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:38:15 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:16 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:38:17 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:38:19 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:20 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:21 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:22 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:38:23 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8"
```


### ILB:

On a VM within GCP:
```bash
$ docker run --add-host server.domain.com:10.128.0.95 \
   -t salrashid123/grpc_backend /grpc_client \
   --host server.domain.com:443

2020/09/21 21:39:55 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:39:56 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:39:57 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:39:58 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:39:59 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:40:00 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:40:01 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:40:02 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc" 
2020/09/21 21:40:03 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-7l4g8" 
2020/09/21 21:40:04 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-5956bb98d-sm7xc"
```
