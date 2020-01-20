# GKE gRPC Ingress LoadBalancing

Baseline samle showing gRPC clients connecting via Ingress.

In this mode one gRPC connection sends 10 rpc messages.  Ingress L7 intercepts the ssl conneciton and then transmits each RPC back to differnet pods.
Since each RPC goes to differnet endpoints, the load is more evenly distributed between all pods.

As of 4/19, GKE does not support [gRPC HelthChecks](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) and
instead relies on ordinary HTTP healthchecks which must retrun a `200` (by default against `/` endpoint on the _same_ service
as the gRPC service).
 - [GKE Ingress HealthChecks]https://cloud.google.com/kubernetes-engine/docs/concepts/ingress#health_checks)

What this implies is you cannot use the same gRPC listener port within your deployment as-is.  You must implement a mux handler or proxy
that processes both oridnary http2 requests for healthchecks and the gRPC service itself.

For example, `/` endpoint is handled by your application as a healthcheck that can return `200 OK` back to GCE's healtcheck.
GRPC requests for aservice (eg `/echo.EchoService/SayHello`) must be routeable on the _same_ port.  This requires either a mux capable of
both HTTP2 and gRPC requests (the latter over http2, ofcourse)

There are two variations/implementations within the `gke_ingress_lb/` folder that demonstrates these workarounds:

* 

* `gke_ingress_lb/gke_ingress_lb_mux`:  golang mux handler where one port `:50051` on the container can process HTTP2 traffic that is
both non-GRPC and HTTP2 healthchecks for the `/` endpoint.  The mux handler delegates the request to one of the backend handlers based
on the inbound content-type.

* `gke_ingress_lb/gke_ingress_lb_envoy`:  Envoy service sidecar handles all requests first.  If the inbound request is a healthcheck,
the proxy executes a LUA script that checks the upsteream health status (i.,e envoy's own healthcheck status for the upstream service).
If envoy sees the upstream service is healthy, the LUA script returns a  `200 OK`.  If the rquest is for a non healthcheck request (i.,e an actual
grpc request), it reroute to the backend listener capable of handling gRPC.  Note, the upstream healthcheck envoy runs is an implementation of
grpc Healthchecks.

Finally, note that while there utilities such as [grpc-health-probe](https://github.com/grpc-ecosystem/grpc-health-probe), but what that fulfills
is liveness and readiness checks for the container only; it does not address the HTTP healthcheck requests inbound for GCP  


- Image: 
* `salrashid123/grpc_backend`: gRPC client/server application with http2 mux
* `salrashid123/grpc_only_backend`: gRPC client/server application without http2 mux
  gRPC and http2 HC port: `:50051`

## Setup


Setup a GKE clsuter

```
gcloud container  clusters create cluster-grpc --zone us-central1-a  --num-nodes 3 --enable-ip-alias
```

setup a firewall rule to test direct access to the gRPC server via Network LB (just to show the diffence)

```
gcloud compute firewall-rules create grpc-nlb-firewall --allow tcp:50051
```

## Deploy

* `gke_ingress_lb/gke_ingress_lb_mux`:

```
kubectl apply -f fe-deployment.yaml -f fe-ingress.yaml -f fe-secret.yaml -f fe-srv-ingress.yaml -f fe-srv-lb.yaml
```

Wait maybe 10 mins for the Ingress object to give an IP and provision the LB (yes, it may take up 10mins)


or

* `gke_ingress_lb/gke_ingress_lb_envoy`:

```
kubectl apply -f envoy-configmap.yaml -f fe-secret.yaml
```

```
kubectl apply -f fe-ingress.yaml -f fe-srv-ingress.yaml -f  fe-deployment.yaml -f fe-srv-ingress.yaml -f fe-srv-lb.yaml
```

Note, the envoy config that allows for upstream custom healthchecks (`/_ah/health`) for gRPC is based on a simple LUA filter that queries the local admin instance.

```yaml
          http_filters:
          - name: envoy.lua
            config:
              inline_code: |
                package.path = "/etc/envoy/lua/?.lua;/usr/share/lua/5.1/nginx/?.lua;/etc/envoy/lua/" .. package.path

                function envoy_on_request(request_handle)
                
                  if request_handle:headers():get(":path") == "/_ah/health" then

                    local headers, body = request_handle:httpCall(
                    "local_admin",
                    {
                      [":method"] = "GET",
                      [":path"] = "/clusters",
                      [":authority"] = "local_admin"
                    },"", 50)
                    
                    request_handle:logWarn(body)                    
                    str = "local_grpc_endpoint::127.0.0.1:50051::health_flags::healthy"
                    if string.match(body, str) then
                       request_handle:respond({[":status"] = "200"},"ok")
                    else
                       request_handle:respond({[":status"] = "503"},"unavailable")
                    end

                  end

                end     
                
  clusters:

  - name: local_admin
    connect_timeout: 0.05s
    type:  STATIC
    lb_policy: ROUND_ROBIN
    hosts:
    - socket_address:
        address: 127.0.0.1
        port_value: 9000                

```

The healthcheck endpoint handled by envoy corresponds to the custom healthcheck in the Deployment (`fe-deployment.yaml`):

```yaml
        livenessProbe:
          httpGet:
            path: /_ah/health
            scheme: HTTPS
            port: fe
        readinessProbe:
          httpGet:
            path: /_ah/health
            scheme: HTTPS
            port: fe
        ports:
        - name: fe
          containerPort: 8080
          protocol: TCP
```

## Test

```
$ kubectl get no,po,rs,ing,svc
NAME                                               STATUS    ROLES     AGE       VERSION
node/gke-cluster-grpc-default-pool-aeb308a0-89dt   Ready     <none>    9m        v1.11.7-gke.12
node/gke-cluster-grpc-default-pool-aeb308a0-hv5f   Ready     <none>    9m        v1.11.7-gke.12
node/gke-cluster-grpc-default-pool-aeb308a0-vsf4   Ready     <none>    9m        v1.11.7-gke.12

NAME                                READY     STATUS    RESTARTS   AGE
pod/fe-deployment-9ff8b7c84-b6w6s   1/1       Running   0          1m
pod/fe-deployment-9ff8b7c84-hvfsm   1/1       Running   0          1m

NAME                                            DESIRED   CURRENT   READY     AGE
replicaset.extensions/fe-deployment-9ff8b7c84   2         2         2         1m

NAME                            HOSTS     ADDRESS          PORTS     AGE
ingress.extensions/fe-ingress   *         35.227.244.196   80, 443   1m

NAME                     TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)           AGE
service/fe-srv-ingress   NodePort       10.23.249.183   <none>           50051:31655/TCP   1m
service/fe-srv-lb        LoadBalancer   10.23.246.176   35.226.254.240   50051:30513/TCP   1m
service/kubernetes       ClusterIP      10.23.240.1     <none>           443/TCP           10m
```

Note the Loadbalancer and Ingress IPs assigned.  In the example above,

- LB: `35.226.254.240`
- Ingress: `35.227.244.196`


Make gRPC calls via the LB:

The response back shows the podname that handled the request.  Note all 10 requests from the client is handled by one pod.  This is imbalanced load.
```
$ docker run --add-host server.domain.com:35.226.254.240 -t salrashid123/grpc_backend /grpc_client --host server.domain.com:50051

2019/04/14 18:16:55 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:16:57 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:16:58 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:16:59 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:00 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:02 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:03 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:04 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:06 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:17:07 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
```


Now connect via the ingress L8.  Each response is from a differnt pod (blanced)

```
$ docker run --add-host server.domain.com:35.227.244.196   -t salrashid123/grpc_backend /grpc_client --host server.domain.com:443
2019/04/14 18:21:58 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:21:59 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-hvfsm"
2019/04/14 18:22:01 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:22:02 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:22:04 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:22:05 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-hvfsm"
2019/04/14 18:22:06 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-hvfsm"
2019/04/14 18:22:07 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-hvfsm"
2019/04/14 18:22:09 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
2019/04/14 18:22:10 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-9ff8b7c84-b6w6s"
```



## Internal L7 ILB

