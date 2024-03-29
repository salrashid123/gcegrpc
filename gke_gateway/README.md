# GKE gRPC Gateway LoadBalancing


>> Copied from [gRPC on Gateway Controller](https://github.com/GoogleCloudPlatform/gke-networking-recipes/tree/main/gateway/grpc)  (besides, i wrote that example anyway :)


Sample showing gRPC clients connecting via GKE Gateway.

* Deploy gRPC application on GKE
* Enable Gateways to handle both internet facing and internal-only traffic.
* Verify gRPC LoadBalancing through Gateway

```bash
gcloud container  clusters create cluster-grpc \
   --zone us-central1-a --gateway-api=standard --num-nodes 3 --enable-ip-alias

cd gke_gateway
```

optionally create SSL Certificate for use with statically defined certificates (`networking.gke.io/pre-shared-certs`)

```bash
gcloud compute ssl-certificates create gcp-cert-grpc-global  --global --certificate server.crt --private-key server.key 

gcloud compute ssl-certificates create gcp-cert-grpc-us-central   --region=us-central1 --certificate server.crt --private-key server.key 
```

or use the default `spec.listeners.tls.certificateRef`.   For reference see [GatewayClass capabilities](https://cloud.google.com/kubernetes-engine/docs/how-to/gatewayclass-capabilities#gateway)

Wait maybe 10 mins for the Gateway controllers to get initialized.

Deploy application

```bash
kubectl apply -f .
```

> Please note the deployments here use the health_check proxy and sample gRPC applications hosted on `docker.io/`.  You can build and deploy these images into your own repository as well.

Wait another 8mins for the IP address for the loadbalancers to get initialized

Check gateway status

```bash
$ kubectl get gatewayclass,gateway
NAME                                                                      CONTROLLER                  ACCEPTED   AGE
gatewayclass.gateway.networking.k8s.io/gke-l7-global-external-managed     networking.gke.io/gateway   True       72m
gatewayclass.gateway.networking.k8s.io/gke-l7-gxlb                        networking.gke.io/gateway   True       72m
gatewayclass.gateway.networking.k8s.io/gke-l7-regional-external-managed   networking.gke.io/gateway   True       72m
gatewayclass.gateway.networking.k8s.io/gke-l7-rilb                        networking.gke.io/gateway   True       72m

NAME                                               CLASS                            ADDRESS          PROGRAMMED   AGE
gateway.gateway.networking.k8s.io/gke-l7-gxlb-gw   gke-l7-global-external-managed   34.102.243.138   True         2m59s
gateway.gateway.networking.k8s.io/gke-l7-rilb-gw   gke-l7-rilb                      10.128.0.28      True         2m59s
```


Get Gateway IPs

```bash
export GW_XLB_VIP=$(kubectl get gateway gke-l7-gxlb-gw -o json | jq '.status.addresses[].value' -r)
echo $GW_XLB_VIP

export GW_ILB_VIP=$(kubectl get gateway gke-l7-rilb-gw -o json | jq '.status.addresses[].value' -r)
echo $GW_ILB_VIP
```

#### Test External

Verify external loadbalancing by transmitting 10 RPCs over one channel.  The responses will show different pods that handled each request

```bash
docker run --add-host grpc.domain.com:$GW_XLB_VIP \
  -t salrashid123/grpc_backend /grpc_client \
  --host grpc.domain.com:443


2024/03/29 12:53:33 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:34 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:35 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:53:36 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:37 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:53:38 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:53:39 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:40 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:41 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:53:42 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
```



#### Test Internal

To test the internal loadbalancer, you must configure a VM from within an [allocated network](https://cloud.google.com/load-balancing/docs/l7-internal/setting-up-l7-internal#configuring_the_proxy-only_subnet) and export the environment variable `$GW_ILB_VIP` locally.  You can either install docker on that VM or Go.  Once that is done, invoke the Gateway using the ILB address:

```bash
docker run --add-host grpc.domain.com:$GW_ILB_VIP \
  -t salrashid123/grpc_backend /grpc_client \
  --host grpc.domain.com:443

2024/03/29 12:52:10 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:52:11 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:52:12 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:52:13 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:52:14 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:52:15 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:52:16 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:52:17 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
2024/03/29 12:52:18 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-wv7jn"
2024/03/29 12:52:19 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-6478dd7c9-r7jw2"
```

---

Source images used in this example can be found here:
  - [docker.io/salrashid123/grpc_health_proxy](https://github.com/salrashid123/grpc_health_proxy)
  - [docker.io/salrashid123/grpc_app](https://github.com/salrashid123/grpc_health_proxy/tree/master/example)
