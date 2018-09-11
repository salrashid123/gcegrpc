# gRPC client/server test app


# Setup

```
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

go get  github.com/golang/protobuf/proto \
        golang.org/x/net/context \
        google.golang.org/grpc \
        google.golang.org/grpc/credentials \
        google.golang.org/grpc/health \
        google.golang.org/grpc/health/grpc_health_v1 \
        google.golang.org/grpc/metadata \
        golang.org/x/net/trace \
        golang.org/x/net/http2 \
        golang.org/x/net/http2/hpack

go get -u github.com/golang/protobuf/protoc-gen-go
```


Compile .proto

```
protoc --go_out=plugins=grpc:. src/echo/echo.proto
```



Edit /etc/hosts
  Since this is just a demo/POC, statically set the IP to resolve to .domain.com as shown below

```
/etc/hosts
35.241.41.138 server.domain.com grpcweb.domain.com grpc.domain.com
```

RUN Server

```
go run src/grpc_server.go -grpcport 0.0.0.0:50051
```

RUN Cient

```
go run src/grpc_client.go --host grpc.domain.com:50051
```