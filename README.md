
# gRPC on GKE and Istio


Samples for running gRPC on GKE and Istio:


## gRPC Loadbalancing on GKE with L7 Ingress

`client_grpc_app  (via gRPC wire protocol) --> ingress --> (grpc Service on GKE)`

- Folder: [gke_ingress_lb/](gke_ingress_lb/)

## gRPC Loadbalancing on GKE Gateway

`client_grpc_app  (via gRPC wire protocol) --> Gateway --> (grpc Service on GKE)`

- Folder: [gke_gateway/](gke_gateway/)

## gRPC Loadbalancing on GKE with TrafficDirector (xds loadbalancing)

`client_grpc_app  (acquire gRPC lookaside loadbalancing/XDS) --> GKE pod NEG address`

- Folder: [gke_td_xds/](gke_td_xds/)

## gRPC-web via Ingress

 Javascript clients:  
     
`client(browser) (via grpc-web protocol) --> Ingress --> (grpc Service on GKE)`

 - [grpc_web_with_gke](https://github.com/salrashid123/grpc_web_with_gke)

## gRPC for GKE internal Service->Service

- Folder [gke_svc_lb/](gke_svc_lb/)
      
`client (pod running in GKE) --> k8s Service --> (gRPC endpoint in same GKE cluster)`


## gRPC w/ Managed Instance Group with Container Optimized OS

- Folder [gce](gce/)

`client_grpc_app --> L7LB --> ManagedInstanceGroup`


## Istio gRPC Loadbalancing with GCP Internal LoadBalancer (ILB)

- folder `istio/`

`client_grpc_app (on GCEVM) --> (GCP ILB) --> Istio --> Service`

`client_grpc_app (external) --> (GCP ExternalLB) --> Istio --> Service`

Also see:
- [Summary of Cloud load balancers](https://cloud.google.com/load-balancing/docs/choosing-load-balancer#summary_of_cloud_load_balancers)
- [GCP LoadBalancing Overview](https://cloud.google.com/load-balancing/docs/load-balancing-overview#internal_tcpudp_load_balancing)


---

## Source and Dockerfile

You can find the source here:

- Client and Server: [app/grpc_backend/](app/grpc_backend)

And the various docker images on dockerhub


- `docker.io/salrashid123/grpc_backend`


To run the gRPC server locally to see message replay:

- Server:
    `docker run  -p 50051:50051 -t salrashid123/grpc_backend ./grpc_server -grpcport 0.0.0.0:50051`

- Client:
    `docker run --net=host --add-host grpc.domain.com:127.0.0.1 -t salrashid123/grpc_backend /grpc_client --host grpc.domain.com:50051`  

What client app makes ONE connection out to the Server and then proceeds to send 10 RPC calls over that conn.  For each call, the server
will echo back the hostname of the server/pod that handled that request.  In the example here locally, it will be all from one host.

## References

 - [https://grpc.io/blog/loadbalancing](https://grpc.io/blog/loadbalancing)
 - [gRPC Loadbalancing on Kubernetes (Kubecon Europe 2018)](https://www.youtube.com/watch?v=F2znfxn_5Hg)
 - [gRPC-web "helloworld"](https://github.com/salrashid123/gcegrpc/tree/master/grpc-web)
 - [gRPC with curl](https://github.com/salrashid123/grpc_curl)
