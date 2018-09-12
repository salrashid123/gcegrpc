
# gRPC --> GCE L7 lb --> Container Optimized Managed Instance groups


Sample gRPC server set deployed as an autoscaled Instance Group running [Container Optimized OS](https://cloud.google.com/container-optimized-os/docs/how-to/run-container-instance).

Basically, this runs the gRPC server directly on a VM as a container.  Inbound requests from outside GCP to it are loadblanced via L7 LB (which also terminates SSL)

Ref

- [https://cloud.google.com/load-balancing/docs/https/cross-region-example](https://cloud.google.com/load-balancing/docs/https/cross-region-example)


## Setup:

- First get an external IP address

```
gcloud compute addresses create gke-ingress --global
gcloud compute addresses describe  gke-ingress --global --format="value(address)"
```

- Use the certificates provided to setup LB SSL

```
gcloud compute ssl-certificates create grpcert --certificate=server_crt.pem --private-key=server_key.pem
```

- Configure the instance template for the Managed Instance Group

```
gcloud beta compute instance-templates create-with-container grpctemplate  --machine-type g1-small --tags grpcserver      --container-image="salrashid123/grpc_backend" --container-command="./grpc_server" --container-arg="--grpcport=0.0.0.0:50051"      --container-arg="--httpport=0.0.0.0:8081"
```

- Setup an Initial size for the MIG

```
gcloud compute instance-groups managed create web-group-a  --base-instance-name webpool  --zone us-central1-a   --size 2   --template grpctemplate
```

- Setup some firewall rules

```
gcloud  compute  firewall-rules create firewall-rules-http2 --allow=tcp:8081 --source-ranges=130.211.0.0/22,35.191.0.0/16  --target-tags=grpcserver
gcloud  compute  firewall-rules create firewall-rules-grpc --allow=tcp:50051 --source-ranges=130.211.0.0/22,35.191.0.0/16  --target-tags=grpcserver
```

Also suggest setting up direct access to the VMs (just for testing, you can remove this)

```
gcloud  compute  firewall-rules create firewall-rules-allow-test-grpc --allow=tcp:50051 --source-ranges=0.0.0.0/0  --target-tags=grpcserver
```


- Setup the MIG, HealthChecks, BackendServices

```
gcloud compute instance-groups managed set-named-ports web-group-a --named-ports=grpc-port:50051 --zone us-central1-a

gcloud beta compute health-checks create  http2  webpool-basic-check --port=8081 --request-path="/_ah/health"

gcloud beta compute backend-services create webpool-map-backend-service --port-name=grpc-port  --protocol=http2 --health-checks=webpool-basic-check --global

gcloud beta compute backend-services add-backend webpool-map-backend-service  --instance-group web-group-a --instance-group-zone us-central1-a --global

gcloud compute url-maps create webpool-map --default-service webpool-map-backend-service

gcloud alpha compute target-https-proxies create https-lb-proxy  --url-map=webpool-map  --ssl-certificates=grpcert --global
```

- Final steps

```
gcloud compute forwarding-rules create https-content-rule --address `gcloud compute addresses describe  gke-ingress --global --format="value(address)"`  --global --target-https-proxy https-lb-proxy --ports 443
```

>> now wait upto 10mins for the LB to provisoin



## Test Direct

Test the gRPC server by directly accessing each VM with its public IP

```
gcloud compute instances list --filter="tags.items=grpcserver"
docker run --add-host server.domain.com:<VM_PUBLIC_IP>  -t salrashid123/grpc_backend /grpc_client --host server.domain.com:50051
```


You should see respones direct from each VM

eg:
```
$ gcloud compute instances list --filter="tags.items=grpcserver"
NAME          ZONE           MACHINE_TYPE  PREEMPTIBLE  INTERNAL_IP  EXTERNAL_IP     STATUS
webpool-c57z  us-central1-a  g1-small                   10.128.0.3   35.184.249.181  RUNNING
webpool-dw5f  us-central1-a  g1-small                   10.128.0.2   35.232.128.12   RUNNING

$ docker run --add-host server.domain.com:35.184.249.181  -t salrashid123/grpc_backend /grpc_client --host server.domain.com:50051
2018/09/10 23:58:26 RPC Response: 0 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:27 RPC Response: 1 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:28 RPC Response: 2 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:29 RPC Response: 3 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:30 RPC Response: 4 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:31 RPC Response: 5 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:32 RPC Response: 6 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:33 RPC Response: 7 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:34 RPC Response: 8 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:58:35 RPC Response: 9 message:"Hello unary RPC msg   from hostname webpool-c57z" 

$ docker run --add-host server.domain.com:35.232.128.12  -t salrashid123/grpc_backend /grpc_client --host server.domain.com:50051
2018/09/10 23:58:52 RPC Response: 0 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:53 RPC Response: 1 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:54 RPC Response: 2 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:55 RPC Response: 3 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:57 RPC Response: 4 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:58 RPC Response: 5 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:58:59 RPC Response: 6 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:59:00 RPC Response: 7 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:59:01 RPC Response: 8 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:59:02 RPC Response: 9 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
```


## Test via L7 (after ~10mins L7 programming)

Test the gRPC server via LB:

```
export LB_IP=`gcloud compute addresses describe  gke-ingress --global --format="value(address)"`
docker run --add-host server.domain.com:$LB_IP  -t salrashid123/grpc_backend /grpc_client --host server.domain.com:443
```


You should see responses loadbalanced across nodes:

eg:
```
$ docker run --add-host server.domain.com:35.241.41.138  -t salrashid123/grpc_backend /grpc_client --host server.domain.com:443
2018/09/10 23:57:36 RPC Response: 0 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:57:37 RPC Response: 1 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:57:38 RPC Response: 2 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:57:39 RPC Response: 3 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:57:41 RPC Response: 4 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:57:42 RPC Response: 5 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:57:43 RPC Response: 6 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:57:44 RPC Response: 7 message:"Hello unary RPC msg   from hostname webpool-c57z" 
2018/09/10 23:57:45 RPC Response: 8 message:"Hello unary RPC msg   from hostname webpool-dw5f" 
2018/09/10 23:57:46 RPC Response: 9 message:"Hello unary RPC msg   from hostname webpool-c57z" 
```

## Delete

Runt he following to delete the setup

```
gcloud compute forwarding-rules delete  https-content-rule  -q --global
gcloud compute target-https-proxies delete https-lb-proxy -q
gcloud compute url-maps delete  webpool-map -q
gcloud compute backend-services  remove-backend webpool-map-backend-service --instance-group=web-group-a --global --instance-group-zone us-central1-a -q
gcloud compute backend-services delete webpool-map-backend-service --global -q
gcloud beta compute health-checks delete  webpool-basic-check -q
gcloud compute  firewall-rules delete firewall-rules-http2 -q
gcloud compute  firewall-rules delete firewall-rules-grpc -q
gcloud compute  firewall-rules delete firewall-rules-allow-test-grpc -q
gcloud compute instance-groups managed delete web-group-a --zone us-central1-a -q
gcloud compute instance-templates delete grpctemplate -q
```



