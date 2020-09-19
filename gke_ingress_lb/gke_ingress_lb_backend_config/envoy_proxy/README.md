### Using Envoy to proxy TLS and Http healthcheck

-->> **this config is not advised**

In this envoy listens on both the http healthcheck port `:8080`, and the grpc service port `:50051`.  Envoy uses its own admin endpoint to find out if the actual backend grpc service listening on port
`:50052` is healthy not (note the actual grpc server listens on `:50052`)


essentially
podSpec(`:8080`: envoy_hc_http_server_tls;  `:50051`: envoy_proxy_tls;  `:50052`: grpc_server_no_tls )


to use

```bash
cd gcegrpc/gke_ingress_lb/gke_ingress_lb_backend_config/
kubectl apply -f fe-configmap.yaml -f fe-hc-secret.yaml -f fe-ilb-ingress.yaml -f fe-ingress.yaml -f fe-server-secret.yaml -f fe-secret.yaml -f fe-srv-ingress.yaml -f fe-srv-lb.yaml -f envoy_proxy/envoy-configmap.yaml -f  envoy_proxy/fe-deployment.yaml

```

```
$  kubectl get no,po,rs,ing,svc
NAME                                               STATUS   ROLES    AGE   VERSION
node/gke-cluster-grpc-default-pool-6af0271a-c7fb   Ready    <none>   56m   v1.18.6-gke.4801
node/gke-cluster-grpc-default-pool-6af0271a-nqs1   Ready    <none>   56m   v1.18.6-gke.4801
node/gke-cluster-grpc-default-pool-6af0271a-s0pq   Ready    <none>   56m   v1.18.6-gke.4801

NAME                                READY   STATUS    RESTARTS   AGE
pod/fe-deployment-58546ffd9-8s7x8   2/2     Running   0          6m48s
pod/fe-deployment-58546ffd9-xr49t   2/2     Running   0          6m19s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/fe-deployment-58546ffd9    2         2         2       6m48s
replicaset.apps/fe-deployment-85d898f44f   0         0         0       8m52s

NAME                                CLASS    HOSTS   ADDRESS         PORTS     AGE
ingress.extensions/fe-ilb-ingress   <none>   *                       80, 443   9m30s
ingress.extensions/fe-ingress       <none>   *       34.120.99.146   80, 443   9m30s

NAME                     TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)           AGE
service/fe-srv-ingress   NodePort       10.4.0.77    <none>          50051:32657/TCP   9m30s
service/fe-srv-lb        LoadBalancer   10.4.3.107   35.224.182.33   50051:30497/TCP   9m30s
service/kubernetes       ClusterIP      10.4.0.1     <none>          443/TCP           57m


$ docker run --add-host server.domain.com:35.224.182.33 -t salrashid123/grpc_backend /grpc_client --host server.domain.com:50051
2020/09/19 21:22:42 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:43 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:44 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:45 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:46 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:47 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:48 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:49 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:50 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:22:51 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 

$ docker run --add-host server.domain.com:34.120.99.146 -t salrashid123/grpc_backend /grpc_client --host server.domain.com:443
2020/09/19 21:23:10 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-xr49t" 
2020/09/19 21:23:11 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-xr49t" 
2020/09/19 21:23:12 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:23:13 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:23:14 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:23:15 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:23:16 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-xr49t" 
2020/09/19 21:23:17 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-xr49t" 
2020/09/19 21:23:18 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8" 
2020/09/19 21:23:19 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-58546ffd9-8s7x8"
```