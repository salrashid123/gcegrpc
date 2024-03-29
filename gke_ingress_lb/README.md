# GKE gRPC Ingress LoadBalancing

Sample showing gRPC clients connecting via Ingress.

In this mode one gRPC connection sends 10 rpc messages.  Ingress L7 intercepts the ssl connection and then transmits each RPC back to different pods.
Since each RPC goes to different endpoints, the load is more evenly distributed between all pods.



## Setup

```bash
$ gcloud container  clusters create cluster-grpc \
   --zone us-central1-a  --num-nodes 3 --enable-ip-alias 
```


If using ILB, see [Setting up ILB Subnet](https://cloud.google.com/load-balancing/docs/l7-internal/setting-up-l7-internal#configuring_the_proxy-only_subnet)

first create an ILB subnet in the appropriate range (in this case, its `192.168.0.0/23`)

```bash
$ gcloud compute firewall-rules create allow-grpc-inbound-50051  --action allow --direction INGRESS    --source-ranges 0.0.0.0/0     --rules tcp:50051

$ gcloud compute networks subnets create proxy-only-subnet \
  --purpose=REGIONAL_MANAGED_PROXY     --role=ACTIVE \
    --region=us-central1     --network=default     --range=192.168.0.0/23   
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
$ docker run --add-host grpc.domain.com:35.224.74.161 \
  -t salrashid123/grpc_backend /grpc_client \
  --host grpc.domain.com:50051

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
$ docker run --add-host grpc.domain.com:34.120.99.146 \
  -t salrashid123/grpc_backend /grpc_client \
  --host grpc.domain.com:443

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
$ docker run --add-host grpc.domain.com:10.128.0.95 \
   -t salrashid123/grpc_backend /grpc_client \
   --host grpc.domain.com:443

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
