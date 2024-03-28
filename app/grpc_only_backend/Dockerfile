
FROM golang:1.20 as build

RUN apt-get update -y && apt-get install -y build-essential wget unzip curl


RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip && \
    unzip protoc-25.1-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/* /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/


WORKDIR /go/src/app

COPY . .
RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --go-grpc_opt=paths=source_relative src/echo/echo.proto


#RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
#    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#    chmod +x /bin/grpc_health_probe

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/grpc_server src/grpc_server.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/grpc_client src/grpc_client.go

FROM gcr.io/distroless/base
COPY --from=build /go/src/app/server_crt.pem /
COPY --from=build /go/src/app/server_key.pem /
COPY --from=build /go/src/app/CA_crt.pem /
COPY --from=build /go/bin/grpc_server /
COPY --from=build /go/bin/grpc_client /

EXPOSE 50051

#ENTRYPOINT ["grpc_server", "--grpcport", ":50051"]
#ENTRYPOINT ["grpc_client", "--host",  "server.domain.com:50051"]
