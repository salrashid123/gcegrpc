# gRPC client/server test app


* `salrashid123/grpc_backend`: gRPC client/server application with http2 mux
* `salrashid123/grpc_only_backend`: gRPC client/server application without http2 mux
* `salrashid123/http_frontnd`: http frontend webapp use by the GKE Interal LB sample (`gke_svc_lb`)

# Setup

```apt-get update -y && apt-get install -y build-essential wget unzip curl
```

get protoc

```
 curl -OL https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-linux-x86_64.zip && \
    unzip protoc-3.2.0-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/* /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/
```


Build app

```
cd grpc_backend

export GOPATH=`pwd`

go get golang.org/x/net/context \
        golang.org/x/net/http2 \
        google.golang.org/grpc \
        google.golang.org/grpc/credentials \
        google.golang.org/grpc/health \
        google.golang.org/grpc/health/grpc_health_v1 \
        google.golang.org/grpc/metadata

go get -u github.com/golang/protobuf/protoc-gen-go
```


optionally compile .proto

```
protoc --go_out=plugins=grpc:. src/echo/echo.proto
```


or just RUN Server

```
go run src/grpc_server.go --grpcport 0.0.0.0:50051
```

RUN Cient

```
go run src/grpc_client.go --host localhost:50051 --servername grpc.domain.com --cacert CA_crt.pem
```