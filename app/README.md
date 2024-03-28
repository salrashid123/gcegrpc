# gRPC client/server test app


* `salrashid123/grpc_backend`: gRPC client/server application with http2 mux
* `salrashid123/grpc_only_backend`: gRPC client/server application without http2 mux
* `salrashid123/http_frontend`: http frontend webapp use by the GKE Interal LB sample (`gke_svc_lb`)


```bash
go run src/grpc_server.go --grpcport 0.0.0.0:50051
```

RUN Client

```bash
go run src/grpc_client.go --host localhost:50051 --servername grpc.domain.com --cacert CA_crt.pem
```


### Using Envoy for TLS Proxy

You can use envoy to proxy TLS connections for a gRPC backend service.

For example, you can run the sample client/server here without TLS:

```bash
cd grpc_only_backend/
go run src/grpc_server.go --grpcport :50051 --insecure
go run src/grpc_client.go --host localhost:50051 --insecure
```

Or with TLS

```bash
go run src/grpc_server.go --grpcport :50051 --tlsCert server_crt.pem --tlsKey server_key.pem
go run src/grpc_client.go --host localhost:50051 --cacert CA_crt.pem
```

But if you need to run envoy in the middle to handle the TLS:

```bash
 go run src/grpc_server.go --grpcport :50051 --insecure

envoy -c envoy_config.yaml 

go run src/grpc_client.go --host localhost:8081 --cacert CA_crt.pem
```

you can wrap envoy config provided under `grpc_only_backend/envoy_config.yaml` using a configmap as shown [here](https://github.com/salrashid123/gcegrpc/blob/master/gke_ingress_lb/gke_ingress_lb_envoy/fe-deployment.yaml#L19)


The default `server_crt.pem` here contains a variety of SAN definitions:

```bash
openssl x509 -in server_crt.pem -noout -text

X509v3 extensions:
            Netscape Comment: 
                OpenSSL Generated Certificate
            X509v3 Subject Alternative Name: 
                DNS:server.domain.com, DNS:grpc.domain.com, DNS:be-srv, DNS:be-srv.default.svc.cluster.local, DNS:be-srv-lb, DNS:be-srv-lb.default.svc.cluster.local, DNS:grpc.domain.com, DNS:grpcweb.domain.com, IP Address:127.0.0.1

```

If you need to define your own CA, you can use openssl or follow the snippet provided in the following repo: [CA Scratchpad](https://github.com/salrashid123/ca_scratchpad)



